package metrics

import (
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

func TestHandler_Value(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		useCaseResult  string
		useCaseErr     error
		wantStatusCode int
		wantBody       string
		wantCommand    usecase.ValueMetricCommand
	}{
		{
			name:           "success gauge returns 200 and text value",
			path:           "/value/gauge/Alloc",
			useCaseResult:  "123.45",
			wantStatusCode: http.StatusOK,
			wantBody:       "123.45",
			wantCommand:    usecase.ValueMetricCommand{Type: "gauge", Name: "Alloc"},
		},
		{
			name:           "success counter returns 200 and text value",
			path:           "/value/counter/PollCount",
			useCaseResult:  "7",
			wantStatusCode: http.StatusOK,
			wantBody:       "7",
			wantCommand:    usecase.ValueMetricCommand{Type: "counter", Name: "PollCount"},
		},
		{
			name:           "metric not found returns 404",
			path:           "/value/gauge/Unknown",
			useCaseErr:     applicationerror.ErrMetricNotFound,
			wantStatusCode: http.StatusNotFound,
			wantCommand:    usecase.ValueMetricCommand{Type: "gauge", Name: "Unknown"},
		},
		{
			name:           "unsupported metric type returns 400",
			path:           "/value/hist/Alloc",
			useCaseErr:     metric.ErrUnsupportedMetricType,
			wantStatusCode: http.StatusBadRequest,
			wantCommand:    usecase.ValueMetricCommand{Type: "hist", Name: "Alloc"},
		},
		{
			name:           "empty name from usecase returns 400",
			path:           "/value/gauge/Alloc",
			useCaseErr:     metric.ErrNameEmpty,
			wantStatusCode: http.StatusBadRequest,
			wantCommand:    usecase.ValueMetricCommand{Type: "gauge", Name: "Alloc"},
		},
		{
			name:           "unexpected error returns 500",
			path:           "/value/gauge/Alloc",
			useCaseErr:     errors.New("boom"),
			wantStatusCode: http.StatusInternalServerError,
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
			r.Get("/value/{type}/{name}", h.Value)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			// Act
			r.ServeHTTP(rec, req)

			// Assert
			if rec.Code != tt.wantStatusCode {
				t.Fatalf("expected status %d, got %d", tt.wantStatusCode, rec.Code)
			}

			if !valueSpy.called {
				t.Fatal("expected value use case to be called")
			}

			if valueSpy.command != tt.wantCommand {
				t.Fatalf("expected command %+v, got %+v", tt.wantCommand, valueSpy.command)
			}

			if tt.wantStatusCode == http.StatusOK {
				ct := rec.Header().Get("Content-Type")
				if !strings.Contains(ct, "text/plain") {
					t.Fatalf("expected text/plain content type, got %q", ct)
				}
				if rec.Body.String() != tt.wantBody {
					t.Fatalf("expected body %q, got %q", tt.wantBody, rec.Body.String())
				}
			}
		})
	}
}
