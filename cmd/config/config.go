package config

import "flag"

type serverConfig struct {
	FlagRunAddr     string
	RedirectBaseURL string
}

var Config serverConfig

func ParseFlags() {
	flag.StringVar(&Config.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&Config.RedirectBaseURL, "b", "http://localhost:8080", "server uri prefix")
	flag.Parse()
}
