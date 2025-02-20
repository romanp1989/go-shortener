package app

import (
	"context"
	"crypto/tls"
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/config"
	grpcHandlers "github.com/romanp1989/go-shortener/internal/grpc/handlers"
	"github.com/romanp1989/go-shortener/internal/grpc/interceptors"
	"github.com/romanp1989/go-shortener/internal/grpc/proto"
	"github.com/romanp1989/go-shortener/internal/handlers"
	"github.com/romanp1989/go-shortener/internal/logger"
	"github.com/romanp1989/go-shortener/internal/route"
	shortener_service "github.com/romanp1989/go-shortener/internal/shortener-service"
	"github.com/romanp1989/go-shortener/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// App Application configuration
type App struct {
	flagRunPort string
	chi         *chi.Mux
}

// RunServer run application server
func RunServer(cfg *config.ConfigENV) error {
	var err error

	errChan := make(chan error, 1)

	if err = logger.Initialize(cfg.LogLevel); err != nil {
		logger.Log.Info(err.Error())
		return err
	}

	logger.Log.Info("Running server on ", zap.String("port", cfg.ServerAddress))

	s := storage.Init(cfg.DatabaseDsn, cfg.FileStorage)

	appService := shortener_service.NewShortenerService(s, cfg)
	jwtService := auth.NewJwtService(cfg.SecretKey)

	h := handlers.New(appService)

	r := route.New(appService, h, jwtService)

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	if cfg.HTTPS.Enable {
		srv.TLSConfig = &tls.Config{}

		go func() {
			if err := srv.ListenAndServeTLS(
				cfg.HTTPS.Pem,
				cfg.HTTPS.Key,
			); err != nil {
				logger.Log.Fatal(err.Error())
				errChan <- err
			}
		}()
	} else {
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				logger.Log.Info(err.Error())
				errChan <- err
			}
		}()
	}

	for {
		select {
		case err := <-errChan:
			return err
		case <-sig:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				logger.Log.Fatal("HTTP Server Shutdown error: %v", zap.String("error", err.Error()))
				return err
			}

			return nil
		}

	}
}

func RunGRPCServer(cfg *config.ConfigENV) error {
	var err error

	s := storage.Init(cfg.DatabaseDsn, cfg.FileStorage)
	jwtService := auth.NewJwtService(cfg.SecretKey)

	appService := shortener_service.NewShortenerService(s, cfg)
	h := grpcHandlers.New(appService)
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor([]grpc.UnaryServerInterceptor{
		interceptors.AuthInterceptor(jwtService),
		interceptors.IPRestrictionInterceptor(cfg.TrustedSubnet),
	}...))
	proto.RegisterInternalServer(grpcServer, h)

	listener, err := net.Listen("tcp", cfg.GRPCServerAddress)
	if err != nil {
		logger.Log.Fatal("Failed to start gRPC server", zap.Error(err))
		return err
	}
	logger.Log.Info("Starting gRPC server", zap.String("address", cfg.GRPCServerAddress))

	if err := grpcServer.Serve(listener); err != nil {
		logger.Log.Fatal("gRPC server encountered an error", zap.Error(err))
		return err
	}

	return nil
}

func ReadConfig() (*config.ConfigENV, error) {
	cfg, err := config.ParseFlags()
	if err != nil {
		logger.Log.Info(err.Error())
		return nil, err
	}

	return cfg, err
}
