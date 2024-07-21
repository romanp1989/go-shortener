package app

import (
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/handlers"
	"github.com/romanp1989/go-shortener/internal/logger"
	"go.uber.org/zap"
	"net/http"
)

func RunServer() error {
	err := config.ParseFlags()
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	if err := logger.Initialize(config.Options.FlagLogLevel); err != nil {
		logger.Log.Fatal(err.Error())
	}

	logger.Log.Info("Running server on ", zap.String("port", config.Options.FlagRunPort))

	return http.ListenAndServe(config.Options.FlagRunPort, handlers.ShortenerRouter())
}
