package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/eac0de/getmetrics/internal/config"
	"github.com/eac0de/getmetrics/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serverConfig := config.NewHTTPServerConfig()
	s := server.NewMetricsService(serverConfig)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	go func() {
		s.Run(ctx)
	}()
	<-sigChan
	s.Stop(cancel)
}
