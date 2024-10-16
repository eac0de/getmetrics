package middlewares

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Тестирует LoggerMiddleware для успешного запроса.
func TestLoggerMiddleware_Success(t *testing.T) {
	// Создаем фейковый логгер для захвата логов.
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)

	// Создаем фейковый обработчик, который будет возвращать 200 OK.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Создаем тестовый HTTP-запрос.
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Оборачиваем обработчик в наш middleware.
	middleware := LoggerMiddleware(handler)

	// Выполняем запрос через middleware.
	middleware.ServeHTTP(rec, req)

	// Проверяем, что ответ имеет правильный статус.
	if status := rec.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	// Проверяем, что тело ответа записано корректно.
	expectedBody := "test response"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, rec.Body.String())
	}

	// Проверяем содержимое логов.
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "GET 200 /test") {
		t.Errorf("Log output does not contain expected data: %s", logOutput)
	}
	if !strings.Contains(logOutput, "bytes") {
		t.Errorf("Log output does not contain byte size: %s", logOutput)
	}
}

// Тестирует LoggerMiddleware для запроса с ошибкой (например, 404).
func TestLoggerMiddleware_NotFound(t *testing.T) {
	// Создаем фейковый логгер для захвата логов.
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)

	// Создаем фейковый обработчик, который будет возвращать 404 Not Found.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	// Создаем тестовый HTTP-запрос.
	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	rec := httptest.NewRecorder()

	// Оборачиваем обработчик в наш middleware.
	middleware := LoggerMiddleware(handler)

	// Выполняем запрос через middleware.
	middleware.ServeHTTP(rec, req)

	// Проверяем, что ответ имеет правильный статус.
	if status := rec.Code; status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, status)
	}

	// Проверяем содержимое логов.
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "GET 404 /not-found") {
		t.Errorf("Log output does not contain expected data: %s", logOutput)
	}
}

// Тестирует LoggerMiddleware для медленного запроса (проверка продолжительности).
func TestLoggerMiddleware_SlowRequest(t *testing.T) {
	// Создаем фейковый логгер для захвата логов.
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)

	// Создаем фейковый обработчик, который будет обрабатывать запрос с задержкой.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("slow response"))
	})

	// Создаем тестовый HTTP-запрос.
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()

	// Оборачиваем обработчик в наш middleware.
	middleware := LoggerMiddleware(handler)

	// Выполняем запрос через middleware.
	middleware.ServeHTTP(rec, req)

	// Проверяем, что ответ имеет правильный статус.
	if status := rec.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	// Проверяем содержимое логов.
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "GET 200 /slow") {
		t.Errorf("Log output does not contain expected data: %s", logOutput)
	}
	if !strings.Contains(logOutput, "bytes") {
		t.Errorf("Log output does not contain byte size: %s", logOutput)
	}
}
