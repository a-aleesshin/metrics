package metrics

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
)

func (h *Handler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	if !h.IsJSON(r) {
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

	err := h.updateMetric.Execute(command)

	if err != nil {
		h.writeError(w, err)
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
