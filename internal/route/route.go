package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/handlers"
	"github.com/romanp1989/go-shortener/internal/middlewares"
	shortener_service "github.com/romanp1989/go-shortener/internal/shortener-service"
)

// New Factory for create routes
func New(app *shortener_service.ShortenerService, h handlers.Handlers, jwtService *auth.JWTService) *chi.Mux {
	r := chi.NewRouter()

	m := middlewares.Middleware{
		Cfg:        app.Cfg,
		JwtService: jwtService,
	}

	r.Use(m.GzipMiddleware)
	r.Use(m.WithLogging)

	r.With(m.AuthMiddlewareSet).Post("/", h.Encode())
	r.Get("/{id}", h.Decode())
	r.Get("/ping", h.PingDB())
	r.Route("/api", func(r chi.Router) {
		r.With(m.AuthMiddlewareRead).Get("/user/urls", h.GetURLs())
		r.With(m.AuthMiddlewareRead).Delete("/user/urls", h.DeleteURLs())
		r.Route("/shorten", func(r chi.Router) {
			r.With(m.AuthMiddlewareSet).Post("/", h.Shorten())
			r.With(m.AuthMiddlewareSet).Post("/batch", h.SaveBatch())
		})
		r.With(m.ValidateSubnet).Get("/internal/stats", h.GetStats())

	})

	return r
}
