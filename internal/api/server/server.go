package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	Addr string
}

func New(addr string) *Server {
	return &Server{Addr: addr}
}

func (s *Server) Run(router chi.Router) {
	err := http.ListenAndServe(s.Addr, router)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func (s *Server) RunTLS(router chi.Router, privateKeyPath string) {
	err := http.ListenAndServeTLS(s.Addr, "server.crt", privateKeyPath, router)
	if err != nil {
		log.Fatal(err.Error())
	}
}
