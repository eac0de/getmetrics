package server

import (
	"fmt"
	"net/http"

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
	r := chi.NewRouter()
	r.Use(logger.LoggerMiddleware)
	r.Get("/", handlers.ShowMetricsSummaryHandler(metricsStorage))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetricHandler(metricsStorage))
	r.Get("/value/{metricType}/{metricName}", handlers.GetMetricHandler(metricsStorage))
	logger.Log.Info("Running server", zap.String("address", fmt.Sprintf("http://%s", s.addr)))
	err := http.ListenAndServe(s.addr, r)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
}
