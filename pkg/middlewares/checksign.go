// Package middlewares предоставляет промежуточные обработчики для проверки подписи запросов.
//
// Этот пакет реализует проверку HMAC-подписи для обеспечения целостности данных, отправляемых в HTTP-запросах.
// Основные функции пакета включают:
// - Проверку подписи HMAC для запросов с использованием заданного секретного ключа.
package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

// GetCheckSignMiddleware возвращает промежуточный обработчик для проверки подписи HMAC запросов.
//
// Принимает секретный ключ в виде строки. Если ключ пуст, возвращается промежуточный обработчик,
// который не выполняет проверку подписи. В противном случае, проверяет, соответствует ли
// HMAC-подпись из заголовка HashSHA256 телу запроса.
//
// Если подпись не соответствует, отправляет ответ с кодом ошибки 400 (Bad Request).
// В случае успешной проверки передает управление следующему обработчику.
func GetCheckSignMiddleware(secretKey string) func(http.Handler) http.Handler {
	if secretKey == "" {
		return func(next http.Handler) http.Handler {
			fn := func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			}
			return http.HandlerFunc(fn)
		}
	}
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			sign := r.Header.Get("HashSHA256")
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Unable to read body", http.StatusInternalServerError)
				return
			}
			h := hmac.New(sha256.New, []byte(secretKey))
			h.Write(bodyBytes)
			dst := h.Sum(nil)
			hash := hex.EncodeToString(dst)
			if hash != sign {
				http.Error(w, "Signature does not match data", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
