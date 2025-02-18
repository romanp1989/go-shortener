package interceptors

import (
	"context"
	"github.com/romanp1989/go-shortener/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor GRPC interceptor for authentication
func AuthInterceptor(jwt *auth.JWTService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		authHeaders := md.Get("auth")
		if len(authHeaders) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing auth header")
		}

		uid, err := jwt.GetUserID(authHeaders[0])
		if uid == nil || err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}

		md.Set(jwt.TokenName, uid.String())
		ctx = auth.Context(ctx, *uid)

		return handler(metadata.NewOutgoingContext(ctx, md), req)
	}

}
