package server

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/eac0de/getmetrics/internal/config"
	"github.com/eac0de/getmetrics/internal/handlers"
	"github.com/eac0de/getmetrics/internal/logger"
	"github.com/eac0de/getmetrics/internal/routers"
	"github.com/eac0de/getmetrics/internal/storage"
)

type MetrciServerApp struct {
	conf    *config.HTTPServerConfig
	storage *storage.MetricsStorage
	router  routers.Router
}

func NewMetrciServerApp(
	ctx context.Context,
	conf *config.HTTPServerConfig,
) *MetrciServerApp {
	logger.InitLogger(conf.LogLevel)
	storage := storage.NewMetricsStorage(ctx, *conf)
	metricsHandlerService := handlers.NewMetricsHandlerService(storage)
	router := routers.NewRouter(metricsHandlerService)
	return &MetrciServerApp{
		conf:    conf,
		storage: storage,
		router:  *router,
	}
}

func (s *MetrciServerApp) Stop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	<-ctx.Done()
	if s.storage != nil {
		s.storage.Close()
	}
	log.Println("Server stopped")
}

func (s *MetrciServerApp) Run() {
	log.Printf("Server http://%s is running. Press Ctrl+C to stop", s.conf.Addr)
	err := http.ListenAndServe(s.conf.Addr, s.router)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

}
