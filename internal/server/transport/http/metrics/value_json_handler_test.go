package metrics

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	applicationerror "github.com/a-aleesshin/metrics/internal/server/application/error"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
	"github.com/go-chi/chi/v5"
)

func TestHandler_ValueJSON(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		contentType    string
		useCaseErr     error
		useCaseResult  string
		wantStatusCode int
		wantCalled     bool
		wantCommand    usecase.ValueMetricCommand
		wantResp       *Metrics
	}{
		{
			name:           "success gauge returns JSON",
			body:           `{"id":"Alloc","type":"gauge"}`,
			contentType:    "application/json",
			useCaseResult:  "123.45",
			wantStatusCode: http.StatusOK,
			wantCalled:     true,
			wantCommand:    usecase.ValueMetricCommand{Type: "gauge", Name: "Alloc"},
			wantResp: &Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: float64Ptr(123.45),
			},
		},
		{
			name:           "success counter returns JSON",
			body:           `{"id":"PollCount","type":"counter"}`,
			contentType:    "application/json",
			useCaseResult:  "7",
			wantStatusCode: http.StatusOK,
			wantCalled:     true,
			wantCommand:    usecase.ValueMetricCommand{Type: "counter", Name: "PollCount"},
			wantResp: &Metrics{
				ID:    "PollCount",
				MType: "counter",
				Delta: int64Ptr(7),
			},
		},
		{
			name:           "invalid content type returns 400",
			body:           `{"id":"Alloc","type":"gauge"}`,
			contentType:    "text/plain",
			wantStatusCode: http.StatusBadRequest,
			wantCalled:     false,
		},
		{
			name:           "invalid json returns 400",
			body:           `{"id":"Alloc","type":"gauge"`,
			contentType:    "application/json",
			wantStatusCode: http.StatusBadRequest,
			wantCalled:     false,
		},
		{
			name:           "metric not found returns 404",
			body:           `{"id":"Unknown","type":"gauge"}`,
			contentType:    "application/json",
			useCaseErr:     applicationerror.ErrMetricNotFound,
			wantStatusCode: http.StatusNotFound,
			wantCalled:     true,
			wantCommand:    usecase.ValueMetricCommand{Type: "gauge", Name: "Unknown"},
		},
		{
			name:           "unsupported metric type returns 400",
			body:           `{"id":"Alloc","type":"hist"}`,
			contentType:    "application/json",
			useCaseErr:     metric.ErrUnsupportedMetricType,
			wantStatusCode: http.StatusBadRequest,
			wantCalled:     true,
			wantCommand:    usecase.ValueMetricCommand{Type: "hist", Name: "Alloc"},
		},
		{
			name:           "empty name from usecase returns 400",
			body:           `{"id":"","type":"gauge"}`,
			contentType:    "application/json",
			useCaseErr:     metric.ErrNameEmpty,
			wantStatusCode: http.StatusBadRequest,
			wantCalled:     true,
			wantCommand:    usecase.ValueMetricCommand{Type: "gauge", Name: ""},
		},
		{
			name:           "unexpected error returns 500",
			body:           `{"id":"Alloc","type":"gauge"}`,
			contentType:    "application/json",
			useCaseErr:     errors.New("boom"),
			wantStatusCode: http.StatusInternalServerError,
			wantCalled:     true,
			wantCommand:    usecase.ValueMetricCommand{Type: "gauge", Name: "Alloc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			valueSpy := &valueMetricsUseCaseSpy{
				result: tt.useCaseResult,
				err:    tt.useCaseErr,
			}

			h := NewHandler(updateUseCaseNoop{}, valueSpy, listUseCaseNoop{})

			r := chi.NewRouter()
			r.Post("/value", h.ValueJSON)

			req := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(tt.body))

			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			rec := httptest.NewRecorder()

			// Act
			r.ServeHTTP(rec, req)

			// Assert
			if rec.Code != tt.wantStatusCode {
				t.Fatalf("expected status %d, got %d", tt.wantStatusCode, rec.Code)
			}

			if valueSpy.called != tt.wantCalled {
				t.Fatalf("expected usecase called=%v, got %v", tt.wantCalled, valueSpy.called)
			}

			if tt.wantCalled && valueSpy.command != tt.wantCommand {
				t.Fatalf("expected command %+v, got %+v", tt.wantCommand, valueSpy.command)
			}

			if tt.wantStatusCode == http.StatusOK {
				ct := rec.Header().Get("Content-Type")
				if !strings.Contains(ct, "application/json") {
					t.Fatalf("expected application/json content type, got %q", ct)
				}

				var got Metrics
				if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode response json: %v", err)
				}

				if got.ID != tt.wantResp.ID {
					t.Fatalf("expected id %q, got %q", tt.wantResp.ID, got.ID)
				}
				if got.MType != tt.wantResp.MType {
					t.Fatalf("expected type %q, got %q", tt.wantResp.MType, got.MType)
				}

				if tt.wantResp.Value != nil {
					if got.Value == nil {
						t.Fatal("expected value to be set")
					}
					if *got.Value != *tt.wantResp.Value {
						t.Fatalf("expected value %v, got %v", *tt.wantResp.Value, *got.Value)
					}
				} else if got.Value != nil {
					t.Fatalf("expected value to be nil, got %v", *got.Value)
				}

				if tt.wantResp.Delta != nil {
					if got.Delta == nil {
						t.Fatal("expected delta to be set")
					}
					if *got.Delta != *tt.wantResp.Delta {
						t.Fatalf("expected delta %d, got %d", *tt.wantResp.Delta, *got.Delta)
					}
				} else if got.Delta != nil {
					t.Fatalf("expected delta to be nil, got %d", *got.Delta)
				}
			}
		})
	}
}

func float64Ptr(v float64) *float64 { return &v }
func int64Ptr(v int64) *int64       { return &v }
