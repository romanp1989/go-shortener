package interceptors

import (
	"context"
	"fmt"
	"github.com/romanp1989/go-shortener/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
)

// IPRestrictionInterceptor GRPC interceptor for validate ip
func IPRestrictionInterceptor(trustedSubnets string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var subnet *net.IPNet
		var err error

		if trustedSubnets != "" {
			_, subnet, err = net.ParseCIDR(trustedSubnets)
			if err != nil {
				logger.Log.Error("Parse error trusted subnet config: %v", zap.String("error", err.Error()))
				return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Parse error trusted subnet config: %v", zap.String("error", err.Error())))
			}
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "Metadata is missing")
		}

		clientIPHeader := md.Get("X-Real-IP")
		if len(clientIPHeader) == 0 {
			return nil, status.Error(codes.PermissionDenied, "Access Forbidden: Missing X-Real-IP header")
		}

		ip := net.ParseIP(clientIPHeader[0])

		if ip == nil {
			return nil, status.Errorf(codes.PermissionDenied, "Access Forbidden: Invalid X-Real-IP value")
		}

		if !subnet.Contains(ip) {
			return nil, status.Errorf(codes.PermissionDenied, "Access Forbidden: IP not in trusted subnet")
		}

		return handler(ctx, req)
	}

}
