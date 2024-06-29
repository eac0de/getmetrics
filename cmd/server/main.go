package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/eac0de/getmetrics/internal/server"
)

type HTTPServerConfig struct {
	Addr     string `env:"ADDRESS"`
	LogLevel string `env:"ADDRESS"`
}

const (
	defaultAddr     = "localhost:8080"
	defaultLogLevel = "info"
)

func readServerFlags(c *HTTPServerConfig) {
	flag.StringVar(&c.Addr, "a", defaultAddr, "server address")
	flag.StringVar(&c.LogLevel, "ll", defaultLogLevel, "server log level")

	flag.Parse()
}

func readEnvConfig(c *HTTPServerConfig) {
	err := env.Parse(c)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	serverConfig := new(HTTPServerConfig)
	readServerFlags(serverConfig)
	readEnvConfig(serverConfig)
	s := server.NewMetricsServer(serverConfig.Addr)
	s.Run(serverConfig.LogLevel)
}
