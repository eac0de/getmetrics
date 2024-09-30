// Package hasher предоставляет функции для вычисления HMAC-хешей.
//
// Этот пакет реализует HMAC с использованием SHA-256 для обеспечения целостности данных.
// Основные функции пакета включают:
// - Вычисление HMAC-хеша для заданных данных с использованием ключа.
package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// HashSumToString вычисляет HMAC-хеш для заданных данных с использованием указанного ключа.
//
// Принимает данные в виде массива байт и ключ в виде строки.
// Возвращает строковое представление хеша в шестнадцатеричном формате.
// Если ключ пуст, возвращает пустую строку.
func HashSumToString(data []byte, key string) string {
	if key != "" {
		h := hmac.New(sha256.New, []byte(key))
		h.Write(data)
		dst := h.Sum(nil)
		return hex.EncodeToString(dst)
	}
	return ""
}
