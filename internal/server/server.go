package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/eac0de/getmetrics/internal/config"
	"github.com/eac0de/getmetrics/internal/handlers"
	"github.com/eac0de/getmetrics/internal/logger"
	"github.com/eac0de/getmetrics/internal/middlewares"
	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

type metricsService struct {
	conf    *config.HTTPServerConfig
	storage *storage.MetricsStorage
}

func NewMetricsService(
	conf *config.HTTPServerConfig,
	storage *storage.MetricsStorage) *metricsService {
	return &metricsService{
		conf:    conf,
		storage: storage,
	}
}

func (s *metricsService) Stop(cancel context.CancelFunc) {
	err := s.storage.SaveMetricsToFile(s.conf.FileStoragePath)
	if err != nil {
		fmt.Printf("saving metrics error: %s", err.Error())
	}
	cancel()
	log.Println("Server stopped.")
}

func (s *metricsService) Run(ctx context.Context) {
	logger.InitLogger(s.conf.LogLevel)

	err := s.storage.LoadMetricsFromFile(s.conf.FileStoragePath)
	if err != nil {
		log.Printf("load metrics error: %s", err.Error())
	}
	go s.storage.StartSavingMetricsToFile(ctx, s.conf.FileStoragePath, s.conf.StoreInterval)

	r := chi.NewRouter()
	r.Use(middlewares.LoggerMiddleware)
	contentTypesForCompress := "application/json text/html"
	r.Use(middlewares.GetGzipMiddleware(contentTypesForCompress))

	handlers.RegisterMetricsHandlers(r, s.storage)

	log.Printf("Server http://%s is running. Press Ctrl+C to stop.", s.conf.Addr)
	err = http.ListenAndServe(s.conf.Addr, r)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
}
