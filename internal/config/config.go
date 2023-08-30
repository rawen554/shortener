package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	RunAddr         string `env:"SERVER_ADDRESS"`
	RedirectBaseURL string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Secret          string `env:"SECRET"`
}

var config ServerConfig

func ParseFlags() (*ServerConfig, error) {
	flag.StringVar(&config.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&config.RedirectBaseURL, "b", "http://localhost:8080", "server URI prefix")
	flag.StringVar(&config.FileStoragePath, "f", "", "file storage path")
	flag.StringVar(&config.DatabaseDSN, "d", "", "Data Source Name (DSN)")
	flag.StringVar(&config.Secret, "s", "b4952c3809196592c026529df00774e46bfb5be0", "Secret")
	flag.Parse()

	if err := env.Parse(&config); err != nil {
		return nil, fmt.Errorf("error parsing env variables: %w", err)
	}

	return &config, nil
}
