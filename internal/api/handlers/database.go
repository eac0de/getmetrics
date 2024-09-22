package handlers

import (
	"context"
	"net/http"

	"github.com/eac0de/getmetrics/pkg/errors"
)

// IDatabase определяет интерфейс для работы с базой данных.
type IDatabase interface {
	PingContext(ctx context.Context) error
}

// DatabaseHandlers предоставляет обработчики для работы с базой данных.
type DatabaseHandlers struct {
	// Database хранит экземпляр базы данных, реализующий интерфейс IDatabase.
	Database IDatabase
}

// NewDatabaseHandlers создает новый экземпляр DatabaseHandlers.
//
// Принимает объект, реализующий интерфейс IDatabase.
func NewDatabaseHandlers(database IDatabase) *DatabaseHandlers {
	return &DatabaseHandlers{
		Database: database,
	}
}

// PingHandler возвращает HTTP-обработчик для проверки подключения к базе данных.
//
// Если база данных не инициализирована, возвращает ошибку 500.
// В противном случае проверяет подключение и возвращает статус 200.
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
