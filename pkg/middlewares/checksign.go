package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
)

func GetCheckSignMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}
			sign := r.Header.Get("HashSHA256")
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Unable to read body", http.StatusInternalServerError)
				return
			}
			h := hmac.New(sha256.New, []byte(key))
			h.Write(bodyBytes)
			dst := h.Sum(nil)
			hash := hex.EncodeToString(dst)
			if hash != sign {
				fmt.Println(key)
				fmt.Println(hash)
				fmt.Println(sign)
				http.Error(w, "Signature does not match data", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
