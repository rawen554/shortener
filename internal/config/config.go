package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	FlagRunAddr     string `env:"SERVER_ADDRESS"`
	RedirectBaseURL string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

var config ServerConfig

func ParseFlags() (*ServerConfig, error) {
	flag.StringVar(&config.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&config.RedirectBaseURL, "b", "http://localhost:8080", "server URI prefix")
	flag.StringVar(&config.FileStoragePath, "f", "/tmp/short-url-db.json", "file storage path")
	flag.Parse()

	return &config, env.Parse(&config)
}
