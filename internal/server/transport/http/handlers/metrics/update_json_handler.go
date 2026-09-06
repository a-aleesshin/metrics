package metrics

import (
	"encoding/json"
	"net/http"
	"strconv"

	platformhttp "github.com/a-aleesshin/metrics/internal/platform/http"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/httperror"
)

// UpdateJSONHandler обновляет одну метрику из JSON-тела запроса.
type UpdateJSONHandler struct {
	updateMetric UpdateMetricsUseCase
	audit        AuditPublisher
}

// NewUpdateJSONHandler создаёт хендлер обновления метрики из JSON;
// audit может быть nil — тогда аудит отключён.
func NewUpdateJSONHandler(usecase UpdateMetricsUseCase, audit AuditPublisher) *UpdateJSONHandler {
	return &UpdateJSONHandler{updateMetric: usecase, audit: audit}
}

// UpdateJSON обрабатывает POST /update с JSON-телом Metrics: обновляет метрику,
// публикует событие аудита и возвращает 200 с эхом запроса; 400 — неверный
// Content-Type, тело, тип метрики или отсутствующее значение.
func (h *UpdateJSONHandler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	if !platformhttp.IsJSON(r) {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	var req Metrics
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var rawValue string

	switch req.MType {
	case "gauge":
		if req.Value == nil {
			http.Error(w, "value is required", http.StatusBadRequest)
			return
		}
		rawValue = strconv.FormatFloat(*req.Value, 'f', -1, 64)

	case "counter":
		if req.Delta == nil {
			http.Error(w, "delta is required", http.StatusBadRequest)
			return
		}
		rawValue = strconv.FormatInt(*req.Delta, 10)

	default:
		http.Error(w, "unsupported type", http.StatusBadRequest)
		return
	}

	command := usecase.UpdateMetricCommand{
		Type:  req.MType,
		Name:  req.ID,
		Value: rawValue,
	}

	err := h.updateMetric.Execute(r.Context(), command)

	if err != nil {
		httperror.WriteError(w, err)
		return
	}

	if h.audit != nil {
		h.audit.Publish(r.Context(), []string{req.ID}, platformhttp.ClientIP(r))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// чтобы json был не пустым
	if err := json.NewEncoder(w).Encode(req); err != nil {
		return
	}

}
