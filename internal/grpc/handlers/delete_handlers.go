package grpc

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/grpc/proto/shortener"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteURLs function for delete urls
func (gh *GRPCHandlers) DeleteURLs(ctx context.Context, req *shortener.RequestDeleteURLs) (*empty.Empty, error) {
	userID := auth.UIDFromContext(ctx)
	if userID == nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	go gh.appService.DeleteURLs(userID, req.GetShortUrls())

	return &empty.Empty{}, nil
}
