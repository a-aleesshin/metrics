package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func gzipBytes(t *testing.T, s string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte(s))

	if err != nil {
		t.Fatalf("gzip write failed: %v", err)
	}

	if err := gw.Close(); err != nil {
		t.Fatalf("gzip close failed: %v", err)
	}

	return buf.Bytes()
}

func gunzipString(t *testing.T, b []byte) string {
	t.Helper()

	gr, err := gzip.NewReader(bytes.NewReader(b))

	if err != nil {
		t.Fatalf("gzip reader failed: %v", err)
	}

	defer gr.Close()

	out, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("gzip read failed: %v", err)
	}

	return string(out)
}

func TestDecompressRequest(t *testing.T) {
	tests := []struct {
		name            string
		contentEncoding string
		body            []byte
		wantStatus      int
		wantBody        string
	}{
		{
			name:            "decompresses gzip body",
			contentEncoding: "gzip",
			body:            gzipBytes(t, `{"id":"Alloc"}`),
			wantStatus:      http.StatusOK,
			wantBody:        `{"id":"Alloc"}`,
		},
		{
			name:            "passes plain body without gzip",
			contentEncoding: "",
			body:            []byte(`{"id":"Alloc"}`),
			wantStatus:      http.StatusOK,
			wantBody:        `{"id":"Alloc"}`,
		},
		{
			name:            "returns 400 on invalid gzip",
			contentEncoding: "gzip",
			body:            []byte("not-gzip"),
			wantStatus:      http.StatusBadRequest,
			wantBody:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(raw)
			})

			h := DecompressRequest(next)

			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(tt.body))
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}

			rec := httptest.NewRecorder()

			// Act
			h.ServeHTTP(rec, req)

			// Assert
			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if tt.wantStatus == http.StatusOK {
				got := rec.Body.String()
				if got != tt.wantBody {
					t.Fatalf("expected body %q, got %q", tt.wantBody, got)
				}
			}
		})
	}
}

func TestCompressResponse(t *testing.T) {
	tests := []struct {
		name           string
		acceptEncoding string
		contentType    string
		handlerBody    string
		wantGzip       bool
		wantBody       string
	}{
		{
			name:           "compresses application/json when client accepts gzip",
			acceptEncoding: "gzip",
			contentType:    "application/json",
			handlerBody:    `{"ok":true}`,
			wantGzip:       true,
			wantBody:       `{"ok":true}`,
		},
		{
			name:           "compresses text/html when client accepts gzip",
			acceptEncoding: "gzip",
			contentType:    "text/html; charset=utf-8",
			handlerBody:    "<h1>ok</h1>",
			wantGzip:       true,
			wantBody:       "<h1>ok</h1>",
		},
		{
			name:           "does not compress unsupported content type",
			acceptEncoding: "gzip",
			contentType:    "text/plain",
			handlerBody:    "plain",
			wantGzip:       false,
			wantBody:       "plain",
		},
		{
			name:           "does not compress when client does not accept gzip",
			acceptEncoding: "",
			contentType:    "application/json",
			handlerBody:    `{"ok":true}`,
			wantGzip:       false,
			wantBody:       `{"ok":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.handlerBody))
			})

			h := CompressResponse(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			rec := httptest.NewRecorder()

			// Act
			h.ServeHTTP(rec, req)

			// Assert
			if rec.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", rec.Code)
			}

			gotEncoding := rec.Header().Get("Content-Encoding")
			if tt.wantGzip {
				if gotEncoding != "gzip" {
					t.Fatalf("expected Content-Encoding gzip, got %q", gotEncoding)
				}
				if !strings.Contains(rec.Header().Get("Vary"), "Accept-Encoding") {
					t.Fatalf("expected Vary to contain Accept-Encoding, got %q", rec.Header().Get("Vary"))
				}

				gotBody := gunzipString(t, rec.Body.Bytes())
				if gotBody != tt.wantBody {
					t.Fatalf("expected body %q, got %q", tt.wantBody, gotBody)
				}
			} else {
				if gotEncoding != "" {
					t.Fatalf("expected no Content-Encoding, got %q", gotEncoding)
				}
				if rec.Body.String() != tt.wantBody {
					t.Fatalf("expected body %q, got %q", tt.wantBody, rec.Body.String())
				}
			}
		})
	}
}
