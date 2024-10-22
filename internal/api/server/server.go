package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	Addr           string
	PrivateKeyPath string
}

func New(addr string, privateKeyPath string) *Server {
	return &Server{Addr: addr, PrivateKeyPath: privateKeyPath}
}

func (s *Server) Run(router chi.Router) {
	err := http.ListenAndServeTLS(s.Addr, "server.crt", s.PrivateKeyPath, router)
	if err != nil {
		log.Fatal(err.Error())
	}
}
