package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/handlers"
	"github.com/romanp1989/go-shortener/internal/logger"
	"github.com/romanp1989/go-shortener/internal/route"
	"github.com/romanp1989/go-shortener/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

type App struct {
	flagRunPort string
	chi         *chi.Mux
}

func RunServer() error {
	server := NewApp()
	return server.ListenAndServe()
}

func NewApp() *http.Server {
	err := config.ParseFlags()
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	if err := logger.Initialize(config.Options.FlagLogLevel); err != nil {
		logger.Log.Fatal(err.Error())
	}

	logger.Log.Info("Running server on ", zap.String("port", config.Options.FlagRunPort))

	s := storage.Init(config.Options.FlagDatabaseDsn, config.Options.FlagFileStorage)

	h := handlers.New(*s)

	deleteHandler, err := handlers.NewDelete(s)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	r := route.New(h, deleteHandler)

	return &http.Server{
		Addr:    config.Options.FlagRunPort,
		Handler: r,
	}
}
