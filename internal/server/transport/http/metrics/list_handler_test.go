package metrics

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-aleesshin/metrics/internal/server/application/dto"
	"github.com/go-chi/chi/v5"
)

type listMetricsUseCaseSpy struct {
	result dto.ListMetricsResult
	err    error
	called bool
}

func (s *listMetricsUseCaseSpy) Execute() (dto.ListMetricsResult, error) {
	s.called = true
	if s.err != nil {
		return dto.ListMetricsResult{}, s.err
	}
	return s.result, nil
}

func TestHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		useCaseResult  dto.ListMetricsResult
		useCaseErr     error
		wantStatusCode int
		wantBodyParts  []string
	}{
		{
			name: "success returns html page with metrics",
			useCaseResult: dto.ListMetricsResult{
				Items: []dto.MetricView{
					{Type: "gauge", Name: "Alloc", Value: "123.45"},
					{Type: "counter", Name: "PollCount", Value: "7"},
				},
			},
			wantStatusCode: http.StatusOK,
			wantBodyParts: []string{
				"<h1>Metrics</h1>",
				"gauge Alloc = 123.45",
				"counter PollCount = 7",
			},
		},
		{
			name:           "usecase error returns 500",
			useCaseErr:     errors.New("boom"),
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			listSpy := &listMetricsUseCaseSpy{
				result: tt.useCaseResult,
				err:    tt.useCaseErr,
			}
			h := NewHandler(updateUseCaseNoop{}, valueUseCaseNoop{}, listSpy)

			r := chi.NewRouter()
			r.Get("/", h.List)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			// Act
			r.ServeHTTP(rec, req)

			// Assert
			if rec.Code != tt.wantStatusCode {
				t.Fatalf("expected status %d, got %d", tt.wantStatusCode, rec.Code)
			}

			if !listSpy.called {
				t.Fatal("expected list use case to be called")
			}

			if tt.wantStatusCode == http.StatusOK {
				ct := rec.Header().Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Fatalf("expected text/html content type, got %q", ct)
				}

				body := rec.Body.String()
				for _, part := range tt.wantBodyParts {
					if !strings.Contains(body, part) {
						t.Fatalf("expected body to contain %q, got body: %s", part, body)
					}
				}
			}
		})
	}
}
