package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestLogger_LogsRequestAndResponse(t *testing.T) {
	// Arrange
	core, recorded := observer.New(zapcore.InfoLevel)
	log := zap.New(core)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})

	h := RequestLogger(log)(next)

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/Alloc", nil)
	rec := httptest.NewRecorder()

	// Act
	h.ServeHTTP(rec, req)

	// Assert
	entries := recorded.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	fields := entries[0].ContextMap()
	if fields["uri"] != "/value/gauge/Alloc" {
		t.Fatalf("expected uri %q, got %v", "/value/gauge/Alloc", fields["uri"])
	}
	if fields["method"] != http.MethodGet {
		t.Fatalf("expected method %q, got %v", http.MethodGet, fields["method"])
	}
	if fields["status"] != int64(http.StatusOK) {
		t.Fatalf("expected status %d, got %v", http.StatusOK, fields["status"])
	}
	if fields["size"] != int64(5) {
		t.Fatalf("expected size %d, got %v", 5, fields["size"])
	}
	if _, ok := fields["duration"]; !ok {
		t.Fatal("expected duration field")
	}
}

func TestRequestLogger_LogsStatusFromWriteHeader(t *testing.T) {
	// Arrange
	core, recorded := observer.New(zapcore.InfoLevel)
	log := zap.New(core)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	h := RequestLogger(log)(next)

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()

	// Act
	h.ServeHTTP(rec, req)

	// Assert
	entries := recorded.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	fields := entries[0].ContextMap()
	if fields["status"] != int64(http.StatusNotFound) {
		t.Fatalf("expected status %d, got %v", http.StatusNotFound, fields["status"])
	}
	if fields["size"] != int64(0) {
		t.Fatalf("expected size 0, got %v", fields["size"])
	}
}
