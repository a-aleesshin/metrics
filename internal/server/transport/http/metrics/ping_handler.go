package metrics

import (
	"context"
	"net/http"

	"github.com/a-aleesshin/metrics/internal/platform/health"
)

type HealthService interface {
	Check(ctx context.Context) health.Report
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	report := h.healthService.Check(r.Context())

	statusCode := http.StatusOK
	if report.Status != "ok" {
		statusCode = http.StatusInternalServerError
	}

	h.WriteJSON(w, statusCode, report.Checks)
}
