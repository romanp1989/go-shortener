package app

import (
	"crypto/tls"
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/handlers"
	"github.com/romanp1989/go-shortener/internal/logger"
	"github.com/romanp1989/go-shortener/internal/route"
	"github.com/romanp1989/go-shortener/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

// App Application configuration
type App struct {
	flagRunPort string
	chi         *chi.Mux
}

// RunServer run application server
//func RunServer() error {
//	server := NewApp()
//	return server.ListenAndServe()
//}

// RunServer run application server
func RunServer() error {
	cfg, err := config.ParseFlags()
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	if err = logger.Initialize(cfg.LogLevel); err != nil {
		logger.Log.Fatal(err.Error())
	}

	logger.Log.Info("Running server on ", zap.String("port", cfg.ServerAddress))

	s := storage.Init(cfg.DatabaseDsn, cfg.FileStorage)

	h := handlers.New(*s, cfg)

	deleteHandler, err := handlers.NewDelete(s)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	r := route.New(h, deleteHandler)

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	if cfg.HTTPS.Enable {
		srv.TLSConfig = &tls.Config{}

		return srv.ListenAndServeTLS(
			cfg.HTTPS.Pem,
			cfg.HTTPS.Key,
		)
	}

	return srv.ListenAndServe()
}
