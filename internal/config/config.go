package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

var Options struct {
	FlagRunPort  string
	FlagShortURL string
	FlagLogLevel string
}

type ConfigENV struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	LogLevel      string `env:"LOG_LEVEL"`
}

func ParseFlags() error {
	flag.StringVar(&Options.FlagRunPort, "a", ":8080", "port to run server")
	flag.StringVar(&Options.FlagShortURL, "b", "http://localhost:8080", "address to run server")
	flag.StringVar(&Options.FlagLogLevel, "l", "info", "log level")
	flag.Parse()

	var cfg ConfigENV

	err := env.Parse(&cfg)
	if err != nil {
		log.Printf("Ошибка при парсинге переменных окружения %s", err.Error())
		return err
	}

	if cfg.ServerAddress != "" {
		Options.FlagRunPort = cfg.ServerAddress
	}

	if cfg.BaseURL != "" {
		Options.FlagShortURL = cfg.BaseURL
	}

	if cfg.LogLevel != "" {
		Options.FlagLogLevel = cfg.LogLevel
	}

	return nil
}
