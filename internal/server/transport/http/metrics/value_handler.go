package metrics

import (
	"errors"
	"net/http"

	applicationerror "github.com/a-aleesshin/metrics/internal/server/application/error"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
	"github.com/go-chi/chi/v5"
)

type ValueMetricUseCase interface {
	Execute(cmd usecase.ValueMetricCommand) (string, error)
}

func (h *Handler) Value(w http.ResponseWriter, r *http.Request) {
	typeMetric := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	command := usecase.ValueMetricCommand{
		Type: typeMetric,
		Name: name,
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

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(result))
}
