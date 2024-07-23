package routers

import "github.com/go-chi/chi/v5"

type router struct {
	mux *chi.Mux
}

func NewRouter(mux *chi.Mux) *router {
	return &router{mux: mux}
}

func (r *router) RegisterHandlers() {

}
