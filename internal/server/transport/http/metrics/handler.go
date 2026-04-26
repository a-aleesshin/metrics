package metrics

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
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
	r.Post("/update", h.UpdateJSON)
	r.Post("/update/{type}/{name}/{value}", h.Update)
	r.Post("/value", h.ValueJSON)
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

func (h *Handler) IsJSON(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))

	if err != nil || mediaType != "application/json" {
		return false
	}

	return true
}

func (h *Handler) DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	if dec.More() {
		return fmt.Errorf("invalid json: multiple objects")
	}

	return nil
}

func (h *Handler) WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
