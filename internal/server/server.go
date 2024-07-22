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
	conf *config.HTTPServerConfig
}

func NewMetricsService(
	conf *config.HTTPServerConfig,
) *metricsService {
	return &metricsService{
		conf: conf,
	}
}

func (s *metricsService) Stop(cancel context.CancelFunc) {
	cancel()
	log.Println("Server stopped.")
}

func (s *metricsService) Run(ctx context.Context) {
	storage := storage.NewMetricsStorage(ctx, *s.conf)
	defer storage.Close()
	logger.InitLogger(s.conf.LogLevel)

	r := chi.NewRouter()
	r.Use(middlewares.LoggerMiddleware)
	contentTypesForCompress := "application/json text/html"
	r.Use(middlewares.GetGzipMiddleware(contentTypesForCompress))

	handlers.RegisterMetricsHandlers(r, storage)

	log.Printf("Server http://%s is running. Press Ctrl+C to stop.", s.conf.Addr)
	err := http.ListenAndServe(s.conf.Addr, r)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
}
