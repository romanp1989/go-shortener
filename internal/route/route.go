package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/compress"
	"github.com/romanp1989/go-shortener/internal/handlers"
	"github.com/romanp1989/go-shortener/internal/logger"
)

func New(h handlers.Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(compress.GzipMiddleware)
	r.Use(logger.WithLogging)

	r.Post("/", h.Encode())
	r.Get("/{id}", h.Decode())
	r.Post("/api/shorten", h.Shorten())
	r.Get("/ping", h.PingDB())
	r.Post("/api/shorten/batch", h.SaveBatch())

	return r
}
