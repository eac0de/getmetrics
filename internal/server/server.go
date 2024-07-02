package server

import (
	"log"
	"net/http"
	"time"

	"github.com/eac0de/getmetrics/internal/compressor"
	"github.com/eac0de/getmetrics/internal/handlers"
	"github.com/eac0de/getmetrics/internal/logger"
	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

type MetricsServer struct {
	addr            string
	logLevel        string
	fileStoragePath string
	restore         bool
	storeInterval   time.Duration
	exit            chan struct{}
	storage         *storage.MetricsStorage
}

func NewMetricsServer(
	addr string,
	logLevel string,
	fileStoragePath string,
	restore bool,
	storeInterval time.Duration,
) *MetricsServer {
	return &MetricsServer{
		addr:            addr,
		logLevel:        logLevel,
		fileStoragePath: fileStoragePath,
		restore:         restore,
		storeInterval:   storeInterval,
		storage:         storage.NewMetricsStorage(),
	}
}

func (s *MetricsServer) Stop() {
	SaveMetricsToFile(s.fileStoragePath, s.storage)
	close(s.exit)
	log.Println("Server stopped.")
}

func (s *MetricsServer) Run() {
	s.exit = make(chan struct{})

	logger.InitLogger(s.logLevel)
	err := LoadMetricsFromFile(s.fileStoragePath, s.storage)
	if err != nil {
		log.Printf("load metrics error: %s", err.Error())
	}
	if s.fileStoragePath != "" {
		go SaveMetricsToFileGorutine(s, s.storage)
	}
	r := chi.NewRouter()
	r.Use(logger.LoggerMiddleware)
	contentTypesForCompress := "application/json text/html"
	r.Use(compressor.GetGzipMiddleware(contentTypesForCompress))

	r.Get("/", handlers.ShowMetricsSummaryHandler(s.storage))

	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetricHandler(s.storage))
	r.Post("/update/", handlers.UpdateMetricJSONHandler(s.storage))

	r.Get("/value/{metricType}/{metricName}", handlers.GetMetricHandler(s.storage))
	r.Post("/value/", handlers.GetMetricJSONHandler(s.storage))
	log.Printf("Server http://%s is running. Press Ctrl+C to stop.", s.addr)
	err = http.ListenAndServe(s.addr, r)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
	<-s.exit
}
