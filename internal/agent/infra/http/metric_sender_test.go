package httpadapter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
	dto2 "github.com/a-aleesshin/metrics/internal/agent/infra/dto"
)

func TestMetricSender_Send_GaugeJSON(t *testing.T) {
	// Arrange
	var gotMethod, gotPath, gotContentType string
	var gotBody dto2.MetricsSend

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")

		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sender := NewMetricSender(ts.URL, ts.Client())

	// Act
	err := sender.Send(dto.MetricDTO{
		Type:  "gauge",
		Name:  "Alloc",
		Value: "123.45",
	})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("expected method POST, got %s", gotMethod)
	}
	if gotPath != "/update" {
		t.Fatalf("expected path /update, got %s", gotPath)
	}
	if !strings.Contains(gotContentType, "application/json") {
		t.Fatalf("expected content-type application/json, got %q", gotContentType)
	}
	if gotBody.ID != "Alloc" || gotBody.MType != "gauge" {
		t.Fatalf("unexpected body id/type: %+v", gotBody)
	}
	if gotBody.Value == nil || *gotBody.Value != 123.45 {
		t.Fatalf("expected value 123.45, got %+v", gotBody.Value)
	}
	if gotBody.Delta != nil {
		t.Fatalf("expected delta nil, got %+v", gotBody.Delta)
	}
}

func TestMetricSender_Send_CounterJSON(t *testing.T) {
	// Arrange
	var gotBody dto2.MetricsSend

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sender := NewMetricSender(ts.URL, ts.Client())

	// Act
	err := sender.Send(dto.MetricDTO{
		Type:  "counter",
		Name:  "PollCount",
		Value: "7",
	})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody.ID != "PollCount" || gotBody.MType != "counter" {
		t.Fatalf("unexpected body id/type: %+v", gotBody)
	}
	if gotBody.Delta == nil || *gotBody.Delta != 7 {
		t.Fatalf("expected delta 7, got %+v", gotBody.Delta)
	}
	if gotBody.Value != nil {
		t.Fatalf("expected value nil, got %+v", gotBody.Value)
	}
}

func TestMetricSender_Send_InvalidGaugeValue(t *testing.T) {
	// Arrange
	sender := NewMetricSender("http://localhost:8080", http.DefaultClient)

	// Act
	err := sender.Send(dto.MetricDTO{
		Type:  "gauge",
		Name:  "Alloc",
		Value: "not-a-float",
	})

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMetricSender_Send_InvalidCounterValue(t *testing.T) {
	// Arrange
	sender := NewMetricSender("http://localhost:8080", http.DefaultClient)

	// Act
	err := sender.Send(dto.MetricDTO{
		Type:  "counter",
		Name:  "PollCount",
		Value: "not-an-int",
	})

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMetricSender_Send_UnsupportedType(t *testing.T) {
	// Arrange
	sender := NewMetricSender("http://localhost:8080", http.DefaultClient)

	// Act
	err := sender.Send(dto.MetricDTO{
		Type:  "histogram",
		Name:  "Any",
		Value: "1",
	})

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMetricSender_Send_Non200Status(t *testing.T) {
	// Arrange
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	sender := NewMetricSender(ts.URL, ts.Client())

	// Act
	err := sender.Send(dto.MetricDTO{
		Type:  "gauge",
		Name:  "Alloc",
		Value: "1.23",
	})

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMetricSender_Send_NormalizesAddressWithoutScheme(t *testing.T) {
	// Arrange
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	addrWithoutScheme := strings.TrimPrefix(ts.URL, "http://")
	sender := NewMetricSender(addrWithoutScheme, ts.Client())

	// Act
	err := sender.Send(dto.MetricDTO{
		Type:  "gauge",
		Name:  "Alloc",
		Value: "10.5",
	})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected server to be called")
	}
}
