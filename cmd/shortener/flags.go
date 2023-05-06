package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	FlagRunAddr     string `env:"SERVER_ADDRESS"`
	RedirectBaseURL string `env:"BASE_URL"`
}

var config ServerConfig

func parseFlags() {
	flag.StringVar(&config.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&config.RedirectBaseURL, "b", "http://localhost:8080", "server uri prefix")
	flag.Parse()

	err := env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}
}
