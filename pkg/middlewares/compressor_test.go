package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Тест для сжатого ответа с Gzip
func TestGetGzipMiddleware_ResponseWithGzip(t *testing.T) {
	// Создаем фейковый обработчик, который возвращает простой текст.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, World!"))
	})

	// Создаем тестовый HTTP-запрос с Accept-Encoding: gzip
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	// Создаем тестовый Recorder для записи ответа
	rec := httptest.NewRecorder()

	// Оборачиваем обработчик в Gzip middleware
	middleware := GetGzipMiddleware("text/plain")(handler)

	// Выполняем запрос через middleware
	middleware.ServeHTTP(rec, req)

	// Проверяем, что Content-Encoding заголовок установлен в gzip
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding gzip, got %s", rec.Header().Get("Content-Encoding"))
	}

	// Проверяем, что ответ действительно сжат
	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal("Failed to create gzip reader:", err)
	}
	defer gr.Close()

	// Декомпрессируем данные и проверяем их содержимое
	body, err := io.ReadAll(gr)
	if err != nil {
		t.Fatal("Failed to read gzipped response:", err)
	}

	expectedBody := "Hello, World!"
	if string(body) != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, body)
	}
}

// Тест для декомпрессии запроса с Gzip
func TestGetGzipMiddleware_RequestWithGzip(t *testing.T) {
	// Создаем сжатое тело запроса
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write([]byte("test body"))
	if err != nil {
		t.Fatal("Failed to write gzip body:", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal("Failed to close gzip writer:", err)
	}

	// Создаем фейковый обработчик, который читает тело запроса и проверяет его
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal("Failed to read request body:", err)
		}
		expectedBody := "test body"
		if string(body) != expectedBody {
			t.Errorf("Expected body %q, got %q", expectedBody, body)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Создаем тестовый HTTP-запрос с Content-Encoding: gzip
	req := httptest.NewRequest(http.MethodPost, "/test", &buf)
	req.Header.Set("Content-Encoding", "gzip")

	// Создаем тестовый Recorder для записи ответа
	rec := httptest.NewRecorder()

	// Оборачиваем обработчик в Gzip middleware
	middleware := GetGzipMiddleware("text/plain")(handler)

	// Выполняем запрос через middleware
	middleware.ServeHTTP(rec, req)

	// Проверяем, что ответ имеет правильный статус
	if status := rec.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}
}

// Тест для обычного запроса без Gzip
func TestGetGzipMiddleware_NoGzip(t *testing.T) {
	// Создаем фейковый обработчик, который возвращает простой текст.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	// Создаем тестовый HTTP-запрос без Gzip
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Создаем тестовый Recorder для записи ответа
	rec := httptest.NewRecorder()

	// Оборачиваем обработчик в Gzip middleware
	middleware := GetGzipMiddleware("text/plain")(handler)

	// Выполняем запрос через middleware
	middleware.ServeHTTP(rec, req)

	// Проверяем, что Content-Encoding заголовок не установлен
	if rec.Header().Get("Content-Encoding") != "" {
		t.Errorf("Expected no Content-Encoding, got %s", rec.Header().Get("Content-Encoding"))
	}

	// Проверяем содержимое ответа
	expectedBody := "Hello, World!"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, rec.Body.String())
	}
}
