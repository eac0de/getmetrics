// Package errors предоставляет структуры и функции для обработки ошибок с указанием HTTP-статусов.
//
// Этот пакет реализует механизм создания ошибок с сопутствующими сообщениями и кодами статуса,
// что позволяет упростить обработку ошибок в HTTP-приложениях.
// Основные функции пакета включают:
// - Создание ошибки с указанием сообщения и кода статуса.
// - Получение сообщения и кода статуса из ошибки.
package errors

import (
	"errors"
	"net/http"
)

// ErrorWithHTTPStatus представляет ошибку с сопутствующим HTTP-статусом.
// Включает код статуса, сообщение и оригинальную ошибку.
type ErrorWithHTTPStatus struct {
	statusCode int    // Код HTTP-статуса
	msg        string // Сообщение ошибки
	err        error  // Оригинальная ошибка
}

// Error возвращает сообщение ошибки.
func (ewhs *ErrorWithHTTPStatus) Error() string {
	return ewhs.msg
}

// Unwrap возвращает оригинальную ошибку, если она есть.
func (ewhs *ErrorWithHTTPStatus) Unwrap() error {
	return ewhs.err
}

// NewErrorWithHTTPStatus создает новую ошибку с указанным сообщением и HTTP-статусом.
//
// Принимает оригинальную ошибку, сообщение и код статуса, возвращает указатель на ErrorWithHTTPStatus.
func NewErrorWithHTTPStatus(err error, msg string, statusCode int) error {
	return &ErrorWithHTTPStatus{
		statusCode: statusCode,
		err:        err,
		msg:        msg,
	}
}

// GetMessageAndStatusCode извлекает сообщение и HTTP-статус из переданной ошибки.
//
// Если ошибка является ErrorWithHTTPStatus, возвращает её сообщение и код статуса.
// В противном случае возвращает сообщение ошибки и код 500 (внутренняя ошибка сервера).
func GetMessageAndStatusCode(err error) (string, int) {
	var ewhs *ErrorWithHTTPStatus
	if errors.As(err, &ewhs) {
		return ewhs.msg, ewhs.statusCode
	}
	return err.Error(), http.StatusInternalServerError
}
