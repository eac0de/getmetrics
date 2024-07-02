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
	}
}

func (s *MetricsServer) Stop() {
	close(s.exit)
	log.Println("Server stopped.")
}

func (s *MetricsServer) Run() {
	s.exit = make(chan struct{})

	logger.InitLogger(s.logLevel)
	metricsStorage := storage.NewMetricsStorage()
	LoadMetricsFromFile(s.fileStoragePath, metricsStorage)
	if s.fileStoragePath != "" {
		go SaveMetricsToFileGorutine(s, metricsStorage)
	}
	r := chi.NewRouter()
	r.Use(logger.LoggerMiddleware)
	contentTypesForCompress := "application/json text/html"
	r.Use(compressor.GetGzipMiddleware(contentTypesForCompress))

	r.Get("/", handlers.ShowMetricsSummaryHandler(metricsStorage))

	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetricHandler(metricsStorage))
	r.Post("/update/", handlers.UpdateMetricJSONHandler(metricsStorage))

	r.Get("/value/{metricType}/{metricName}", handlers.GetMetricHandler(metricsStorage))
	r.Post("/value/", handlers.GetMetricJSONHandler(metricsStorage))
	log.Printf("Server http://%s is running. Press Ctrl+C to stop.", s.addr)
	err := http.ListenAndServe(s.addr, r)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
	<-s.exit
	SaveMetricsToFile(s.fileStoragePath, metricsStorage)
}
