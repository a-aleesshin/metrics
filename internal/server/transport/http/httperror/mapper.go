// Package httperror отображает доменные ошибки метрик на HTTP-статусы ответа.
package httperror

import (
	"errors"
	"net/http"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

// WriteError пишет HTTP-статус по доменной ошибке: 404 — пустое имя метрики,
// 400 — неверный тип или значение, 500 — прочие ошибки.
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
