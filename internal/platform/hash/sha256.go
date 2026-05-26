package hash

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

func SumSHA256(data []byte, key string) string {
	h := sha256.New()

	_, _ = h.Write(data)
	_, _ = h.Write([]byte(key))

	return hex.EncodeToString(h.Sum(nil))
}

func VerifySHA256(data []byte, key string, expected string) bool {
	actual := SumSHA256(data, key)

	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}
