package metrics

import (
	"errors"
	"net/http"
	"strconv"

	applicationerror "github.com/a-aleesshin/metrics/internal/server/application/error"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

func (h *Handler) ValueJSON(w http.ResponseWriter, r *http.Request) {
	if !h.IsJSON(r) {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	var req Metrics

	if err := h.DecodeJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	command := usecase.ValueMetricCommand{
		Type: req.MType,
		Name: req.ID,
	}

	result, err := h.getValueMetric.Execute(command)

	if err != nil {
		switch {
		case errors.Is(err, applicationerror.ErrMetricNotFound):
			http.NotFound(w, r)
		case errors.Is(err, metric.ErrUnsupportedMetricType), errors.Is(err, metric.ErrNameEmpty):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	resp := Metrics{
		ID:    req.ID,
		MType: req.MType,
	}

	switch req.MType {
	case "gauge":
		v, err := strconv.ParseFloat(result, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp.Value = &v
	case "counter":
		d, err := strconv.ParseInt(result, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp.Delta = &d
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.WriteJSON(w, http.StatusOK, resp)
}
