package metrics

import (
	"net/http"

	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/httperror"
	"github.com/go-chi/chi/v5"
)

type UpdateMetricsUseCase interface {
	Execute(command usecase.UpdateMetricCommand) error
}

type UpdateHandler struct {
	updateMetric UpdateMetricsUseCase
}

func NewUpdateHandler(updateMetric UpdateMetricsUseCase) *UpdateHandler {
	return &UpdateHandler{updateMetric: updateMetric}
}

func (h *UpdateHandler) Update(w http.ResponseWriter, r *http.Request) {
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
		httperror.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
