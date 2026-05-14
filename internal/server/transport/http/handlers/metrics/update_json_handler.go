package metrics

import (
	"encoding/json"
	"net/http"
	"strconv"

	platformhttp "github.com/a-aleesshin/metrics/internal/platform/http"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/httperror"
)

type UpdateJsonHandler struct {
	updateMetric UpdateMetricsUseCase
}

func NewUpdateJsonHandler(usecase UpdateMetricsUseCase) *UpdateJsonHandler {
	return &UpdateJsonHandler{updateMetric: usecase}
}

func (h *UpdateJsonHandler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// чтобы json был не пустым
	if err := json.NewEncoder(w).Encode(req); err != nil {
		return
	}

	return
}
