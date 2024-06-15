package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	s "github.com/eac0de/getmetrics/internal/server"
)

type ServerConfig struct {
	Addr string `env:"ADDRESS"`
}

func parseFlags(c *ServerConfig) {
	flag.StringVar(&c.Addr, "a", "localhost:8080", "server address")

	flag.Parse()
}

func parseEnv(c *ServerConfig) {
	err := env.Parse(c)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	serverConfig := new(ServerConfig)
	parseFlags(serverConfig)
	parseEnv(serverConfig)
	server := s.NewServer(serverConfig.Addr)
	server.Run()
}
