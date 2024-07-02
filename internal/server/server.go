package server

import (
	"fmt"
	"net/http"

	"github.com/eac0de/getmetrics/internal/compressor"
	"github.com/eac0de/getmetrics/internal/handlers"
	"github.com/eac0de/getmetrics/internal/logger"
	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type MetricsServer struct {
	addr string
}

func NewMetricsServer(addr string) *MetricsServer {
	return &MetricsServer{
		addr: addr,
	}
}

func (s *MetricsServer) Run(logLevel string) {
	logger.InitLogger(logLevel)
	metricsStorage := storage.NewMetricsStorage()
	LoadMetricsFromFile("save_metrics.json", metricsStorage)
	r := chi.NewRouter()
	r.Use(logger.LoggerMiddleware)
	contentTypesForCompress := "application/json text/html"
	r.Use(compressor.GetGzipMiddleware(contentTypesForCompress))
	
	r.Get("/", handlers.ShowMetricsSummaryHandler(metricsStorage))

	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetricHandler(metricsStorage))
	r.Post("/update/", handlers.UpdateMetricJSONHandler(metricsStorage))

	r.Get("/value/{metricType}/{metricName}", handlers.GetMetricHandler(metricsStorage))
	r.Post("/value/", handlers.GetMetricJSONHandler(metricsStorage))

	logger.Log.Info("Running server", zap.String("address", fmt.Sprintf("http://%s", s.addr)))
	err := http.ListenAndServe(s.addr, r)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
}
