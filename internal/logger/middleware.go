package logger

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		size   int
		status int
	}

	logResponseWriter struct {
		responseData *responseData
		http.ResponseWriter
	}
)

func (lw *logResponseWriter) Write(body []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(body)
	if err != nil {
		return 0, err
	}
	lw.responseData.size += size
	return size, err
}

func (lw *logResponseWriter) WriteHeader(statusCode int) {
	lw.ResponseWriter.WriteHeader(statusCode)
	lw.responseData.status = statusCode
}

func LoggerMiddleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var (
			respData = responseData{0, 0}
			lw       = logResponseWriter{responseData: &respData, ResponseWriter: w}
			duration time.Duration
		)
		start := time.Now()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to read request body", http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		h.ServeHTTP(&lw, r)
		duration = time.Since(start)
		Log.Info("HTTP request",
			zap.String("URI", r.URL.Path),
			zap.String("method", r.Method),
			zap.Duration("duration", duration),
			zap.Int("statusCode", lw.responseData.status),
			zap.Int("size", lw.responseData.size),
			zap.ByteString("body", bodyBytes),
		)
	}
	return http.HandlerFunc(fn)
}
