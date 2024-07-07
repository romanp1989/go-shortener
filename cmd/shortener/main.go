package main

import (
	"github.com/romanp1989/go-shortener/internal/handlers"
	"net/http"
)

func main() {
	//mux := http.NewServeMux()
	//mux.HandleFunc("/", http.HandlerFunc(handlers.Encode))
	//mux.HandleFunc("/{id}", http.HandlerFunc(handlers.Decode))

	err := http.ListenAndServe(":8080", handlers.ShortenerRouter())
	if err != nil {
		panic(err)
	}
}
