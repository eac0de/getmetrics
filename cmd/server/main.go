package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"github.com/eac0de/getmetrics/internal/api/handlers"
	"github.com/eac0de/getmetrics/internal/api/server"
	"github.com/eac0de/getmetrics/internal/config"
	"github.com/eac0de/getmetrics/internal/storage/fileservice"
	"github.com/eac0de/getmetrics/internal/storage/memstore"
	"github.com/eac0de/getmetrics/internal/storage/pgstore"
	"github.com/eac0de/getmetrics/pkg/middlewares"
	"github.com/eac0de/getmetrics/pkg/utils"
	"github.com/go-chi/chi/v5"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func setupRouter(
	metricsStore handlers.IMetricsStore,
	database handlers.IDatabase,
	secretKey string,
) *chi.Mux {
	mh := handlers.NewMetricsHandlers(metricsStore, secretKey)
	dh := handlers.NewDatabaseHandlers(database)

	r := chi.NewRouter()
	r.Use(middlewares.LoggerMiddleware)
	r.Use(middlewares.GetCheckSignMiddleware(secretKey))
	contentTypesForCompress := "application/json text/html"
	r.Use(middlewares.GetGzipMiddleware(contentTypesForCompress))

	r.Get("/", mh.ShowMetricsSummaryHandler())
	r.Post("/update/{metricType}/{metricName}/{metricValue}", mh.UpdateMetricHandler())
	r.Post("/update/", mh.UpdateMetricJSONHandler())
	r.Post("/updates/", mh.UpdateMetricsJSONHandler())
	r.Get("/value/{metricType}/{metricName}", mh.GetMetricHandler())
	r.Post("/value/", mh.GetMetricJSONHandler())

	r.Get("/ping", dh.PingHandler())
	return r
}

func main() {
	fmt.Printf("Build version: %s\n", utils.GetValueOrDefault(buildVersion))
	fmt.Printf("Build date: %s\n", utils.GetValueOrDefault(buildDate))
	fmt.Printf("Build commit: %s\n", utils.GetValueOrDefault(buildCommit))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadAppConfig()
	if err != nil {
		log.Fatal(err)
	}
	var metricStore handlers.IMetricsStore
	var database handlers.IDatabase

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

	r := setupRouter(metricStore, database, cfg.SecretKey)
	s := server.New(cfg.Addr)
	go func() {
		// Запускаем pprof на отдельном порту, если это необходимо
		http.ListenAndServe(":6060", nil)
	}()
	if cfg.PrivateKeyPath == "" {
		go s.Run(r)
		log.Printf("Server http://%s is running. Press Ctrl+C to stop", s.Addr)
	} else {
		go s.RunTLS(r, cfg.PrivateKeyPath)
		log.Printf("Server https://%s is running. Press Ctrl+C to stop", s.Addr)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	<-sigChan
}
