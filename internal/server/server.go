package server

import (
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
	exit    chan struct{}
	storage *storage.MetricsStorage
}

func NewMetricsService(
	conf *config.HTTPServerConfig,
	store *storage.MetricsStorage) *metricsService {
	return &metricsService{
		conf:    conf,
		storage: store,
	}
}

func (s *metricsService) Stop() {
	err := s.storage.LoadMetricsFromFile(s.conf.FileStoragePath)
	if err != nil {
		fmt.Printf("saving metrics error: %s", err.Error())
	}
	close(s.exit)
	log.Println("Server stopped.")
}

func (s *metricsService) Run() {
	logger.InitLogger(s.conf.LogLevel)
	s.exit = make(chan struct{})

	err := s.storage.LoadMetricsFromFile(s.conf.FileStoragePath)
	if err != nil {
		log.Printf("load metrics error: %s", err.Error())
	}
	go s.storage.StartSavingMetricsToFile(s.conf.FileStoragePath, s.conf.StoreInterval, s.exit)

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
