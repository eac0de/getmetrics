package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/eac0de/getmetrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	db, err := sql.Open("sqlite", "video.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		panic(err)
	}
	// ...
}

type databaseHandlerService struct {
	Database *storage.DatabaseSQL
}

func NewDatabaseHandlerService(database *storage.DatabaseSQL) *databaseHandlerService {
	return &databaseHandlerService{
		Database: database,
	}
}

func RegisterDatabaseHandlers(r chi.Router, database *storage.DatabaseSQL) {
	metricsHandlerService := NewDatabaseHandlerService(database)

	r.Get("/ping", metricsHandlerService.PingHandler())
}

func (dhs *databaseHandlerService) PingHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := dhs.Database.Ping(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
