package grpc

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PingDb handler for ping server connection
func (gh *GRPCHandlers) PingDB(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	if err := gh.appService.PingDB(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "error database connect ping: %v", err)
	}

	return &empty.Empty{}, nil
}
