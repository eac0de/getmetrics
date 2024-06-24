package logger

import (
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
		h.ServeHTTP(&lw, r)
		duration = time.Since(start)
		Log.Info("HTTP request",
			zap.String("URI", r.URL.Path),
			zap.String("method", r.Method),
			zap.Duration("duration", duration),
			zap.Int("statusCode", lw.responseData.status),
			zap.Int("size", lw.responseData.size),
		)
	}
	return http.HandlerFunc(fn)
}
