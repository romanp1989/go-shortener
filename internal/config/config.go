package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

// Options console flag configuration struct
var Options struct {
	FlagRunPort     string
	FlagShortURL    string
	FlagLogLevel    string
	FlagFileStorage string
	FlagDatabaseDsn string
	FlagSecretKey   string
}

// ConfigENV env configuration params
type ConfigENV struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	LogLevel      string `env:"LOG_LEVEL"`
	FileStorage   string `env:"FILE_STORAGE_PATH"`
	DatabaseDsn   string `env:"DATABASE_DSN"`
	SecretKey     string `env:"SECRET_KEY"`
}

// ParseFlags function for parse application run flags
func ParseFlags() error {
	if Options.FlagRunPort != "" {
		return nil
	}

	flag.StringVar(&Options.FlagRunPort, "a", ":8080", "port to run server")
	flag.StringVar(&Options.FlagShortURL, "b", "http://localhost:8080", "address to run server")
	flag.StringVar(&Options.FlagLogLevel, "l", "info", "log level")
	flag.StringVar(&Options.FlagFileStorage, "f", "/tmp/shortener.txt", "file storage")
	flag.StringVar(&Options.FlagDatabaseDsn, "d", "", "Database DSN")
	flag.StringVar(&Options.FlagSecretKey, "sk", "verycomplexsecretkey", "Secret key")
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

	if cfg.FileStorage != "" {
		Options.FlagFileStorage = cfg.FileStorage
	}

	if cfg.DatabaseDsn != "" {
		Options.FlagDatabaseDsn = cfg.DatabaseDsn
	}
	if cfg.SecretKey != "" {
		Options.FlagSecretKey = cfg.SecretKey
	}

	return nil
}
