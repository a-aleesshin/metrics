package httpadapter

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"

	platformcrypto "github.com/a-aleesshin/metrics/internal/platform/crypto"
)

// EncryptingClient — декоратор HTTPClient, шифрующий тело запроса публичным
// ключом сервера (гибридная схема RSA-OAEP + AES-256-GCM) и помечающий запрос
// заголовком X-Content-Encrypted. Запросы без тела проходят без изменений.
type EncryptingClient struct {
	client    HTTPClient
	publicKey *rsa.PublicKey
}

// NewEncryptingClient создаёт шифрующий декоратор поверх client.
func NewEncryptingClient(client HTTPClient, publicKey *rsa.PublicKey) *EncryptingClient {
	return &EncryptingClient{
		client:    client,
		publicKey: publicKey,
	}
}

// Do шифрует тело запроса и передаёт запрос вложенному клиенту.
// GetBody переустанавливается, чтобы ретраи повторяли зашифрованное тело.
func (c *EncryptingClient) Do(request *http.Request) (*http.Response, error) {
	if request.Body == nil {
		return c.client.Do(request)
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)
	}

	if err := request.Body.Close(); err != nil {
		return nil, fmt.Errorf("close request body: %w", err)
	}

	encrypted, err := platformcrypto.Encrypt(c.publicKey, body)
	if err != nil {
		return nil, fmt.Errorf("encrypt request body: %w", err)
	}

	request.Body = io.NopCloser(bytes.NewReader(encrypted))
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(encrypted)), nil
	}
	request.ContentLength = int64(len(encrypted))
	request.Header.Set(platformcrypto.HeaderEncrypted, platformcrypto.HeaderEncryptedValue)

	return c.client.Do(request)
}
