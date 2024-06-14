package main

import (
	"flag"

	s "github.com/eac0de/getmetrics/internal/server"
)

var addr string

func parseFlags() {
	flag.StringVar(&addr, "a", "localhost:8080", "server address")

	flag.Parse()
}

func main() {
	parseFlags()
	server := s.NewServer(addr)
	server.Run()
}
