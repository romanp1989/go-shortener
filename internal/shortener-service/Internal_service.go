package shortenerservice

import (
	"context"
	"github.com/romanp1989/go-shortener/internal/logger"
	"github.com/romanp1989/go-shortener/internal/models"
	"go.uber.org/zap"
)

// GetStats Get statistic for URLs
func (s *ShortenerService) GetStats(ctx context.Context) (models.StorageStats, error) {
	stats, err := s.storage.GetStats(ctx)
	if err != nil {
		logger.Log.Debug("Ошибка при получении статистики", zap.Error(err))
		return models.StorageStats{}, nil
	}

	return stats, nil

}
