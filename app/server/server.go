package server

import (
	"context"
	"log"
	"net/http"

	"github.com/eac0de/getmetrics/internal/config"
	"github.com/eac0de/getmetrics/internal/handlers"
	"github.com/eac0de/getmetrics/internal/logger"
	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/eac0de/getmetrics/pkg/middlewares"
	"github.com/go-chi/chi/v5"
)

type metricsService struct {
	conf    *config.HTTPServerConfig
	storage *storage.MetricsStorage
	chi.Mux
}

func NewMetricsService(
	conf *config.HTTPServerConfig,
) *metricsService {
	logger.InitLogger(s.conf.LogLevel)
	storage := storage.NewMetricsStorage(ctx, *s.conf)
	s.storage = storage
	r := chi.NewRouter()
	r.Use(middlewares.LoggerMiddleware)
	contentTypesForCompress := "application/json text/html"
	r.Use(middlewares.GetGzipMiddleware(contentTypesForCompress))

	metricsHandlerService := handlers.NewMetricsHandlerService(storage)

	r.Get("/", metricsHandlerService.ShowMetricsSummaryHandler())

	r.Post("/update/{metricType}/{metricName}/{metricValue}", metricsHandlerService.UpdateMetricHandler())
	r.Post("/update/", metricsHandlerService.UpdateMetricJSONHandler())
	r.Post("/updates/", metricsHandlerService.UpdateManyMetricsJSONHandler())

	r.Get("/value/{metricType}/{metricName}", metricsHandlerService.GetMetricHandler())
	r.Post("/value/", metricsHandlerService.GetMetricJSONHandler())

	r.Get("/ping", metricsHandlerService.PingHandler())
	return &metricsService{
		conf: conf,
	}
}

func (s *metricsService) Stop(cancel context.CancelFunc) {
	if s.storage != nil {
		s.storage.Close()
	}
	cancel()
	log.Println("Server stopped")
}

func (s *metricsService) Run(ctx context.Context) {
	mux := http.NewServeMux()
	log.Printf("Server http://%s is running. Press Ctrl+C to stop", s.conf.Addr)
	err := mux.ListenAndServe(s.conf.Addr, r)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
}
