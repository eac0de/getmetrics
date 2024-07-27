package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/eac0de/getmetrics/app/server"
	"github.com/eac0de/getmetrics/internal/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serverConfig := config.NewHTTPServerConfig()
	s := server.NewMetrciServerApp(ctx, serverConfig)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	var wg sync.WaitGroup
	wg.Add(1)
	go s.Run()
	go s.Stop(ctx, &wg)
	<-sigChan
	cancel()
	wg.Wait()
}
