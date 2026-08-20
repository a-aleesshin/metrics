package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func generateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	return key
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := generateKey(t)

	payloads := [][]byte{
		[]byte("small"),
		bytes.Repeat([]byte("metrics"), 100_000), // ~700 КБ — больше блока RSA
		{},
	}

	for _, payload := range payloads {
		encrypted, err := Encrypt(&key.PublicKey, payload)
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}

		if bytes.Contains(encrypted, payload) && len(payload) > 0 {
			t.Fatal("ciphertext contains plaintext")
		}

		decrypted, err := Decrypt(key, encrypted)
		if err != nil {
			t.Fatalf("decrypt: %v", err)
		}

		if !bytes.Equal(decrypted, payload) {
			t.Fatalf("roundtrip mismatch: want %d bytes, got %d", len(payload), len(decrypted))
		}
	}
}

func TestDecrypt_WrongKeyFails(t *testing.T) {
	key := generateKey(t)
	otherKey := generateKey(t)

	encrypted, err := Encrypt(&key.PublicKey, []byte("secret"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	if _, err := Decrypt(otherKey, encrypted); err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}
}

func TestDecrypt_TamperedDataFails(t *testing.T) {
	key := generateKey(t)

	encrypted, err := Encrypt(&key.PublicKey, []byte("secret"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	encrypted[len(encrypted)-1] ^= 0xFF

	if _, err := Decrypt(key, encrypted); err == nil {
		t.Fatal("expected error when decrypting tampered message")
	}
}

func TestLoadKeys_PEMRoundTrip(t *testing.T) {
	key := generateKey(t)
	dir := t.TempDir()

	privatePath := filepath.Join(dir, "private.pem")
	privateBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err := os.WriteFile(privatePath, privateBytes, 0o600); err != nil {
		t.Fatalf("write private key: %v", err)
	}

	publicPath := filepath.Join(dir, "public.pem")
	publicDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	publicBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})
	if err := os.WriteFile(publicPath, publicBytes, 0o644); err != nil {
		t.Fatalf("write public key: %v", err)
	}

	loadedPrivate, err := LoadPrivateKey(privatePath)
	if err != nil {
		t.Fatalf("load private key: %v", err)
	}

	loadedPublic, err := LoadPublicKey(publicPath)
	if err != nil {
		t.Fatalf("load public key: %v", err)
	}

	encrypted, err := Encrypt(loadedPublic, []byte("payload"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	decrypted, err := Decrypt(loadedPrivate, encrypted)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}

	if string(decrypted) != "payload" {
		t.Fatalf("unexpected roundtrip result: %q", decrypted)
	}
}
