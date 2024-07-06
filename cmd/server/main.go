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
	fmt.Println("Addr:", serverConfig.Addr)
	fmt.Println("LogLevel:", serverConfig.LogLevel)
	fmt.Println("FileStoragePath:", serverConfig.FileStoragePath)
	fmt.Println("Restore:", serverConfig.Restore)
	fmt.Println("StoreInterval:", serverConfig.StoreInterval)
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
