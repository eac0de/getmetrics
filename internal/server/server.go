package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/eac0de/getmetrics/internal/compressor"
	"github.com/eac0de/getmetrics/internal/config"
	"github.com/eac0de/getmetrics/internal/handlers"
	"github.com/eac0de/getmetrics/internal/logger"
	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

type MetricsServer struct {
	conf        *config.HTTPServerConfig
	exit        chan struct{}
	storage     storage.MetricsStorer
	fileService *MetricsFileService
}

func NewMetricsServer(
	conf *config.HTTPServerConfig,
	store storage.MetricsStorer) *MetricsServer {
	fileService := MetricsFileService{
		filename:    conf.FileStoragePath,
		metricStore: store,
	}
	return &MetricsServer{
		conf:        conf,
		storage:     store,
		fileService: &fileService,
	}
}

func (s *MetricsServer) Stop() {
	err := s.fileService.SaveMetrics()
	if err != nil {
		fmt.Printf("saving metrics error: %s", err.Error())
	}
	close(s.exit)
	log.Println("Server stopped.")
}

func (s *MetricsServer) Run() {
	logger.InitLogger(s.conf.LogLevel)
	s.exit = make(chan struct{})

	err := s.fileService.LoadMetrics()
	if err != nil {
		log.Printf("load metrics error: %s", err.Error())
	}
	go s.fileService.SaveMetricsToFileGorutine(s)

	r := chi.NewRouter()
	r.Use(logger.LoggerMiddleware)
	contentTypesForCompress := "application/json text/html"
	r.Use(compressor.GetGzipMiddleware(contentTypesForCompress))

	r.Get("/", handlers.ShowMetricsSummaryHandler(s.storage))

	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetricHandler(s.storage))
	r.Post("/update/", handlers.UpdateMetricJSONHandler(s.storage))

	r.Get("/value/{metricType}/{metricName}", handlers.GetMetricHandler(s.storage))
	r.Post("/value/", handlers.GetMetricJSONHandler(s.storage))
	log.Printf("Server http://%s is running. Press Ctrl+C to stop.", s.conf.Addr)
	err = http.ListenAndServe(s.conf.Addr, r)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
}
