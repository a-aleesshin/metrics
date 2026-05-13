package healths

import (
	"context"
	"net/http"

	"github.com/a-aleesshin/metrics/internal/platform/health"
	platformhttp "github.com/a-aleesshin/metrics/internal/platform/http"
)

type HealthService interface {
	Check(ctx context.Context) health.Report
}

type PingHandler struct {
	healthService HealthService
}

func NewPingHandler(healthService HealthService) *PingHandler {
	return &PingHandler{healthService: healthService}
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	report := h.healthService.Check(r.Context())

	statusCode := http.StatusOK
	if report.Status != "ok" {
		statusCode = http.StatusInternalServerError
	}

	platformhttp.WriteJSON(w, statusCode, report.Checks)
}
