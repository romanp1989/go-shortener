package app

import (
	"fmt"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/handlers"
	"log"
	"net/http"
)

func RunServer() error {
	err := config.ParseFlags()
	if err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Println("Running server on ", config.Options.FlagRunPort)

	return http.ListenAndServe(config.Options.FlagRunPort, handlers.ShortenerRouter())
}
