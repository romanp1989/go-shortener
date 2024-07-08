package main

import (
	"fmt"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/handlers"
	"net/http"
)

func main() {
	config.ParseFlags()

	fmt.Println("Running server on ", config.Options.FlagRunPort)
	err := http.ListenAndServe(config.Options.FlagRunPort, handlers.ShortenerRouter())
	if err != nil {
		panic(err)
	}
}
