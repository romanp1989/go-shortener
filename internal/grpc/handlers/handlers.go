package grpc

import (
	"context"
	"errors"
	"github.com/romanp1989/go-shortener/internal/auth"
	proto "github.com/romanp1989/go-shortener/internal/grpc/proto"
	"github.com/romanp1989/go-shortener/internal/grpc/proto/shortener"
	"github.com/romanp1989/go-shortener/internal/logger"
	shortener_service "github.com/romanp1989/go-shortener/internal/shortener-service"
	"github.com/romanp1989/go-shortener/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/url"
)

// Handlers handlers
type GRPCHandlers struct {
	proto.UnimplementedInternalServer
	appService *shortener_service.ShortenerService
}

// New Factory for create handlers
func New(appService *shortener_service.ShortenerService) *GRPCHandlers {
	return &GRPCHandlers{
		appService: appService,
	}
}

func (gh *GRPCHandlers) Encode(ctx context.Context, req *shortener.RequestEncode) (*shortener.ResponseEncode, error) {
	userID := auth.UIDFromContext(ctx)
	if userID == nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	if _, err := url.ParseRequestURI(req.GetUrl()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	shortURL, err := gh.appService.Encode(ctx, req.GetUrl())

	if err != nil {
		logger.Log.Debug("Ошибка добавления данных", zap.Error(err))

		var errConflict *storage.URLConflictError
		if errors.As(err, &errConflict) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &shortener.ResponseEncode{ShortUrl: shortURL}, nil
}

// Decode handler for getting the original URL from short URL
func (gh *GRPCHandlers) Decode(ctx context.Context, req *shortener.RequestDecode) (*shortener.ResponseDecode, error) {
	id := req.GetUrl()

	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	fullURL, err := gh.appService.Decode(id)
	if err != nil {
		var errURLDeleted *storage.AlreadyDeleted
		if errors.As(err, &errURLDeleted) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		logger.Log.Debug("error get url response", zap.Error(err))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if fullURL != "" {
		return &shortener.ResponseDecode{Result: fullURL}, nil
	}

	return nil, status.Error(codes.NotFound, "url not found")
}

// Shorten handler for creating a shortened URL based on the original one
func (gh *GRPCHandlers) Shorten(ctx context.Context, req *shortener.RequestShorten) (*shortener.ResponseShorten, error) {
	logger.Log.Debug("decoding request")

	userID := auth.UIDFromContext(ctx)
	if userID == nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	shortURL, err := gh.appService.Shorten(ctx, req.GetUrl(), userID)
	if err != nil {
		logger.Log.Debug("Ошибка добавления данных", zap.Error(err))

		var errConflict *storage.URLConflictError
		if errors.As(err, &errConflict) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &shortener.ResponseShorten{Result: shortURL}, nil
}

// SaveBatch handler for creating a shortened URL based on the original one
func (gh *GRPCHandlers) SaveBatch(ctx context.Context, req *shortener.RequestSaveBatch) (*shortener.ResponseSaveBatch, error) {
	var err error
	var shortURL string

	userID := auth.UIDFromContext(ctx)
	if userID == nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	response := &shortener.ResponseSaveBatch{Items: make([]*shortener.Item, len(req.GetItems()))}

	for _, value := range req.GetItems() {
		if _, err = url.ParseRequestURI(value.GetUrl()); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		shortURL, err = gh.appService.Shorten(ctx, value.GetUrl(), userID)

		response.Items = append(response.Items, &shortener.Item{
			CorrelationId: value.GetCorrelationId(),
			Url:           shortURL,
		})
	}

	return response, nil
}
