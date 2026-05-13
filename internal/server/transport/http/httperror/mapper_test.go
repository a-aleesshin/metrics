package httperror

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

func TestMapper_WriteError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"name empty -> 404", metric.ErrNameEmpty, http.StatusNotFound},
		{"unsupported type -> 400", metric.ErrUnsupportedMetricType, http.StatusBadRequest},
		{"invalid type -> 400", metric.ErrInvalidMetricType, http.StatusBadRequest},
		{"invalid value -> 400", metric.ErrInvalidMetricValue, http.StatusBadRequest},
		{"unknown -> 500", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			rec := httptest.NewRecorder()

			// Act
			WriteError(rec, tt.err)

			// Assert
			if rec.Code != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, rec.Code)
			}
		})
	}
}
