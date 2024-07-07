package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/eac0de/getmetrics/internal/config"
	"github.com/eac0de/getmetrics/internal/server"
	"github.com/eac0de/getmetrics/internal/storage"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serverConfig := config.NewHTTPServerConfig()
	storage := storage.NewMetricsStorage()
	s := server.NewMetricsService(serverConfig, storage)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	go func() {
		s.Run(ctx)
	}()
	<-sigChan
	s.Stop(cancel)
}
