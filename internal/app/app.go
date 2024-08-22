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
	app := NewApp()
	return http.ListenAndServe(app.flagRunPort, app.chi)
}

func NewApp() *App {
	err := config.ParseFlags()
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	if err := logger.Initialize(config.Options.FlagLogLevel); err != nil {
		logger.Log.Fatal(err.Error())
	}

	logger.Log.Info("Running server on ", zap.String("port", config.Options.FlagRunPort))

	s := storage.Init(config.Options.FlagFileStorage)

	h := handlers.New(s)

	r := route.New(h)

	return &App{
		flagRunPort: config.Options.FlagRunPort,
		chi:         r,
	}
}
