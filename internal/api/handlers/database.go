package handlers

import (
	"context"
	"net/http"

	"github.com/eac0de/getmetrics/pkg/errors"
)

type IDatabase interface {
	PingContext(ctx context.Context) error
}

type DatabaseHandlers struct {
	Database IDatabase
}

func NewDatabaseHandlers(database IDatabase) *DatabaseHandlers {
	return &DatabaseHandlers{
		Database: database,
	}
}

func (h *DatabaseHandlers) PingHandler() func(http.ResponseWriter, *http.Request) {
	if h.Database == nil {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Database not init", http.StatusInternalServerError)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.Database.PingContext(r.Context()); err != nil {
			msg, statusCode := errors.GetMessageAndStatusCode(err)
			http.Error(w, msg, statusCode)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
