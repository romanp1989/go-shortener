package config

import "flag"

var Options struct {
	FlagRunPort  string
	FlagShortURL string
}

func ParseFlags() {
	flag.StringVar(&Options.FlagRunPort, "a", ":8080", "port to run server")
	flag.StringVar(&Options.FlagShortURL, "b", "http://localhost:8080", "address to run server")
	flag.Parse()
}
