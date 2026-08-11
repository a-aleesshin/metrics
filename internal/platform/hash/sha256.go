// Package hash предоставляет вычисление и проверку подписи HMAC-SHA256.
package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// SumSHA256 возвращает HMAC-SHA256 от data с ключом key в hex-кодировке.
func SumSHA256(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	_, _ = h.Write(data)

	return hex.EncodeToString(h.Sum(nil))
}

// VerifySHA256 проверяет подпись expected для data с ключом key;
// сравнение выполняется за константное время.
func VerifySHA256(data []byte, key string, expected string) bool {
	actual := SumSHA256(data, key)

	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}
