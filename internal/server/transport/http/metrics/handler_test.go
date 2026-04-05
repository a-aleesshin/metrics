package metrics

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
	"github.com/go-chi/chi/v5"
)

func TestHandler_writeError(t *testing.T) {
	h := &Handler{}

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
			h.writeError(rec, tt.err)

			// Assert
			if rec.Code != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, rec.Code)
			}
		})
	}
}

func TestHandler_RegisterRoutes_Smoke(t *testing.T) {
	// Arrange
	h := NewHandler(updateUseCaseNoop{}, valueUseCaseNoop{}, listUseCaseNoop{})
	r := chi.NewRouter()

	// Act
	h.RegisterRoutes(r)

	// Assert
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/update/gauge/Alloc/1"},
		{http.MethodGet, "/value/gauge/Alloc"},
		{http.MethodGet, "/"},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code == http.StatusNotFound {
			t.Fatalf("%s %s should be registered, got 404", tc.method, tc.path)
		}
	}
}
