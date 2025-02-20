package handlers

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/middlewares"
	shortener_service "github.com/romanp1989/go-shortener/internal/shortener-service"
	"github.com/romanp1989/go-shortener/internal/storage"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

// Rrequest example for route /api/shorten
func Example_handlers_Shorten() {
	store := storage.Init("", "./shortener.txt")
	cfg := &config.ConfigENV{
		ServerAddress: ":8080",
		BaseURL:       "http://localhost:8080",
	}

	appService := shortener_service.NewShortenerService(store, cfg)
	handlers := New(appService)

	router := chi.NewRouter()

	srv := httptest.NewServer(router)
	defer srv.Close()

	m := middlewares.Middleware{
		Cfg: cfg,
	}

	router.Use(m.GzipMiddleware)
	router.Use(m.WithLogging)
	router.Post("/api/shorten", handlers.Shorten())

	requestBody := `{"url": "https://ya.ru"}`
	statusCode, err := requestExample(srv, http.MethodPost, "/api/shorten", requestBody)
	if err != nil {
		log.Fatal(err)
	}

	if statusCode != http.StatusCreated {
		log.Fatalf("expected status code 201, got %d", statusCode)
	}
}

func requestExample(server *httptest.Server, method string, path string, requestBody string) (int, error) {
	body := strings.NewReader(requestBody)
	r := httptest.NewRequest(method, path, body)

	jwtService := auth.NewJwtService("verycomplexsecretkey")
	userID := jwtService.EnsureRandom()

	rctx := context.WithValue(r.Context(), auth.AuthKey, userID)
	r = r.WithContext(rctx)
	r.Header.Set("Content-Type", "application/json")

	resp, err := server.Client().Do(r)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(resp.Body)

	return resp.StatusCode, nil
}
