package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/eac0de/getmetrics/internal/config"
	"github.com/eac0de/getmetrics/internal/server"
	"github.com/eac0de/getmetrics/internal/storage"
)

func main() {
	serverConfig := config.NewHTTPServerConfig()
	source, _ := os.Getwd()
	fmt.Println("pwd: ", source)
	storage := storage.NewMetricsStorage()
	s := server.NewMetricsServer(serverConfig, storage)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	go func() {
		s.Run()
	}()
	<-sigChan
	s.Stop()
}
