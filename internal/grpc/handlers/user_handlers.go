package grpc

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/grpc/proto/shortener"
	"github.com/romanp1989/go-shortener/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetUserURL handler for creating a shortened URL based on the original one
func (gh *GRPCHandlers) GetUserURL(ctx context.Context, req *empty.Empty) (*shortener.ResponseGetUserURL, error) {
	userID := auth.UIDFromContext(ctx)
	if userID == nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	allUrls, err := gh.appService.GetURLs(ctx, userID)
	if err != nil {
		logger.Log.Debug("Ошибка при получении urls пользователя", zap.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	response := &shortener.ResponseGetUserURL{Items: make([]*shortener.UserURL, len(allUrls))}
	for _, url := range allUrls {
		response.Items = append(response.Items, &shortener.UserURL{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		})
	}

	return response, nil
}
