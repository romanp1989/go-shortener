package app

import (
	"context"
	"crypto/tls"
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/handlers"
	"github.com/romanp1989/go-shortener/internal/logger"
	"github.com/romanp1989/go-shortener/internal/route"
	"github.com/romanp1989/go-shortener/internal/storage"
	"go.uber.org/zap"
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
func RunServer() error {
	cfg, err := config.ParseFlags()
	if err != nil {
		logger.Log.Info(err.Error())
		return err
	}

	errChan := make(chan error, 1)

	if err = logger.Initialize(cfg.LogLevel); err != nil {
		logger.Log.Info(err.Error())
		return err
	}

	logger.Log.Info("Running server on ", zap.String("port", cfg.ServerAddress))

	s := storage.Init(cfg.DatabaseDsn, cfg.FileStorage)

	h := handlers.New(*s, cfg)

	deleteHandler, err := handlers.NewDelete(s)
	if err != nil {
		logger.Log.Info(err.Error())
		return err
	}

	r := route.New(h, deleteHandler)

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
