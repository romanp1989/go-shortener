package handlers

import (
	"context"
	"github.com/romanp1989/go-shortener/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func (h *Handlers) PingDB() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err := h.storage.Ping(ctx); err != nil {
			logger.Log.Debug("error database connect ping", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}

	return http.HandlerFunc(fn)
}
