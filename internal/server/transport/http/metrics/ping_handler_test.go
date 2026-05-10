package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-aleesshin/metrics/internal/platform/health"
	"github.com/stretchr/testify/require"
)

type stubHealthServiceStub struct {
	report health.Report
}

func (s *stubHealthServiceStub) Check(ctx context.Context) health.Report {
	return s.report
}

func TestHandler_Ping_OK(t *testing.T) {
	// Arrange
	handler := Handler{
		healthService: &stubHealthServiceStub{
			report: health.Report{
				Status: "ok",
				Checks: []health.CheckResult{
					{
						Name:   "postgres",
						Status: "ok",
					},
				},
			},
		},
	}

	// Act
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	handler.Ping(rec, req)

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	require.JSONEq(t,
		`[
    {
        "name": "postgres",
        "status": "ok"
    }
]
`,
		rec.Body.String(),
	)
}

func TestHandler_Ping_Unhealthy(t *testing.T) {
	// Arrange
	handler := Handler{
		healthService: &stubHealthServiceStub{
			report: health.Report{
				Status: "error",
				Checks: []health.CheckResult{
					{
						Name:   "postgres",
						Status: "error",
					},
				},
			},
		},
	}

	// Act
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	handler.Ping(rec, req)

	// Assert
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}
