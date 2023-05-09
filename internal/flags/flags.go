package flags

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	FlagRunAddr     string `env:"SERVER_ADDRESS"`
	RedirectBaseURL string `env:"BASE_URL"`
}

var Config ServerConfig

func ParseFlags() error {
	flag.StringVar(&Config.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&Config.RedirectBaseURL, "b", "http://localhost:8080", "server URI prefix")
	flag.Parse()

	return env.Parse(&Config)
}
