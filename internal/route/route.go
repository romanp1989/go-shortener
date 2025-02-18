package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/handlers"
	"github.com/romanp1989/go-shortener/internal/middlewares"
)

// New Factory for create routes
func New(h handlers.Handlers, delete *handlers.DeleteBatch) *chi.Mux {
	r := chi.NewRouter()
	m := middlewares.Middleware{
		Cfg: h.Cfg,
	}

	r.Use(m.GzipMiddleware)
	r.Use(m.WithLogging)

	r.With(m.AuthMiddlewareSet).Post("/", h.Encode())
	r.Get("/{id}", h.Decode())
	r.Get("/ping", h.PingDB())
	r.Route("/api", func(r chi.Router) {
		r.With(m.AuthMiddlewareRead).Get("/user/urls", h.GetURLs())
		r.With(m.AuthMiddlewareRead).Delete("/user/urls", delete.DeleteURLs())
		r.Route("/shorten", func(r chi.Router) {
			r.With(m.AuthMiddlewareSet).Post("/", h.Shorten())
			r.With(m.AuthMiddlewareSet).Post("/batch", h.SaveBatch())
		})

	})

	return r
}
