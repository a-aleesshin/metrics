package metrics

import (
	"context"
	"errors"
	"net/http"

	applicationerror "github.com/a-aleesshin/metrics/internal/server/application/error"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
	"github.com/go-chi/chi/v5"
)

// ValueMetricUseCase — usecase получения строкового значения метрики по типу и имени.
type ValueMetricUseCase interface {
	Execute(ctx context.Context, cmd usecase.ValueMetricCommand) (string, error)
}

// ValueHandler отдаёт значение метрики в виде текста по URL-параметрам.
type ValueHandler struct {
	getValueMetric ValueMetricUseCase
}

// NewValueHandler создаёт хендлер чтения значения метрики через URL-параметры.
func NewValueHandler(getValueMetric ValueMetricUseCase) *ValueHandler {
	return &ValueHandler{getValueMetric: getValueMetric}
}

// Value обрабатывает GET /value/{type}/{name}: возвращает значение метрики
// как text/plain со статусом 200; 404 — метрика не найдена, 400 — неверный тип или имя.
func (h *ValueHandler) Value(w http.ResponseWriter, r *http.Request) {
	typeMetric := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	command := usecase.ValueMetricCommand{
		Type: typeMetric,
		Name: name,
	}

	result, err := h.getValueMetric.Execute(r.Context(), command)

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
