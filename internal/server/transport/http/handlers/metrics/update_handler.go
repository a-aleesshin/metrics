package metrics

import (
	"context"
	"net/http"

	platformhttp "github.com/a-aleesshin/metrics/internal/platform/http"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/httperror"
	"github.com/go-chi/chi/v5"
)

// UpdateMetricsUseCase — usecase обновления одной метрики по строковому значению.
type UpdateMetricsUseCase interface {
	Execute(ctx context.Context, command usecase.UpdateMetricCommand) error
}

// AuditPublisher публикует событие аудита с именами изменённых метрик и IP клиента.
type AuditPublisher interface {
	Publish(ctx context.Context, metricNames []string, ipAddress string)
}

// UpdateHandler обновляет метрику по значениям из URL-параметров.
type UpdateHandler struct {
	updateMetric UpdateMetricsUseCase
	audit        AuditPublisher
}

// NewUpdateHandler создаёт хендлер обновления метрики через URL-параметры;
// audit может быть nil — тогда аудит отключён.
func NewUpdateHandler(updateMetric UpdateMetricsUseCase, audit AuditPublisher) *UpdateHandler {
	return &UpdateHandler{updateMetric: updateMetric, audit: audit}
}

// Update обрабатывает POST /update/{type}/{name}/{value}: обновляет метрику
// и публикует событие аудита. Отвечает 200; коды ошибок задаёт httperror.WriteError.
func (h *UpdateHandler) Update(w http.ResponseWriter, r *http.Request) {
	typeMetric := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	command := usecase.UpdateMetricCommand{
		Type:  typeMetric,
		Name:  name,
		Value: value,
	}

	err := h.updateMetric.Execute(r.Context(), command)

	if err != nil {
		httperror.WriteError(w, err)
		return
	}

	if h.audit != nil {
		h.audit.Publish(r.Context(), []string{name}, platformhttp.ClientIP(r))
	}

	w.WriteHeader(http.StatusOK)
}
