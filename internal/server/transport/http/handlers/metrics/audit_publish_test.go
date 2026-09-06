package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
	"github.com/go-chi/chi/v5"
)

type auditPublisherSpy struct {
	called      bool
	metricNames []string
	ipAddress   string
}

func (s *auditPublisherSpy) Publish(_ context.Context, metricNames []string, ipAddress string) {
	s.called = true
	s.metricNames = metricNames
	s.ipAddress = ipAddress
}

func TestUpdateHandler_PublishesAuditOnSuccess(t *testing.T) {
	auditSpy := &auditPublisherSpy{}
	handler := NewUpdateHandler(&updateMetricUseCaseSpy{}, auditSpy)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handler.Update)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/123.45", nil)
	req.RemoteAddr = "192.168.0.42:54321"
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if !auditSpy.called {
		t.Fatal("expected audit event to be published")
	}

	if len(auditSpy.metricNames) != 1 || auditSpy.metricNames[0] != "Alloc" {
		t.Fatalf("expected metrics [Alloc], got %v", auditSpy.metricNames)
	}

	if auditSpy.ipAddress != "192.168.0.42" {
		t.Fatalf("expected ip 192.168.0.42, got %s", auditSpy.ipAddress)
	}
}

func TestUpdateHandler_NoAuditOnError(t *testing.T) {
	auditSpy := &auditPublisherSpy{}
	handler := NewUpdateHandler(&updateMetricUseCaseSpy{err: metric.ErrInvalidMetricValue}, auditSpy)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handler.Update)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/invalid", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if auditSpy.called {
		t.Fatal("expected no audit event on use case error")
	}
}

func TestUpdateJSONHandler_PublishesAuditOnSuccess(t *testing.T) {
	auditSpy := &auditPublisherSpy{}
	handler := NewUpdateJSONHandler(&updateMetricUseCaseSpy{}, auditSpy)

	body := `{"id":"Alloc","type":"gauge","value":123.45}`
	req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "10.0.0.7:1234"
	rec := httptest.NewRecorder()

	handler.UpdateJSON(rec, req)

	if !auditSpy.called {
		t.Fatal("expected audit event to be published")
	}

	if len(auditSpy.metricNames) != 1 || auditSpy.metricNames[0] != "Alloc" {
		t.Fatalf("expected metrics [Alloc], got %v", auditSpy.metricNames)
	}

	if auditSpy.ipAddress != "10.0.0.7" {
		t.Fatalf("expected ip 10.0.0.7, got %s", auditSpy.ipAddress)
	}
}

func TestUpdatesHandler_PublishesAuditWithAllNames(t *testing.T) {
	auditSpy := &auditPublisherSpy{}
	handler := NewUpdatesHandler(&updatesMetricsUseCaseSpy{}, auditSpy)

	body := `[{"id":"Alloc","type":"gauge","value":1.5},{"id":"PollCount","type":"counter","delta":3}]`
	req := httptest.NewRequest(http.MethodPost, "/updates", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "172.16.0.9:8888"
	rec := httptest.NewRecorder()

	handler.Updates(rec, req)

	if !auditSpy.called {
		t.Fatal("expected audit event to be published")
	}

	want := []string{"Alloc", "PollCount"}
	if len(auditSpy.metricNames) != len(want) {
		t.Fatalf("expected metrics %v, got %v", want, auditSpy.metricNames)
	}

	for i := range want {
		if auditSpy.metricNames[i] != want[i] {
			t.Fatalf("expected metrics %v, got %v", want, auditSpy.metricNames)
		}
	}

	if auditSpy.ipAddress != "172.16.0.9" {
		t.Fatalf("expected ip 172.16.0.9, got %s", auditSpy.ipAddress)
	}
}
