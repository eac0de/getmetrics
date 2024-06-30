package compressor

import (
	"net/http"
	"strings"
)

func GetGzipMiddleware(contentTypes string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ow := w
			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")

			if supportsGzip {
				cw := newCompressWriter(w, contentTypes)
				ow = cw
				defer cw.Close()
			}
			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := newCompressReader(r.Body)
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
