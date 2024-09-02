package handlers

import (
	"bytes"
	"compress/gzip"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/middlewares"
	"github.com/romanp1989/go-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	type want struct {
		statusCode  int
		responseURL string
	}

	type ctxAuthKey string
	const authKey ctxAuthKey = "auth"

	tests := []struct {
		name        string
		method      string
		requestBody string
		want        want
	}{
		{
			name:        "Valid_URL",
			method:      http.MethodPost,
			requestBody: "https://ya.ru",
			want: want{
				statusCode:  http.StatusCreated,
				responseURL: "http://localhost:8080/6YGS4ZUF",
			},
		},
		{
			name:        "Empty_URL",
			method:      http.MethodPost,
			requestBody: "",
			want: want{
				statusCode:  http.StatusBadRequest,
				responseURL: "",
			},
		},
	}
	config.ParseFlags()
	s := storage.Init(config.Options.FlagDatabaseDsn, config.Options.FlagFileStorage)
	h := New(s)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.requestBody)
			r := httptest.NewRequest(tt.method, "/", body)
			r.Header.Set("Content-Type", "text/plain")

			userID := auth.EnsureRandom()
			rctx := context.WithValue(r.Context(), authKey, userID)
			r = r.WithContext(rctx)

			w := httptest.NewRecorder()

			fn := h.Encode()
			fn(w, r)

			result := w.Result()

			resBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()

			require.NoError(t, err)

			assert.Equal(t, tt.want.responseURL, string(resBody), "Ожидаемый URL %s не совпадает с фактическим %s", tt.want.responseURL, string(resBody))

			assert.Equal(t, tt.want.statusCode, result.StatusCode, "Ожидаемый код ответа %s не совпадаем с фактических %s", tt.want.statusCode, result.StatusCode)

		})
	}
}

func TestDecode(t *testing.T) {
	type want struct {
		statusCode  int
		responseURL string
	}

	tests := []struct {
		name string
		want want
	}{
		{
			name: "Success_redirect",
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				responseURL: "https://ya.ru",
			},
		},
		{
			name: "Fail_redirect",
			want: want{
				statusCode:  http.StatusNotFound,
				responseURL: "",
			},
		},
	}

	s := storage.Init("", "")
	h := New(s)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashID := shortURL(tt.want.responseURL)
			userID := auth.EnsureRandom()
			if tt.want.responseURL != "" {
				s.SaveURL(context.Background(), tt.want.responseURL, hashID, &userID)
			}
			body := httptest.NewRequest(http.MethodGet, "/{id}", nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", hashID)
			r := body.WithContext(context.WithValue(body.Context(), chi.RouteCtxKey, rctx))
			fn := h.Decode()
			fn(w, r)

			result := w.Result()

			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode, "Ожидаемый код ответа %s не совпадаем с фактических %s", tt.want.statusCode, result.StatusCode)

			if tt.want.responseURL != "" {
				url := w.Header().Get("Location")
				assert.Equal(t, tt.want.responseURL, url, "Ожидаемый URL %s не совпадает с фактическим %s", tt.want.responseURL, url)
			}

		})
	}
}

func TestShorten(t *testing.T) {
	type want struct {
		statusCode  int
		responseURL string
	}

	type ctxAuthKey string
	const authKey ctxAuthKey = "auth"

	var tests = []struct {
		name        string
		method      string
		requestBody string
		want        want
	}{
		{
			name:        "Valid_URL",
			method:      http.MethodPost,
			requestBody: `{"url": "https://ya.ru"}`,
			want: want{
				statusCode:  http.StatusCreated,
				responseURL: `{"result":"http://localhost:8080/6YGS4ZUF"}`,
			},
		},
	}

	config.ParseFlags()
	s := storage.Init(config.Options.FlagDatabaseDsn, config.Options.FlagFileStorage)
	h := New(s)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.requestBody)
			r := httptest.NewRequest(tt.method, "/", body)

			userID := auth.EnsureRandom()
			rctx := context.WithValue(r.Context(), authKey, userID)
			r = r.WithContext(rctx)
			r.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			fn := h.Shorten()
			fn(w, r)

			result := w.Result()

			resBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()

			require.NoError(t, err)

			assert.JSONEq(t, tt.want.responseURL, w.Body.String(), "Ожидаемый URL %s не совпадает с фактическим %s", tt.want.responseURL, string(resBody))

			assert.Equal(t, tt.want.statusCode, w.Code, "Ожидаемый код ответа %s не совпадаем с фактических %s", tt.want.statusCode, result.StatusCode)

		})
	}
}

func TestGzipCompression(t *testing.T) {
	config.ParseFlags()
	s := storage.Init(config.Options.FlagDatabaseDsn, config.Options.FlagFileStorage)
	h := New(s)

	handler := middlewares.AuthMiddlewareSet(middlewares.GzipMiddleware(h.Encode()))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	requestBody := `https://ya.ru`

	// ожидаемое содержимое тела ответа при успешном запросе
	successBody := `http://localhost:8080/6YGS4ZUF`

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "")

		userID := auth.EnsureRandom()
		token, _ := auth.CreateToken(&userID)

		cookie := &http.Cookie{
			Name:  "auth",
			Value: token,
			Path:  "/",
		}
		r.AddCookie(cookie)

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		userID := auth.EnsureRandom()
		token, _ := auth.CreateToken(&userID)

		cookie := &http.Cookie{
			Name:  "auth",
			Value: token,
			Path:  "/",
		}
		r.AddCookie(cookie)

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.Equal(t, successBody, string(b))
	})
}
