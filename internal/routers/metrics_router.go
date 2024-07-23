package routers

import "github.com/go-chi/chi/v5"

type metricsRouter struct {
	router
}

func NewMetricsRouter(mux *chi.Mux) *metricsRouter {
	return &metricsRouter{router: router{mux: mux}}
}

func (mr *metricsRouter) RegisterHandlers() {
	
}
