package main

import (
	"github.com/romanp1989/go-shortener/internal/app"
	"log"
	//_ "net/http/pprof"
)

// main Main function for launch application
func main() {
	if err := app.RunServer(); err != nil {
		var msg = err.Error()
		log.Fatalf("error %s", msg)
	}
}
