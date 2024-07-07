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

type MetricsServer struct {
	conf    *config.HTTPServerConfig
	exit    chan struct{}
	storage *storage.MetricsStorage
}

func NewMetricsServer(
	conf *config.HTTPServerConfig,
	store *storage.MetricsStorage) *MetricsServer {
	return &MetricsServer{
		conf:    conf,
		storage: store,
	}
}

func (s *MetricsServer) Stop() {
	err := s.storage.LoadMetricsFromFile(s.conf.FileStoragePath)
	if err != nil {
		fmt.Printf("saving metrics error: %s", err.Error())
	}
	close(s.exit)
	log.Println("Server stopped.")
}

func (s *MetricsServer) Run() {
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

	metricsHandlerService := handlers.NewMetricsHandlerService(s.storage)

	r.Get("/", metricsHandlerService.ShowMetricsSummaryHandler())

	r.Post("/update/{metricType}/{metricName}/{metricValue}", metricsHandlerService.UpdateMetricHandler())
	r.Post("/update/", metricsHandlerService.UpdateMetricJSONHandler())

	r.Get("/value/{metricType}/{metricName}", metricsHandlerService.GetMetricHandler())
	r.Post("/value/", metricsHandlerService.GetMetricJSONHandler())
	log.Printf("Server http://%s is running. Press Ctrl+C to stop.", s.conf.Addr)
	err = http.ListenAndServe(s.conf.Addr, r)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
}
