package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
	"github.com/go-chi/chi/v5"
)

type updatesMetricsUseCaseSpy struct {
	command usecase.UpdatesMetricsCommand
	err     error
	called  bool
}

func (s *updatesMetricsUseCaseSpy) Execute(ctx context.Context, command usecase.UpdatesMetricsCommand) error {
	s.called = true
	s.command = command

	if s.err != nil {
		return s.err
	}

	return nil
}

func TestUpdatesHandler_Updates(t *testing.T) {
	gaugeValue := 123.45
	counterDelta := int64(7)

	tests := []struct {
		name           string
		body           string
		contentType    string
		useCaseErr     error
		wantStatusCode int
		wantCalled     bool
		wantCommand    usecase.UpdatesMetricsCommand
	}{
		{
			name:           "success batch",
			body:           `[{"id":"Alloc","type":"gauge","value":123.45},{"id":"PollCount","type":"counter","delta":7}]`,
			contentType:    "application/json",
			wantStatusCode: http.StatusOK,
			wantCalled:     true,
			wantCommand: usecase.UpdatesMetricsCommand{
				Metrics: []usecase.MetricUpdatesCommand{
					{Name: "Alloc", MType: "gauge", Value: &gaugeValue},
					{Name: "PollCount", MType: "counter", Delta: &counterDelta},
				},
			},
		},
		{
			name:           "empty batch",
			body:           `[]`,
			contentType:    "application/json",
			wantStatusCode: http.StatusOK,
			wantCalled:     false,
		},
		{
			name:           "invalid content type",
			body:           `[{"id":"Alloc","type":"gauge","value":1}]`,
			contentType:    "text/plain",
			wantStatusCode: http.StatusBadRequest,
			wantCalled:     false,
		},
		{
			name:           "invalid json",
			body:           `[{"id":"Alloc","type":"gauge","value":invalid}]`,
			contentType:    "application/json",
			wantStatusCode: http.StatusBadRequest,
			wantCalled:     false,
		},
		{
			name:           "missing gauge value",
			body:           `[{"id":"Alloc","type":"gauge"}]`,
			contentType:    "application/json",
			wantStatusCode: http.StatusBadRequest,
			wantCalled:     false,
		},
		{
			name:           "missing counter delta",
			body:           `[{"id":"PollCount","type":"counter"}]`,
			contentType:    "application/json",
			wantStatusCode: http.StatusBadRequest,
			wantCalled:     false,
		},
		{
			name:           "unsupported metric type",
			body:           `[{"id":"Alloc","type":"unknown","value":123.45}]`,
			contentType:    "application/json",
			wantStatusCode: http.StatusBadRequest,
			wantCalled:     false,
		},
		{
			name:           "usecase invalid value maps 400",
			body:           `[{"id":"Alloc","type":"gauge","value":123.45}]`,
			contentType:    "application/json",
			useCaseErr:     metric.ErrInvalidMetricValue,
			wantStatusCode: http.StatusBadRequest,
			wantCalled:     true,
			wantCommand: usecase.UpdatesMetricsCommand{
				Metrics: []usecase.MetricUpdatesCommand{
					{Name: "Alloc", MType: "gauge", Value: &gaugeValue},
				},
			},
		},
		{
			name:           "usecase unexpected error maps 500",
			body:           `[{"id":"Alloc","type":"gauge","value":123.45}]`,
			contentType:    "application/json",
			useCaseErr:     http.ErrBodyNotAllowed,
			wantStatusCode: http.StatusInternalServerError,
			wantCalled:     true,
			wantCommand: usecase.UpdatesMetricsCommand{
				Metrics: []usecase.MetricUpdatesCommand{
					{Name: "Alloc", MType: "gauge", Value: &gaugeValue},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			useCaseSpy := &updatesMetricsUseCaseSpy{err: tt.useCaseErr}
			handler := NewUpdatesHandler(useCaseSpy)

			r := chi.NewRouter()
			r.Post("/updates", handler.Updates)

			req := httptest.NewRequest(http.MethodPost, "/updates", strings.NewReader(tt.body))
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

			if useCaseSpy.called != tt.wantCalled {
				t.Fatalf("expected use case called=%v, got %v", tt.wantCalled, useCaseSpy.called)
			}

			if tt.wantCalled && !updatesCommandsEqual(useCaseSpy.command, tt.wantCommand) {
				t.Fatalf("expected command %+v, got %+v", tt.wantCommand, useCaseSpy.command)
			}

			if tt.wantStatusCode == http.StatusOK {
				ct := rec.Header().Get("Content-Type")
				if !strings.Contains(ct, "application/json") {
					t.Fatalf("expected Content-Type application/json, got %q", ct)
				}
			}
		})
	}
}
