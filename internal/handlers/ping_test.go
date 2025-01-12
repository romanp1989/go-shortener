package handlers

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/models/mocks"
	"github.com/romanp1989/go-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlers_PingDB(t *testing.T) {
	type want struct {
		statusCode int
	}

	tests := []struct {
		name   string
		method string
		want   want
	}{
		{
			name:   "Ping_Success",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "Ping_Unsuccessful",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	mockCtrl := gomock.NewController(t)
	mockStorageDB := mocks.NewMockStorage(mockCtrl)
	defer mockCtrl.Finish()

	storageURLs := storage.Storage{Storage: mockStorageDB}
	cfg, _ := config.ParseFlags()
	handler := New(storageURLs, cfg)

	firstCall := mockStorageDB.EXPECT().Ping(gomock.Any()).Return(nil).Times(1)
	mockStorageDB.EXPECT().Ping(gomock.Any()).After(firstCall).Return(errors.New("error database connect ping"))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()

			fn := handler.PingDB()
			fn(w, r)

			assert.Equal(t, tt.want.statusCode, w.Code, "Ожидаемый код ответа %s не совпадаем с фактических %s", tt.want.statusCode, w.Code)
		})
	}
}
