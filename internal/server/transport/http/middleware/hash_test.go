package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-aleesshin/metrics/internal/platform/hash"
)

func TestWithHashSHA256(t *testing.T) {
	tests := []struct {
		name            string
		key             string
		body            []byte
		header          string
		wantStatusCode  int
		wantCalled      bool
		wantResponseSig bool
	}{
		{
			name:            "disabled when key is empty",
			key:             "",
			body:            []byte(`{"ok":true}`),
			wantStatusCode:  http.StatusOK,
			wantCalled:      true,
			wantResponseSig: false,
		},
		{
			name:            "valid hash",
			key:             "secret",
			body:            []byte(`{"ok":true}`),
			header:          hash.SumSHA256([]byte(`{"ok":true}`), "secret"),
			wantStatusCode:  http.StatusOK,
			wantCalled:      true,
			wantResponseSig: true,
		},
		{
			name:           "missing hash",
			key:            "secret",
			body:           []byte(`{"ok":true}`),
			wantStatusCode: http.StatusBadRequest,
			wantCalled:     false,
		},
		{
			name:           "invalid hash",
			key:            "secret",
			body:           []byte(`{"ok":true}`),
			header:         "invalid",
			wantStatusCode: http.StatusBadRequest,
			wantCalled:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			})

			handler := WithHashSHA256(tt.key)(next)
			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(tt.body))
			if tt.header != "" {
				req.Header.Set(HashSHA256Header, tt.header)
			}
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			if rec.Code != tt.wantStatusCode {
				t.Fatalf("expected status %d, got %d", tt.wantStatusCode, rec.Code)
			}

			if called != tt.wantCalled {
				t.Fatalf("expected called=%v, got %v", tt.wantCalled, called)
			}

			gotHash := rec.Header().Get(HashSHA256Header)
			if tt.wantResponseSig {
				if gotHash == "" {
					t.Fatal("expected response hash header")
				}
				if !hash.VerifySHA256(rec.Body.Bytes(), tt.key, gotHash) {
					t.Fatalf("invalid response hash %q", gotHash)
				}
			} else if gotHash != "" {
				t.Fatalf("expected empty response hash, got %q", gotHash)
			}
		})
	}
}
