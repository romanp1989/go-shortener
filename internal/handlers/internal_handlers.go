package handlers

import (
	"context"
	"encoding/json"
	"github.com/romanp1989/go-shortener/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// GetStats Get statistic for URLs
func (h *Handlers) GetStats() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		stats, err := h.appService.GetStats(ctx)
		if err != nil {
			logger.Log.Debug("Ошибка при получении статистики", zap.Error(err))
			w.WriteHeader(http.StatusNoContent)
			return
		}

		b, err := json.Marshal(stats)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}

	return http.HandlerFunc(fn)
}
