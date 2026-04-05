package metrics

import (
	"errors"
	"net/http"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	updateMetric   UpdateMetricsUseCase
	getValueMetric ValueMetricUseCase
	listMetric     ListMetricsUseCase
}

func NewHandler(
	updateMetric UpdateMetricsUseCase,
	getValueMetric ValueMetricUseCase,
	listMetric ListMetricsUseCase,
) *Handler {
	return &Handler{
		updateMetric:   updateMetric,
		getValueMetric: getValueMetric,
		listMetric:     listMetric,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/update/{type}/{name}/{value}", h.Update)
	r.Get("/value/{type}/{name}", h.Value)
	r.Get("/", h.List)
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
