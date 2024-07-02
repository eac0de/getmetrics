package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/eac0de/getmetrics/internal/server"
)

type HTTPServerConfig struct {
	Addr            string `env:"ADDRESS"`
	LogLevel        string `env:"LOG_LEVEL"`
	StoreInterval   time.Duration
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

type EnvHTTPServerConfig struct {
	HTTPServerConfig
	StoreInterval int `env:"STORE_INTERVAL"`
}

const (
	defaultAddr            = "localhost:8080"
	defaultLogLevel        = "info"
	defaultStoreInterval   = 300
	defaultFileStoragePath = "/tmp/metrics-db.json"
	defaultRestore         = true
)

func readServerFlags(c *HTTPServerConfig) {
	var storeInterval int
	flag.StringVar(&c.Addr, "a", defaultAddr, "server address")
	flag.StringVar(&c.LogLevel, "ll", defaultLogLevel, "server log level")
	flag.IntVar(&storeInterval, "i", defaultStoreInterval, "server store interval")
	flag.StringVar(&c.FileStoragePath, "f", defaultFileStoragePath, "server file restore path")
	flag.BoolVar(&c.Restore, "r", defaultRestore, "server restore")
	flag.Parse()
	c.StoreInterval = time.Duration(storeInterval) * time.Second

}

func readEnvConfig(c *HTTPServerConfig) {
	DurationToInt := func(d time.Duration) int {
		return int(d.Seconds())
	}
	envConfig := EnvHTTPServerConfig{
		HTTPServerConfig: *c,
		StoreInterval:    DurationToInt(c.StoreInterval),
	}
	err := env.Parse(&envConfig)
	if err != nil {
		log.Fatal(err)
	}
	c.Addr = envConfig.Addr
	c.LogLevel = envConfig.LogLevel
	c.FileStoragePath = envConfig.FileStoragePath
	c.Restore = envConfig.Restore
	c.StoreInterval = time.Duration(envConfig.StoreInterval) * time.Second
}

func main() {
	serverConfig := new(HTTPServerConfig)
	readServerFlags(serverConfig)
	readEnvConfig(serverConfig)
	fmt.Println("Addr:", serverConfig.Addr)
	fmt.Println("LogLevel:", serverConfig.LogLevel)
	fmt.Println("FileStoragePath:", serverConfig.FileStoragePath)
	fmt.Println("Restore:", serverConfig.Restore)
	fmt.Println("StoreInterval:", serverConfig.StoreInterval)
	s := server.NewMetricsServer(
		serverConfig.Addr,
		serverConfig.LogLevel,
		serverConfig.FileStoragePath,
		serverConfig.Restore,
		serverConfig.StoreInterval,
	)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	go func() {
		s.Run()
	}()
	<-sigChan
	s.Stop()
}
