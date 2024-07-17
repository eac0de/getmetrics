package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/eac0de/getmetrics/internal/config"
	"github.com/eac0de/getmetrics/internal/handlers"
	"github.com/eac0de/getmetrics/internal/logger"
	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/eac0de/getmetrics/pkg/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
	var err error
	var db *storage.DatabaseSQL
	for waitTime := 1; waitTime <= 5; waitTime += 2 {
		db, err = storage.NewDatabaseSQL(ctx, s.conf.DatabaseDSN)
		if err != nil {
			fmt.Println(err.Error())
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				fmt.Println(pgErr.Code)
				if pgerrcode.IsConnectionException(pgErr.Code) {
					fmt.Printf("Database connection error. New attempt in %v sec\n", waitTime)
					time.Sleep(time.Duration(waitTime) * time.Second)
					continue
				}
			}
		}
		break
	}

	if err != nil {
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
