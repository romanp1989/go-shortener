package handlers

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/middlewares"
	"github.com/romanp1989/go-shortener/internal/models"
	"github.com/romanp1989/go-shortener/internal/models/mocks"
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

	tests := []struct {
		name        string
		method      string
		userID      uuid.UUID
		requestBody string
		want        want
	}{
		{
			name:        "Valid_URL",
			method:      http.MethodPost,
			userID:      auth.EnsureRandom(),
			requestBody: "https://ya.ru",
			want: want{
				statusCode:  http.StatusCreated,
				responseURL: "http://localhost:8080/6YGS4ZUF",
			},
		},
		{
			name:        "Empty_URL",
			method:      http.MethodPost,
			userID:      auth.EnsureRandom(),
			requestBody: "",
			want: want{
				statusCode:  http.StatusBadRequest,
				responseURL: "",
			},
		},
		{
			name:        "User_Unauthorized",
			method:      http.MethodPost,
			userID:      uuid.UUID{},
			requestBody: "https://ya.ru",
			want: want{
				statusCode:  http.StatusUnauthorized,
				responseURL: "",
			},
		},
		{
			name:        "Bad_URL",
			method:      http.MethodPost,
			userID:      auth.EnsureRandom(),
			requestBody: "https://ya . ru",
			want: want{
				statusCode:  http.StatusBadRequest,
				responseURL: "",
			},
		},
		{
			name:        "Short_URL_Save_Conflict_Error",
			method:      http.MethodPost,
			userID:      auth.EnsureRandom(),
			requestBody: "https://ya.ru",
			want: want{
				statusCode:  http.StatusConflict,
				responseURL: "",
			},
		},
		{
			name:        "Short_URL_Save_Other_Error",
			method:      http.MethodPost,
			userID:      auth.EnsureRandom(),
			requestBody: "https://ya.ru",
			want: want{
				statusCode:  http.StatusBadRequest,
				responseURL: "",
			},
		},
	}
	config.ParseFlags()

	mockCtrl := gomock.NewController(t)
	mockStorageDB := mocks.NewMockStorage(mockCtrl)
	defer mockCtrl.Finish()

	storageURLs := storage.Storage{Storage: mockStorageDB}
	handler := New(storageURLs)

	firstCall := mockStorageDB.EXPECT().Save(gomock.Any(), "https://ya.ru", "6YGS4ZUF", gomock.Any()).Return("6YGS4ZUF", nil)
	secondCall := mockStorageDB.EXPECT().Save(gomock.Any(), "https://ya.ru", "6YGS4ZUF", gomock.Any()).After(firstCall).Return("", storage.NewURLConflictError("6YGS4ZUF", storage.ErrConflict))
	mockStorageDB.EXPECT().Save(gomock.Any(), "https://ya.ru", "6YGS4ZUF", gomock.Any()).After(secondCall).Return("", errors.New("Ошибка вставки URL в БД"))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.requestBody)
			r := httptest.NewRequest(tt.method, "/", body)
			r.Header.Set("Content-Type", "text/plain")

			rctx := context.WithValue(r.Context(), auth.AuthKey, tt.userID)
			r = r.WithContext(rctx)

			w := httptest.NewRecorder()

			fn := handler.Encode()
			fn(w, r)

			result := w.Result()

			resBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()

			require.NoError(t, err)

			if tt.want.responseURL != "" {
				assert.Equal(t, tt.want.responseURL, string(resBody), "Ожидаемый URL %s не совпадает с фактическим %s", tt.want.responseURL, string(resBody))
			}

			assert.Equal(t, tt.want.statusCode, result.StatusCode, "Ожидаемый код ответа %s не совпадаем с фактических %s", tt.want.statusCode, result.StatusCode)

		})
	}
}

func TestDecode(t *testing.T) {
	type want struct {
		statusCode  int
		responseURL string
	}

	firstOriginalURL := "https://ya.ru"
	firstShort := ShortURL("https://ya.ru")

	secondShort := ShortURL("https://yandex.ru")
	thirdShort := ShortURL("https://dzen.ru")
	fourthShort := ShortURL("https://mail.ru")

	tests := []struct {
		name     string
		userID   uuid.UUID
		shortURL string
		want     want
	}{
		{
			name:     "Success_redirect",
			userID:   auth.EnsureRandom(),
			shortURL: firstShort,
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				responseURL: firstOriginalURL,
			},
		},
		{
			name:     "Fail_redirect",
			userID:   auth.EnsureRandom(),
			shortURL: secondShort,
			want: want{
				statusCode:  http.StatusNotFound,
				responseURL: "",
			},
		},
		{
			name:     "Empty_URL",
			userID:   auth.EnsureRandom(),
			shortURL: "",
			want: want{
				statusCode:  http.StatusBadRequest,
				responseURL: "",
			},
		},
		{
			name:     "Already_Deleted_URL",
			userID:   auth.EnsureRandom(),
			shortURL: thirdShort,
			want: want{
				statusCode:  http.StatusGone,
				responseURL: "",
			},
		},
		{
			name:     "Error_Get_URL",
			userID:   auth.EnsureRandom(),
			shortURL: fourthShort,
			want: want{
				statusCode:  http.StatusBadRequest,
				responseURL: "",
			},
		},
	}

	mockCtrl := gomock.NewController(t)
	mockStorageDB := mocks.NewMockStorage(mockCtrl)
	defer mockCtrl.Finish()

	storageURLs := storage.Storage{Storage: mockStorageDB}
	handler := New(storageURLs)

	mockStorageDB.EXPECT().Get(firstShort).Return(firstOriginalURL, nil).Times(1)
	mockStorageDB.EXPECT().Get(secondShort).Return("", nil).Times(1)
	mockStorageDB.EXPECT().Get(thirdShort).Return("", storage.NewAlreadyDeletedError(thirdShort)).Times(1)
	mockStorageDB.EXPECT().Get(fourthShort).Return("", errors.New("error get url response")).Times(1)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			body := httptest.NewRequest(http.MethodGet, "/{id}", nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.shortURL)
			r := body.WithContext(context.WithValue(body.Context(), chi.RouteCtxKey, rctx))
			fn := handler.Decode()
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

	var tests = []struct {
		name        string
		method      string
		userID      uuid.UUID
		requestBody string
		want        want
	}{
		{
			name:        "Valid_URL",
			method:      http.MethodPost,
			userID:      auth.EnsureRandom(),
			requestBody: `{"url": "https://ya.ru"}`,
			want: want{
				statusCode:  http.StatusCreated,
				responseURL: `{"result":"http://localhost:8080/6YGS4ZUF"}`,
			},
		},
		{
			name:        "User_Unauthorized",
			method:      http.MethodPost,
			userID:      uuid.UUID{},
			requestBody: `{"url": "https://ya.ru"}`,
			want: want{
				statusCode:  http.StatusUnauthorized,
				responseURL: "",
			},
		},
		{
			name:        "Bad_Request",
			method:      http.MethodPost,
			userID:      auth.EnsureRandom(),
			requestBody: `343434{"url": "https://ya.ru"}`,
			want: want{
				statusCode:  http.StatusInternalServerError,
				responseURL: "",
			},
		},
		{
			name:        "Short_URL_Save_Conflict_Error",
			method:      http.MethodPost,
			userID:      auth.EnsureRandom(),
			requestBody: `{"url": "https://ya.ru"}`,
			want: want{
				statusCode:  http.StatusConflict,
				responseURL: "",
			},
		},
		{
			name:        "Short_URL_Save_Other_Error",
			method:      http.MethodPost,
			userID:      auth.EnsureRandom(),
			requestBody: `{"url": "https://ya.ru"}`,
			want: want{
				statusCode:  http.StatusBadRequest,
				responseURL: "",
			},
		},
	}

	err := config.ParseFlags()
	if err != nil {
		return
	}

	mockCtrl := gomock.NewController(t)
	mockStorageDB := mocks.NewMockStorage(mockCtrl)
	defer mockCtrl.Finish()

	storageURLs := storage.Storage{Storage: mockStorageDB}
	handler := New(storageURLs)

	firstCall := mockStorageDB.EXPECT().Save(gomock.Any(), "https://ya.ru", "6YGS4ZUF", gomock.Any()).Return("6YGS4ZUF", nil)
	secondCall := mockStorageDB.EXPECT().Save(gomock.Any(), "https://ya.ru", "6YGS4ZUF", gomock.Any()).After(firstCall).Return("", storage.NewURLConflictError("6YGS4ZUF", storage.ErrConflict))
	mockStorageDB.EXPECT().Save(gomock.Any(), "https://ya.ru", "6YGS4ZUF", gomock.Any()).After(secondCall).Return("", errors.New("Ошибка вставки URL в БД"))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.requestBody)
			r := httptest.NewRequest(tt.method, "/", body)

			rctx := context.WithValue(r.Context(), auth.AuthKey, tt.userID)
			r = r.WithContext(rctx)
			r.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			fn := handler.Shorten()
			fn(w, r)

			result := w.Result()

			resBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()

			require.NoError(t, err)

			if tt.want.responseURL != "" {
				assert.JSONEq(t, tt.want.responseURL, w.Body.String(), "Ожидаемый URL %s не совпадает с фактическим %s", tt.want.responseURL, string(resBody))
			}

			assert.Equal(t, tt.want.statusCode, w.Code, "Ожидаемый код ответа %s не совпадаем с фактических %s", tt.want.statusCode, result.StatusCode)

		})
	}
}

func TestHandlers_SaveBatch(t *testing.T) {
	type want struct {
		statusCode  int
		responseURL string
	}

	var tests = []struct {
		name        string
		method      string
		userID      uuid.UUID
		requestBody string
		want        want
	}{
		{
			name:   "Valid_URL",
			method: http.MethodPost,
			userID: auth.EnsureRandom(),
			requestBody: `[
				{
					"correlation_id": "ssdfdsfsfsd",
					"original_url": "https://ya.ru"
				},
				{
					"correlation_id": "rtyuiookjhtr",
					"original_url": "https://dzen.ru"
				}
			]`,
			want: want{
				statusCode: http.StatusCreated,
				responseURL: `[
					{
						"correlation_id": "ssdfdsfsfsd",
						"short_url": "http://localhost:8080/6YGS4ZUF"
					},
					{
						"correlation_id": "rtyuiookjhtr",
						"short_url": "http://localhost:8080/x+5vpM8W"
					}
				]`,
			},
		},
		{
			name:   "User_Unauthorized",
			method: http.MethodPost,
			userID: uuid.UUID{},
			requestBody: `[
				{
					"correlation_id": "ssdfdsfsfsd",
					"original_url": "https://ya.ru"
				}				
			] `,
			want: want{
				statusCode:  http.StatusUnauthorized,
				responseURL: "",
			},
		},
		{
			name:   "Bad_URL",
			method: http.MethodPost,
			userID: auth.EnsureRandom(),
			requestBody: `[
				{
					"correlation_id": "ssdfdsfsfsd",
					"original_url": "gdfgdg dfgdfgfd"
				}				
			] `,
			want: want{
				statusCode:  http.StatusBadRequest,
				responseURL: "",
			},
		},
		{
			name:   "Error_URLs_Save",
			method: http.MethodPost,
			userID: auth.EnsureRandom(),
			requestBody: `[
				{
					"correlation_id": "ssdfdsfsfsd",
					"original_url": "https://ya.ru"
				},
				{
					"correlation_id": "ssdfdsfsfsd",
					"original_url": "https://ya.ru"
				}
			]`,
			want: want{
				statusCode:  http.StatusBadRequest,
				responseURL: ``,
			},
		},
	}

	urlsForSaveErrors := []models.StorageURL{
		{
			UserID:      nil,
			OriginalURL: "https://ya.ru",
			ShortURL:    "6YGS4ZUF",
		},
		{
			UserID:      nil,
			OriginalURL: "https://ya.ru",
			ShortURL:    "6YGS4ZUF",
		},
	}

	err := config.ParseFlags() //@TODO: Добавить моки конфига
	if err != nil {
		return
	}

	mockCtrl := gomock.NewController(t)
	mockStorageDB := mocks.NewMockStorage(mockCtrl)
	defer mockCtrl.Finish()

	storageURLs := storage.Storage{Storage: mockStorageDB}
	handler := New(storageURLs)

	savedUrls := []string{"6YGS4ZUF", "x+5vpM8W"}
	mockStorageDB.EXPECT().Get("https://ya.ru").Return("6YGS4ZUF", nil).Times(3)
	mockStorageDB.EXPECT().Get("https://dzen.ru").Return("x+5vpM8W", nil)
	mockStorageDB.EXPECT().SaveBatch(gomock.Any(), urlsForSaveErrors, gomock.Any()).Return(nil, errors.New("ошибка при вставке записей"))
	mockStorageDB.EXPECT().SaveBatch(gomock.Any(), gomock.Any(), gomock.Any()).Return(savedUrls, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.requestBody)
			r := httptest.NewRequest(tt.method, "/shorten/batch", body)

			rctx := context.WithValue(r.Context(), auth.AuthKey, tt.userID)
			r = r.WithContext(rctx)
			r.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			fn := handler.SaveBatch() //@TODO: Добавить моки БД
			fn(w, r)

			result := w.Result()

			resBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()

			require.NoError(t, err)

			if tt.want.responseURL != "" {
				assert.JSONEq(t, tt.want.responseURL, w.Body.String(), "Ожидаемый URL %s не совпадает с фактическим %s", tt.want.responseURL, string(resBody))
			}

			assert.Equal(t, tt.want.statusCode, w.Code, "Ожидаемый код ответа %s не совпадаем с фактических %s", tt.want.statusCode, result.StatusCode)
		})
	}
}

func TestGzipCompression(t *testing.T) {
	config.ParseFlags()
	s := storage.Init(config.Options.FlagDatabaseDsn, config.Options.FlagFileStorage)
	h := New(*s)

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
func BenchmarkHandlers_ShortURL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ShortURL("https://ya.ru")
	}
}
