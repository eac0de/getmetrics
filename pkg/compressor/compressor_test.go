package compressor

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestWriteHeader проверяет работу WriteHeader с сжатием.
func TestWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	contentTypes := "text/plain"

	cw := NewCompressWriter(rec, contentTypes)

	// Устанавливаем заголовки перед вызовом WriteHeader.
	cw.Header().Set("Content-Type", "text/plain")
	cw.WriteHeader(http.StatusOK)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Проверяем, что Content-Encoding установлен.
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding gzip, got %s", rec.Header().Get("Content-Encoding"))
	}

	// Проверяем, что запись данных работает после WriteHeader.
	data := []byte("hello world")
	_, err := cw.Write(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	err = cw.Close()
	if err != nil {
		t.Fatalf("Unexpected error during close: %v", err)
	}

	// Проверяем, что данные действительно сжаты.
	r, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer r.Close()

	uncompressed, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Failed to decompress response: %v", err)
	}

	if string(uncompressed) != string(data) {
		t.Errorf("Expected %q, got %q", data, uncompressed)
	}
}

// TestCompressReaderClose проверяет корректное закрытие compressReader.
func TestCompressReaderClose(t *testing.T) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	data := []byte("close test")
	if _, err := w.Write(data); err != nil {
		t.Fatalf("Failed to write gzip data: %v", err)
	}
	w.Close()

	r := io.NopCloser(&buf)
	cr, err := NewCompressReader(r)
	if err != nil {
		t.Fatalf("Failed to create compressReader: %v", err)
	}

	err = cr.Close()
	if err != nil {
		t.Fatalf("Failed to close compressReader: %v", err)
	}
}

// TestGzipDataError проверяет обработку ошибок в GzipData.
func TestGzipDataError(t *testing.T) {
	_, err := GzipData([]byte("test"))
	if err != nil {
		t.Error("Expected an nil, got error")
	}
}

// TestNewCompressReaderError проверяет обработку ошибок при создании compressReader.
func TestNewCompressReaderError(t *testing.T) {
	// Передаем некорректный gzip поток.
	r := io.NopCloser(strings.NewReader("invalid gzip data"))
	_, err := NewCompressReader(r)
	if err == nil {
		t.Error("Expected error for invalid gzip data, got nil")
	}
}
