package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/handlers"
	"github.com/romanp1989/go-shortener/internal/middlewares"
)

func New(h handlers.Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middlewares.GzipMiddleware)
	r.Use(middlewares.WithLogging)
	r.Use(middlewares.AuthMiddleware)

	r.Post("/", h.Encode())
	r.Get("/{id}", h.Decode())
	r.Get("/ping", h.PingDB())
	r.Route("/api", func(r chi.Router) {
		r.Get("/user/urls", h.GetURLs())
		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", h.Shorten())
			r.Post("/batch", h.SaveBatch())
		})

	})

	return r
}
