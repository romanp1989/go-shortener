package main

import (
	"github.com/romanp1989/go-shortener/internal/app"
	"log"
	//_ "net/http/pprof"
)

func main() {
	//go func() {
	//	log.Println(http.ListenAndServe(":6060", nil))
	//}()

	if err := app.RunServer(); err != nil {
		var msg = err.Error()
		log.Fatalf("error %s", msg)
	}
}
