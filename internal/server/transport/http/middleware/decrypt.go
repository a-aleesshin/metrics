package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	platformcrypto "github.com/a-aleesshin/metrics/internal/platform/crypto"
)

// DecryptRequest — middleware, расшифровывающее тело запросов с заголовком
// X-Content-Encrypted приватным ключом privateKey. При nil ключе middleware
// отключается; запросы без заголовка проходят без изменений. Запрос
// с заголовком при выключенном шифровании или с некорректным телом — 400.
func DecryptRequest(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if privateKey == nil {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get(platformcrypto.HeaderEncrypted) != "" {
					http.Error(w, "encryption is not enabled on server", http.StatusBadRequest)
					return
				}

				next.ServeHTTP(w, r)
			})
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(platformcrypto.HeaderEncrypted) == "" {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "read request body", http.StatusBadRequest)
				return
			}
			_ = r.Body.Close()

			decrypted, err := platformcrypto.Decrypt(privateKey, body)
			if err != nil {
				http.Error(w, "decrypt request body", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			r.ContentLength = int64(len(decrypted))
			r.Header.Del(platformcrypto.HeaderEncrypted)

			next.ServeHTTP(w, r)
		})
	}
}
