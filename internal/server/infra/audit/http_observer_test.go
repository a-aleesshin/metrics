package audit

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	serveraudit "github.com/a-aleesshin/metrics/internal/server/audit"
)

func TestHTTPObserver_Notify_SendsPostWithJSON(t *testing.T) {
	var gotMethod, gotContentType string
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observer := NewHTTPObserver(server.URL, nil)

	event := serveraudit.Event{TS: 12345678, Metrics: []string{"Alloc"}, IPAddress: "192.168.0.42"}

	if err := observer.Notify(context.Background(), event); err != nil {
		t.Fatalf("notify: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %s", gotMethod)
	}

	if gotContentType != "application/json" {
		t.Fatalf("expected application/json, got %s", gotContentType)
	}

	var got serveraudit.Event
	if err := json.Unmarshal(gotBody, &got); err != nil {
		t.Fatalf("body is not valid json: %v", err)
	}

	if got.TS != event.TS || got.IPAddress != event.IPAddress {
		t.Fatalf("expected event %+v, got %+v", event, got)
	}
}

func TestHTTPObserver_Notify_ServerErrorReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	observer := NewHTTPObserver(server.URL, nil)

	err := observer.Notify(context.Background(), serveraudit.Event{TS: 1})
	if err == nil {
		t.Fatal("expected error on 500 response")
	}
}

func TestHTTPObserver_Notify_UnreachableServerReturnsError(t *testing.T) {
	observer := NewHTTPObserver("http://127.0.0.1:1", nil)

	err := observer.Notify(context.Background(), serveraudit.Event{TS: 1})
	if err == nil {
		t.Fatal("expected error when server is unreachable")
	}
}
