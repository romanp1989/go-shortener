package grpc

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/romanp1989/go-shortener/internal/grpc/proto/shortener"
	"github.com/romanp1989/go-shortener/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetStats Get statistic for URLs
func (gh *GRPCHandlers) GetStats(ctx context.Context, req *empty.Empty) (*shortener.ResponseGetStats, error) {
	stats, err := gh.appService.GetStats(ctx)
	if err != nil {
		logger.Log.Debug("Ошибка при получении статистики", zap.Error(err))
		//w.WriteHeader(http.StatusNoContent)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &shortener.ResponseGetStats{
		Urls:  stats.URLs,
		Users: stats.Users,
	}, nil
}
