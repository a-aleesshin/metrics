package metrics

import (
	"net/http"

	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/go-chi/chi/v5"
)

type UpdateMetricsUseCase interface {
	Execute(command usecase.UpdateMetricCommand) error
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	typeMetric := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	command := usecase.UpdateMetricCommand{
		Type:  typeMetric,
		Name:  name,
		Value: value,
	}

	err := h.updateMetric.Execute(command)

	if err != nil {
		h.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
