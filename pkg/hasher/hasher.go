package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func HashSumToString(data []byte, key string) string {
	if key != "" {
		h := hmac.New(sha256.New, []byte(key))
		h.Write(data)
		dst := h.Sum(nil)
		return hex.EncodeToString(dst)
	}
	return ""
}
