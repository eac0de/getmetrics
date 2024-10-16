package middlewares

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Вспомогательная функция для генерации HMAC подписи
func generateHMAC(body string, secretKey string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(body))
	return hex.EncodeToString(h.Sum(nil))
}

// Тест успешной проверки подписи
func TestCheckSignMiddleware_ValidSignature(t *testing.T) {
	secretKey := "mysecretkey"
	body := "test body"

	// Генерируем правильную HMAC подпись
	signature := generateHMAC(body, secretKey)

	// Создаем фейковый обработчик, который возвращает статус 200
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Создаем тестовый HTTP-запрос с правильной подписью
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set("HashSHA256", signature)

	// Создаем тестовый Recorder для записи ответа
	rec := httptest.NewRecorder()

	// Оборачиваем обработчик в HMAC middleware
	middleware := GetCheckSignMiddleware(secretKey)(handler)

	// Выполняем запрос через middleware
	middleware.ServeHTTP(rec, req)

	// Проверяем, что запрос прошел успешно
	if status := rec.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}
}

// Тест неуспешной проверки подписи
func TestCheckSignMiddleware_InvalidSignature(t *testing.T) {
	secretKey := "mysecretkey"
	body := "test body"
	invalidSignature := "invalidsignature"

	// Создаем фейковый обработчик, который возвращает статус 200
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Создаем тестовый HTTP-запрос с неправильной подписью
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set("HashSHA256", invalidSignature)

	// Создаем тестовый Recorder для записи ответа
	rec := httptest.NewRecorder()

	// Оборачиваем обработчик в HMAC middleware
	middleware := GetCheckSignMiddleware(secretKey)(handler)

	// Выполняем запрос через middleware
	middleware.ServeHTTP(rec, req)

	// Проверяем, что запрос отклонен с кодом 400
	if status := rec.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}

	// Проверяем сообщение об ошибке
	expectedBody := "Signature does not match data\n"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, rec.Body.String())
	}
}

// Тест пропуска проверки, если секретный ключ не задан
func TestCheckSignMiddleware_NoSecretKey(t *testing.T) {
	body := "test body"

	// Создаем фейковый обработчик, который возвращает статус 200
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Создаем тестовый HTTP-запрос без подписи
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))

	// Создаем тестовый Recorder для записи ответа
	rec := httptest.NewRecorder()

	// Оборачиваем обработчик в HMAC middleware без ключа
	middleware := GetCheckSignMiddleware("")(handler)

	// Выполняем запрос через middleware
	middleware.ServeHTTP(rec, req)

	// Проверяем, что запрос прошел успешно
	if status := rec.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}
}
