package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/handlers"
	"github.com/romanp1989/go-shortener/internal/middlewares"
)

// New Factory for create routes
func New(h handlers.Handlers, delete *handlers.DeleteBatch) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middlewares.GzipMiddleware)
	r.Use(middlewares.WithLogging)

	r.With(middlewares.AuthMiddlewareSet).Post("/", h.Encode())
	r.Get("/{id}", h.Decode())
	r.Get("/ping", h.PingDB())
	r.Route("/api", func(r chi.Router) {
		r.With(middlewares.AuthMiddlewareRead).Get("/user/urls", h.GetURLs())
		r.With(middlewares.AuthMiddlewareRead).Delete("/user/urls", delete.DeleteURLs())
		r.Route("/shorten", func(r chi.Router) {
			r.With(middlewares.AuthMiddlewareSet).Post("/", h.Shorten())
			r.With(middlewares.AuthMiddlewareSet).Post("/batch", h.SaveBatch())
		})

	})

	return r
}
