package handlers

import (
	"context"
	"encoding/json"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// GetURLs handler for creating a shortened URL based on the original one
// @Accept string user uuid
// @Success 200 {json} list of user's URLs
// @Failure 204 no content if users haven't URLs
// @Failure 401 error if user unauthorized
func (h *Handlers) GetURLs() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		userID := auth.UIDFromContext(ctx)
		if userID == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		allUrls, err := h.appService.GetURLs(ctx, userID)
		if err != nil {
			logger.Log.Debug("Ошибка при получении urls пользователя", zap.Error(err))
			w.WriteHeader(http.StatusNoContent)
			return
		}

		b, err := json.Marshal(allUrls)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)

		logger.Log.Debug("sending HTTP 200 response")
	}

	return http.HandlerFunc(fn)
}
