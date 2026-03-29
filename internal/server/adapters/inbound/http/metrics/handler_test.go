package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type updateMetricUseCaseSpy struct {
	command usecase.UpdateMetricCommand
	err     error
	called  bool
}

func (u *updateMetricUseCaseSpy) Execute(command usecase.UpdateMetricCommand) error {
	u.called = true
	u.command = command
	return u.err
}

func TestHandler_Update_ErrorMapping(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		useCaseErr     error
		wantStatusCode int
		wantCommand    usecase.UpdateMetricCommand
	}{
		{
			name:           "name empty returns 404",
			path:           "/update/gauge//123.45",
			useCaseErr:     metric.ErrNameEmpty,
			wantStatusCode: http.StatusNotFound,
			wantCommand: usecase.UpdateMetricCommand{
				Type:  "gauge",
				Name:  "",
				Value: "123.45",
			},
		},
		{
			name:           "unsupported metric type returns 400",
			path:           "/update/unknown/Alloc/123.45",
			useCaseErr:     metric.ErrUnsupportedMetricType,
			wantStatusCode: http.StatusBadRequest,
			wantCommand: usecase.UpdateMetricCommand{
				Type:  "unknown",
				Name:  "Alloc",
				Value: "123.45",
			},
		},
		{
			name:           "invalid metric value returns 400",
			path:           "/update/gauge/Alloc/invalid",
			useCaseErr:     metric.ErrInvalidMetricValue,
			wantStatusCode: http.StatusBadRequest,
			wantCommand: usecase.UpdateMetricCommand{
				Type:  "gauge",
				Name:  "Alloc",
				Value: "invalid",
			},
		},
		{
			name:           "unexpected error returns 500",
			path:           "/update/gauge/Alloc/123.45",
			useCaseErr:     http.ErrBodyNotAllowed,
			wantStatusCode: http.StatusInternalServerError,
			wantCommand: usecase.UpdateMetricCommand{
				Type:  "gauge",
				Name:  "Alloc",
				Value: "123.45",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			useCaseSpy := &updateMetricUseCaseSpy{
				err: tt.useCaseErr,
			}
			handler := NewHandler(useCaseSpy)

			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			req.Header.Set("Content-Type", "text/plain")

			rec := httptest.NewRecorder()

			// Act
			handler.Update(rec, req)

			// Assert
			if rec.Code != tt.wantStatusCode {
				t.Fatalf("expected status %d, got %d", tt.wantStatusCode, rec.Code)
			}

			if !useCaseSpy.called {
				t.Fatal("expected use case to be called")
			}

			if useCaseSpy.command != tt.wantCommand {
				t.Fatalf("expected command %+v, got %+v", tt.wantCommand, useCaseSpy.command)
			}
		})
	}
}
