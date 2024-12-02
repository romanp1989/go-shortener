package logger

import (
	"go.uber.org/zap"
	"net/http"
)

// Logger types
type (
	// ResponseData structure for response data
	ResponseData struct {
		Status int
		Size   int
	}

	// LoggingResponseWriter log writer
	LoggingResponseWriter struct {
		http.ResponseWriter
		ResponseData *ResponseData
	}
)

// Log logger
var Log *zap.Logger = zap.NewNop()

// Initialize initializes the logger singleton with the required logging level
func Initialize(level string) error {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень
	cfg.Level = lvl
	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	// устанавливаем синглтон
	Log = zl
	return nil
}

// Write function logging the response
func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.ResponseData.Size += size // захватываем размер
	return size, err
}

// WriteHeader function logging the response status code
func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.Status = statusCode // захватываем код статуса
}
