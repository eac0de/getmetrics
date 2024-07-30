package routers

import (
	"github.com/eac0de/getmetrics/internal/handlers"
	"github.com/eac0de/getmetrics/pkg/middlewares"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	*chi.Mux
}

func NewRouter(handlerService *handlers.MetricsHandlerService, secretKey string) *Router {
	mux := chi.NewRouter()
	router := &Router{mux}
	router.Use(middlewares.LoggerMiddleware)
	router.Use(middlewares.GetCheckSignMiddleware(secretKey))
	contentTypesForCompress := "application/json text/html"
	router.Use(middlewares.GetGzipMiddleware(contentTypesForCompress))
	router.Get("/", handlerService.ShowMetricsSummaryHandler())

	router.Post("/update/{metricType}/{metricName}/{metricValue}", handlerService.UpdateMetricHandler())
	router.Post("/update/", handlerService.UpdateMetricJSONHandler())
	router.Post("/updates/", handlerService.UpdateManyMetricsJSONHandler())

	router.Get("/value/{metricType}/{metricName}", handlerService.GetMetricHandler())
	router.Post("/value/", handlerService.GetMetricJSONHandler())

	router.Get("/ping", handlerService.PingHandler())
	return router
}
