// Package compressor предоставляет функции для сжатия и разжатия данных с использованием алгоритма Gzip.
//
// Этот пакет реализует обертки для http.ResponseWriter и io.ReadCloser, что позволяет легко
// добавлять сжатие данных для HTTP-ответов и разжатие данных для HTTP-запросов.
// Основные функции пакета включают:
// - Создание нового сжатого HTTP-ответа с автоматическим определением типов контента.
// - Чтение и разжатие Gzip-данных из HTTP-запросов.
// - Упрощенное сжатие произвольных данных в формате Gzip.
package compressor

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// compressWriter реализует http.ResponseWriter и используется для сжатия и отправки данных.
// Он проверяет, поддерживает ли указанный тип контента сжатие, и устанавливает необходимые заголовки.
type compressWriter struct {
	w            http.ResponseWriter
	zw           *gzip.Writer
	contentTypes string
	gzipEnabled  bool
}

// NewCompressWriter создает новый экземпляр compressWriter.
//
// Принимает на вход http.ResponseWriter и строку с поддерживаемыми типами контента для сжатия.
func NewCompressWriter(w http.ResponseWriter, contentTypes string) *compressWriter {
	return &compressWriter{
		w:            w,
		contentTypes: contentTypes,
	}
}

// Write записывает данные в ответ, применяя сжатие, если это возможно.
// Если сжатие включено, данные будут записаны в gzip.Writer, иначе они будут записаны напрямую.
func (cw *compressWriter) Write(p []byte) (int, error) {
	if !cw.gzipEnabled {
		contentType := strings.Split(cw.w.Header().Get("Content-Type"), ";")[0]
		isTypeForCompress := strings.Contains(cw.contentTypes, contentType)
		if isTypeForCompress {
			cw.w.Header().Set("Content-Encoding", "gzip")
			cw.zw = gzip.NewWriter(cw.w)
			cw.gzipEnabled = true
		}
	}

	if cw.gzipEnabled {
		return cw.zw.Write(p)
	}
	return cw.w.Write(p)
}

// Header возвращает заголовки ответа.
func (cw *compressWriter) Header() http.Header {
	return cw.w.Header()
}

// WriteHeader отправляет код статуса и устанавливает заголовки ответа.
func (cw *compressWriter) WriteHeader(statusCode int) {
	if !cw.gzipEnabled {
		contentType := strings.Split(cw.w.Header().Get("Content-Type"), ";")[0]
		isTypeForCompress := strings.Contains(cw.contentTypes, contentType)
		if isTypeForCompress {
			cw.w.Header().Set("Content-Encoding", "gzip")
			cw.zw = gzip.NewWriter(cw.w)
			cw.gzipEnabled = true
		}
	}
	cw.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer, если сжатие было включено.
func (cw *compressWriter) Close() error {
	if cw.gzipEnabled {
		return cw.zw.Close()
	}
	return nil
}

// compressReader реализует io.ReadCloser и используется для чтения сжатых данных.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader создает новый экземпляр compressReader.
//
// Принимает на вход io.ReadCloser и инициализирует gzip.Reader для разжатия данных.
func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read читает разжатые данные из gzip.Reader.
func (cr compressReader) Read(p []byte) (n int, err error) {
	return cr.zr.Read(p)
}

// Close закрывает как исходный ReadCloser, так и gzip.Reader.
func (cr *compressReader) Close() error {
	if err := cr.r.Close(); err != nil {
		return err
	}
	return cr.zr.Close()
}

// GzipData сжимает данные в формате Gzip и возвращает их в виде байтового среза.
//
// Принимает байтовый массив и возвращает сжатые данные или ошибку, если сжатие не удалось.
func GzipData(p []byte) ([]byte, error) {
	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}
	_, err = w.Write(p)
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %v", err)
	}
	return b.Bytes(), nil
}
