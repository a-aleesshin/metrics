package metrics

import (
	"context"
	"encoding/json"
	"net/http"

	platformhttp "github.com/a-aleesshin/metrics/internal/platform/http"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/httperror"
)

// UpdatesMetricsUseCase — usecase батчевого обновления набора метрик.
type UpdatesMetricsUseCase interface {
	Execute(ctx context.Context, command usecase.UpdatesMetricsCommand) error
}

// UpdatesHandler обновляет набор метрик из JSON-массива в теле запроса.
type UpdatesHandler struct {
	useCase UpdatesMetricsUseCase
	audit   AuditPublisher
}

// NewUpdatesHandler создаёт хендлер батчевого обновления метрик;
// audit может быть nil — тогда аудит отключён.
func NewUpdatesHandler(useCase UpdatesMetricsUseCase, audit AuditPublisher) *UpdatesHandler {
	return &UpdatesHandler{useCase: useCase, audit: audit}
}

// Updates обрабатывает POST /updates с JSON-массивом Metrics: валидирует и обновляет
// метрики одним батчем, публикует событие аудита. Отвечает 200 (в том числе на пустой
// массив); 400 — неверный Content-Type, тело, тип метрики или отсутствующее значение.
func (h *UpdatesHandler) Updates(w http.ResponseWriter, r *http.Request) {
	if !platformhttp.IsJSON(r) {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	var collection []Metrics
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&collection); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(collection) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		return
	}

	var metricsUpdated = make([]usecase.MetricUpdatesCommand, 0, len(collection))

	for _, metric := range collection {
		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				http.Error(w, "value is required", http.StatusBadRequest)
				return
			}
		case "counter":
			if metric.Delta == nil {
				http.Error(w, "delta is required", http.StatusBadRequest)
				return
			}
		default:
			http.Error(w, "unsupported type", http.StatusBadRequest)
			return
		}

		metricsUpdated = append(metricsUpdated, usecase.MetricUpdatesCommand{
			Name:  metric.ID,
			MType: metric.MType,
			Delta: metric.Delta,
			Value: metric.Value,
		})
	}

	command := usecase.UpdatesMetricsCommand{
		Metrics: metricsUpdated,
	}

	err := h.useCase.Execute(r.Context(), command)

	if err != nil {
		httperror.WriteError(w, err)
		return
	}

	if h.audit != nil {
		names := make([]string, 0, len(metricsUpdated))
		for _, m := range metricsUpdated {
			names = append(names, m.Name)
		}

		h.audit.Publish(r.Context(), names, platformhttp.ClientIP(r))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}
