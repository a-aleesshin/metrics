package grpccreds

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateSelfSigned_ProducesUsableCredentials(t *testing.T) {
	certPEM, keyPEM, err := GenerateSelfSigned("localhost", "127.0.0.1")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if _, err := ServerFromPEM(certPEM, keyPEM); err != nil {
		t.Fatalf("server creds from generated pair: %v", err)
	}

	if _, err := ClientFromPEM(certPEM); err != nil {
		t.Fatalf("client creds from generated cert: %v", err)
	}
}

func TestFileLoaders(t *testing.T) {
	certPEM, keyPEM, err := GenerateSelfSigned("localhost")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	if _, err := Server(certPath, keyPath); err != nil {
		t.Fatalf("server from files: %v", err)
	}
	if _, err := Client(certPath); err != nil {
		t.Fatalf("client from file: %v", err)
	}

	if _, err := Client(filepath.Join(dir, "missing.pem")); err == nil {
		t.Fatal("expected error for missing ca file")
	}
	if _, err := ClientFromPEM([]byte("not a pem")); err == nil {
		t.Fatal("expected error for invalid pem")
	}
}
