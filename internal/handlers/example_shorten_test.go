package handlers

import (
	"context"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/logger"
	"github.com/romanp1989/go-shortener/internal/route"
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

	handlers := New(*store)
	deleteHandler, err := NewDelete(store)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	router := route.New(handlers, deleteHandler)

	srv := httptest.NewServer(router)
	defer srv.Close()

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

	userID := auth.EnsureRandom()

	rctx := context.WithValue(r.Context(), auth.AuthKey, userID)
	r = r.WithContext(rctx)
	r.Header.Set("Content-Type", "application/json")

	resp, err := server.Client().Do(r)
	if err != nil {
		return 0, err
	}

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
