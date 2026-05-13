package httperror

import (
	"errors"
	"net/http"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

func WriteError(w http.ResponseWriter, err error) {
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
