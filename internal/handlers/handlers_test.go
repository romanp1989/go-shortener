package handlers

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/config"
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
		{
			name:        "Wrong_request_type",
			method:      http.MethodPut,
			requestBody: "",
			want: want{
				statusCode:  http.StatusBadRequest,
				responseURL: "Некорректный тип запроса\n",
			},
		},
	}

	config.ParseFlags()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.requestBody)
			r := httptest.NewRequest(tt.method, "/", body)
			r.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()

			fn := Encode()
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashID := shortURL(tt.want.responseURL)
			if tt.want.responseURL != "" {
				urlStore[hashID] = tt.want.responseURL
			}
			body := httptest.NewRequest(http.MethodGet, "/{id}", nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", hashID)
			r := body.WithContext(context.WithValue(body.Context(), chi.RouteCtxKey, rctx))
			fn := Decode()
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
