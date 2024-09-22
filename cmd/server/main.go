package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/eac0de/getmetrics/internal/api/router"
	"github.com/eac0de/getmetrics/internal/api/server"
	"github.com/eac0de/getmetrics/internal/config"
	"github.com/eac0de/getmetrics/internal/storage/fileservice"
	"github.com/eac0de/getmetrics/internal/storage/memstore"
	"github.com/eac0de/getmetrics/internal/storage/pgstore"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadAppConfig()
	if err != nil {
		log.Fatal(err)
	}
	var metricStore router.IMetricsStore
	var database router.IDatabase

	pgStore, err := pgstore.New(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Printf("database connection error: %s\n", err.Error())
		memStore := memstore.New()
		metricStore = memStore
		fileService, err := fileservice.New(memStore, cfg.FileStoragePath)
		if err != nil {
			log.Printf("fileservice init error: %s\n", err.Error())
		} else {
			go fileService.StartSavingMetrics(ctx, cfg.StoreInterval)
			defer fileService.SaveMetrics()
		}
	} else {
		metricStore = pgStore
		database = pgStore
		defer pgStore.Close()
	}

	r := router.New(metricStore, database, cfg.SecretKey)
	s := server.New(cfg.Addr)
	go s.Run(r)
	log.Printf("Server http://%s is running. Press Ctrl+C to stop", s.Addr)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	<-sigChan
}
