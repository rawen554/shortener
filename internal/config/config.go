package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	FlagRunAddr     string `env:"SERVER_ADDRESS"`
	RedirectBaseURL string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Seed            string `env:"SEED"`
}

var config ServerConfig

func ParseFlags() (*ServerConfig, error) {
	flag.StringVar(&config.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&config.RedirectBaseURL, "b", "http://localhost:8080", "server URI prefix")
	flag.StringVar(&config.FileStoragePath, "f", "", "file storage path")
	flag.StringVar(&config.DatabaseDSN, "d", "", "Data Source Name (DSN)")
	flag.StringVar(&config.Seed, "s", "b4952c3809196592c026529df00774e46bfb5be0", "seed")
	flag.Parse()

	return &config, env.Parse(&config)
}
