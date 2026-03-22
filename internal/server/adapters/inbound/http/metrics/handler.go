package metrics

import (
	"errors"
	"net/http"
	"strings"

	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type UpdateMetricUseCase interface {
	Execute(command usecase.UpdateMetricCommand) error
}

type Handler struct {
	updateMetric UpdateMetricUseCase
}

func NewHandler(updateMetric UpdateMetricUseCase) *Handler {
	return &Handler{updateMetric: updateMetric}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(parts) < 4 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if parts[0] != "update" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	command := usecase.UpdateMetricCommand{
		Type:  parts[1],
		Name:  parts[2],
		Value: parts[3],
	}

	err := h.updateMetric.Execute(command)

	if err != nil {
		h.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, metric.ErrNameEmpty):
		w.WriteHeader(http.StatusNotFound)
		return
	case errors.Is(err, metric.ErrUnsupportedMetricType):
		w.WriteHeader(http.StatusBadRequest)
		return
	case errors.Is(err, metric.ErrInvalidMetricType):
		w.WriteHeader(http.StatusBadRequest)
		return
	case errors.Is(err, metric.ErrInvalidMetricValue):
		w.WriteHeader(http.StatusBadRequest)
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
