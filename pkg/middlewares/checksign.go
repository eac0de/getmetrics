package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

func GetCheckSignMiddleware(secretKey string) func(http.Handler) http.Handler {
	if secretKey == "" {
		return func(next http.Handler) http.Handler {
			fn := func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			}
			return http.HandlerFunc(fn)
		}
	}
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			sign := r.Header.Get("HashSHA256")
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Unable to read body", http.StatusInternalServerError)
				return
			}
			h := hmac.New(sha256.New, []byte(secretKey))
			h.Write(bodyBytes)
			dst := h.Sum(nil)
			hash := hex.EncodeToString(dst)
			if hash != sign {
				http.Error(w, "Signature does not match data", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
