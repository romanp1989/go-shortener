package main

import (
	"github.com/romanp1989/go-shortener/internal/app"
	"log"
)

func main() {
	if err := app.RunServer(); err != nil {
		log.Fatalf(err.Error())
	}
}
