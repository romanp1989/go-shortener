package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/models"
	"github.com/romanp1989/go-shortener/internal/models/mocks"
	"github.com/romanp1989/go-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlers_GetURLs(t *testing.T) {
	cfg := &config.ConfigENV{
		ServerAddress: "http://localhost:8080",
		BaseURL:       "http://localhost:8080",
	}
	firstUserID := auth.EnsureRandom()
	firstURLs := []models.StorageURL{
		{
			UserID:      &firstUserID,
			OriginalURL: "https://ya.ru",
			ShortURL:    "6YGS4ZUF",
		},
	}
	firstURLResponse := []models.StorageURL{
		{
			UserID:      nil,
			OriginalURL: firstURLs[0].OriginalURL,
			ShortURL:    fmt.Sprintf("%s/%s", cfg.BaseURL, firstURLs[0].ShortURL),
		},
	}
	firstResponse, _ := json.Marshal(firstURLResponse)

	secondUserID := auth.EnsureRandom()

	type want struct {
		statusCode  int
		responseURL string
	}

	tests := []struct {
		name     string
		userID   uuid.UUID
		shortURL string
		want     want
	}{
		{
			name:   "Success_request",
			userID: firstUserID,
			want: want{
				statusCode:  http.StatusOK,
				responseURL: string(firstResponse),
			},
		},
		{
			name:   "User_Unauthorized",
			userID: uuid.UUID{},
			want: want{
				statusCode:  http.StatusUnauthorized,
				responseURL: "",
			},
		},
		{
			name:   "Empty_URLs",
			userID: secondUserID,
			want: want{
				statusCode:  http.StatusNoContent,
				responseURL: "",
			},
		},
	}

	mockCtrl := gomock.NewController(t)
	mockStorageDB := mocks.NewMockStorage(mockCtrl)
	defer mockCtrl.Finish()

	storageURLs := storage.Storage{Storage: mockStorageDB}
	handler := New(storageURLs, cfg)

	mockStorageDB.EXPECT().GetAllUrlsByUser(gomock.Any(), &firstUserID).Return(firstURLs, nil).Times(1)
	mockStorageDB.EXPECT().GetAllUrlsByUser(gomock.Any(), &secondUserID).Return(make([]models.StorageURL, 0), errors.New("Ошибка при получении urls пользователя")).Times(1)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := httptest.NewRequest(http.MethodGet, "/{id}", nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userID.String())

			contextReq := context.WithValue(body.Context(), chi.RouteCtxKey, rctx)
			contextReq = context.WithValue(contextReq, auth.AuthKey, tt.userID)
			r := body.WithContext(contextReq)
			fn := handler.GetURLs()
			fn(w, r)

			result := w.Result()
			resBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()

			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode, "Ожидаемый код ответа %s не совпадаем с фактических %s", tt.want.statusCode, result.StatusCode)

			if tt.want.responseURL != "" {
				assert.Equal(t, tt.want.responseURL, string(resBody), "Ожидаемый URL %s не совпадает с фактическим %s", tt.want.responseURL, string(resBody))
			}
		})
	}
}
