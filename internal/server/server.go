package server

import (
	"fmt"
	"net/http"

	"github.com/eac0de/getmetrics/internal/handlers"
	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

func (s *Server) Run() {
	metricsStorage := storage.NewMetricsStorage()
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", handlers.GetMetricsSummaryHTMLHandler(metricsStorage))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetricHandler(metricsStorage))
	r.Get("/value/{metricType}/{metricName}", handlers.GetMetricHandler(metricsStorage))
	fmt.Printf("Server started on http://%s\n", s.addr)
	err := http.ListenAndServe(s.addr, r)
	if err != nil {
		fmt.Println(err.Error())
	}
}
