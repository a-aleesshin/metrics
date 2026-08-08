package healths

import (
	"context"
	"net/http"

	"github.com/a-aleesshin/metrics/internal/platform/health"
	platformhttp "github.com/a-aleesshin/metrics/internal/platform/http"
)

// HealthService выполняет проверки состояния зависимостей и возвращает сводный отчёт.
type HealthService interface {
	Check(ctx context.Context) health.Report
}

// PingHandler отдаёт результат проверки состояния сервиса.
type PingHandler struct {
	healthService HealthService
}

// NewPingHandler создаёт хендлер проверки состояния поверх HealthService.
func NewPingHandler(healthService HealthService) *PingHandler {
	return &PingHandler{healthService: healthService}
}

// Ping обрабатывает GET /ping: возвращает JSON с результатами проверок,
// 200 — если статус "ok", иначе 500.
func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	report := h.healthService.Check(r.Context())

	statusCode := http.StatusOK
	if report.Status != "ok" {
		statusCode = http.StatusInternalServerError
	}

	platformhttp.WriteJSON(w, statusCode, report.Checks)
}
