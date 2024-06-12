package server

import (
	"fmt"
	"github.com/eac0de/getmetrics/internal/handlers"
	"github.com/eac0de/getmetrics/internal/storage"
	"net/http"
)

type Server struct {
	addr string
}

func NewServer(host string, port string) *Server {
	addr := host + ":" + port
	return &Server{
		addr: addr,
	}
}

func (s *Server) Run() {
	metricsStorage := storage.NewMetricsStorage()
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handlers.UpdateMetricHandler(metricsStorage))
	fmt.Printf("Server started on %s\n", s.addr)
	err := http.ListenAndServe(s.addr, mux)
	if err != nil {
		fmt.Println(err.Error())
	}
}
