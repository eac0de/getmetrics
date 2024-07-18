package server

import (
	"context"
	"fmt"
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
	storage storage.MetricsStorer
}

func NewMetricsService(
	conf *config.HTTPServerConfig,
) *metricsService {
	return &metricsService{
		conf:    conf,
		storage: nil,
	}
}

func (s *metricsService) Stop(cancel context.CancelFunc) {
	cancel()
	log.Println("Server stopped.")
}

func (s *metricsService) Run(ctx context.Context) {
	logger.InitLogger(s.conf.LogLevel)

	r := chi.NewRouter()
	r.Use(middlewares.LoggerMiddleware)
	contentTypesForCompress := "application/json text/html"
	r.Use(middlewares.GetGzipMiddleware(contentTypesForCompress))

	db, err := storage.NewDatabaseSQL(ctx, s.conf.DatabaseDSN)
	if err != nil {
		fmt.Printf("Database initialization ended with an error: %v\n", err.Error())
		store := storage.NewMetricsStorage(s.conf.FileStoragePath)
		go store.StartSavingMetricsToFile(ctx, s.conf.FileStoragePath, s.conf.StoreInterval)
		s.storage = store
	} else {
		defer db.Close()
		handlers.RegisterDatabaseHandlers(r, db)
		s.storage = db
	}

	handlers.RegisterMetricsHandlers(r, s.storage)

	log.Printf("Server http://%s is running. Press Ctrl+C to stop.", s.conf.Addr)
	err = http.ListenAndServe(s.conf.Addr, r)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
}
