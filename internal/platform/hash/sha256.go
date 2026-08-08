package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

func SumSHA256(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	_, _ = h.Write(data)

	return hex.EncodeToString(h.Sum(nil))
}

func VerifySHA256(data []byte, key string, expected string) bool {
	actual := SumSHA256(data, key)

	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}
