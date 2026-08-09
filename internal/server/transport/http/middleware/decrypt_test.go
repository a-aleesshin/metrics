package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	platformcrypto "github.com/a-aleesshin/metrics/internal/platform/crypto"
)

func echoHandler(t *testing.T) (http.Handler, *bytes.Buffer) {
	t.Helper()

	var got bytes.Buffer

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := got.ReadFrom(r.Body); err != nil {
			t.Errorf("read body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}), &got
}

func TestDecryptRequest_DecryptsEncryptedBody(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	plaintext := []byte(`[{"id":"Alloc","type":"gauge","value":1.5}]`)

	encrypted, err := platformcrypto.Encrypt(&key.PublicKey, plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	handler, got := echoHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(encrypted))
	req.Header.Set(platformcrypto.HeaderEncrypted, platformcrypto.HeaderEncryptedValue)
	rec := httptest.NewRecorder()

	DecryptRequest(key)(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if !bytes.Equal(got.Bytes(), plaintext) {
		t.Fatalf("expected decrypted body %q, got %q", plaintext, got.Bytes())
	}
}

func TestDecryptRequest_PassesPlainRequestsThrough(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	handler, got := echoHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader([]byte("plain")))
	rec := httptest.NewRecorder()

	DecryptRequest(key)(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if got.String() != "plain" {
		t.Fatalf("expected body passed through, got %q", got.String())
	}
}

func TestDecryptRequest_GarbageBodyRejected(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	handler, _ := echoHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader([]byte("garbage")))
	req.Header.Set(platformcrypto.HeaderEncrypted, platformcrypto.HeaderEncryptedValue)
	rec := httptest.NewRecorder()

	DecryptRequest(key)(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDecryptRequest_EncryptedRequestWithoutServerKeyRejected(t *testing.T) {
	handler, _ := echoHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/updates", io.NopCloser(bytes.NewReader([]byte("data"))))
	req.Header.Set(platformcrypto.HeaderEncrypted, platformcrypto.HeaderEncryptedValue)
	rec := httptest.NewRecorder()

	DecryptRequest(nil)(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when server has no key, got %d", rec.Code)
	}
}
