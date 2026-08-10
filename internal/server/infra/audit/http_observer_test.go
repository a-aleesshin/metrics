package audit

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	serveraudit "github.com/a-aleesshin/metrics/internal/server/audit"
)

func fastRetryObserver(url string) *HTTPObserver {
	return &HTTPObserver{
		url: url,
		client: newRetryHTTPClient(
			&http.Client{Timeout: defaultTimeout},
			[]time.Duration{time.Millisecond, time.Millisecond, time.Millisecond},
		),
	}
}

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

func TestHTTPObserver_Notify_ServerErrorRetriedThenReturnsError(t *testing.T) {
	var hits atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	observer := fastRetryObserver(server.URL)

	err := observer.Notify(context.Background(), serveraudit.Event{TS: 1})
	if err == nil {
		t.Fatal("expected error on 500 response")
	}

	// 1 попытка + 3 ретрая
	if got := hits.Load(); got != 4 {
		t.Fatalf("expected 4 attempts on 5xx, got %d", got)
	}
}

func TestHTTPObserver_Notify_RecoversAfterServerErrors(t *testing.T) {
	var hits atomic.Int64
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observer := fastRetryObserver(server.URL)

	event := serveraudit.Event{TS: 42, Metrics: []string{"Alloc"}, IPAddress: "10.0.0.1"}

	if err := observer.Notify(context.Background(), event); err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}

	if got := hits.Load(); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}

	// тело запроса должно пересоздаваться на каждую попытку
	var got serveraudit.Event
	if err := json.Unmarshal(gotBody, &got); err != nil {
		t.Fatalf("body of retried request is not valid json: %v", err)
	}

	if got.TS != event.TS {
		t.Fatalf("expected event %+v, got %+v", event, got)
	}
}

func TestHTTPObserver_Notify_ClientErrorNotRetried(t *testing.T) {
	var hits atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	observer := fastRetryObserver(server.URL)

	err := observer.Notify(context.Background(), serveraudit.Event{TS: 1})
	if err == nil {
		t.Fatal("expected error on 400 response")
	}

	if got := hits.Load(); got != 1 {
		t.Fatalf("expected no retries on 4xx, got %d attempts", got)
	}
}

func TestHTTPObserver_Notify_UnreachableServerReturnsError(t *testing.T) {
	observer := fastRetryObserver("http://127.0.0.1:1")

	err := observer.Notify(context.Background(), serveraudit.Event{TS: 1})
	if err == nil {
		t.Fatal("expected error when server is unreachable")
	}
}
