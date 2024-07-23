package routers

import (
	"github.com/eac0de/getmetrics/internal/handlers"
	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/eac0de/getmetrics/pkg/middlewares"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	*chi.Mux
}

func NewRouter() *Router {
	mux := chi.NewRouter()
	router := &Router{mux}
	return router
}

func (r *Router) AddMiddlewares() {
	r.Use(middlewares.LoggerMiddleware)
	contentTypesForCompress := "application/json text/html"
	r.Use(middlewares.GetGzipMiddleware(contentTypesForCompress))
}

func (r *Router) RegisterMetricsHandlers(storage storage.MetricsStorer) {
	handlerService := handlers.NewMetricsHandlerService(storage)
	r.Get("/", handlerService.ShowMetricsSummaryHandler())

	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlerService.UpdateMetricHandler())
	r.Post("/update/", handlerService.UpdateMetricJSONHandler())
	r.Post("/updates/", handlerService.UpdateManyMetricsJSONHandler())

	r.Get("/value/{metricType}/{metricName}", handlerService.GetMetricHandler())
	r.Post("/value/", handlerService.GetMetricJSONHandler())

	r.Get("/ping", handlerService.PingHandler())
}
