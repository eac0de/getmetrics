// Package middlewares предоставляет промежуточные обработчики для различных целей, включая
// сжатие и декомпрессию данных в HTTP-запросах и ответах.
//
// Этот пакет реализует поддержку Gzip-сжатия для HTTP-трафика. Основные функции пакета включают:
// - Обработку запросов с Gzip-сжатием, а также сжатие ответов для экономии полосы пропускания.
package middlewares

import (
	"net/http"
	"strings"

	"github.com/eac0de/getmetrics/pkg/compressor"
)

// GetGzipMiddleware возвращает промежуточный обработчик для Gzip-сжатия.
//
// Принимает строку, содержащую типы контента, которые должны быть сжаты.
// Если клиент поддерживает Gzip (указан в заголовке Accept-Encoding),
// ответ будет сжат с использованием Gzip. Если запрос содержит Gzip-сжатие
// (указан в заголовке Content-Encoding), тело запроса будет декомпрессировано.
//
// После выполнения обработки запрос будет передан следующему обработчику.
func GetGzipMiddleware(contentTypes string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ow := w
			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")

			if supportsGzip {
				cw := compressor.NewCompressWriter(w, contentTypes)
				ow = cw
				defer cw.Close()
			}
			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := compressor.NewCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer cr.Close()
			}
			next.ServeHTTP(ow, r)
		}
		return http.HandlerFunc(fn)
	}
}
