// Package middlewares предоставляет промежуточные обработчики для различных целей,
// включая логирование запросов и ответов в HTTP-сервере.
//
// Этот пакет реализует функциональность для логирования информации о HTTP-запросах
// и ответах, включая метод запроса, статус ответа, путь URL, продолжительность обработки
// и размер ответа.
package middlewares

import (
	"log"
	"net/http"
	"time"
)

type (
	// responseData содержит информацию о размере ответа и статусе.
	responseData struct {
		size   int // Размер ответа в байтах.
		status int // Статус-код ответа.
	}

	// logResponseWriter оборачивает http.ResponseWriter для логирования ответов.
	logResponseWriter struct {
		responseData *responseData // Указатель на responseData.
		http.ResponseWriter
	}
)

// Write записывает тело ответа и обновляет размер ответа.
func (lw *logResponseWriter) Write(body []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(body)
	if err != nil {
		return 0, err
	}
	lw.responseData.size += size // Увеличиваем общий размер ответа.
	return size, err
}

// WriteHeader записывает статус-код ответа и обновляет статус.
func (lw *logResponseWriter) WriteHeader(statusCode int) {
	lw.ResponseWriter.WriteHeader(statusCode)
	lw.responseData.status = statusCode // Устанавливаем статус-код.
}

// LoggerMiddleware возвращает промежуточный обработчик для логирования запросов и ответов.
//
// Логирует метод запроса, статус-код ответа, путь URL, продолжительность обработки
// и размер ответа. Используйте этот middleware для отслеживания производительности
// и анализа трафика вашего HTTP-сервера.
func LoggerMiddleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var (
			respData = responseData{0, 0} // Инициализируем данные о ответе.
			lw       = logResponseWriter{responseData: &respData, ResponseWriter: w}
			duration time.Duration
		)
		start := time.Now()          // Запоминаем время начала обработки.
		h.ServeHTTP(&lw, r)          // Обрабатываем запрос.
		duration = time.Since(start) // Вычисляем продолжительность обработки.
		log.Printf("%s %v %s %s %v bytes", r.Method, lw.responseData.status, r.URL.Path, duration, lw.responseData.size)
	}
	return http.HandlerFunc(fn)
}
