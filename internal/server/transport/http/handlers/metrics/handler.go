// Package metrics содержит HTTP-хендлеры приёма и выдачи метрик gauge/counter:
// обновление через URL-параметры и JSON (одиночное и батчевое), чтение значений и HTML-список.
package metrics

import (
	"github.com/go-chi/chi/v5"
)

// Handler объединяет хендлеры метрик и регистрирует их маршруты в роутере chi.
type Handler struct {
	update     *UpdateHandler
	updateJSON *UpdateJSONHandler
	updates    *UpdatesHandler
	value      *ValueHandler
	valueJSON  *ValueJSONHandler
	list       *ListMetricsHandler
}

// NewHandler создаёт составной хендлер метрик из отдельных хендлеров операций.
func NewHandler(
	update *UpdateHandler,
	updateJSON *UpdateJSONHandler,
	updates *UpdatesHandler,
	value *ValueHandler,
	valueJSON *ValueJSONHandler,
	list *ListMetricsHandler,
) *Handler {
	return &Handler{
		update:     update,
		updateJSON: updateJSON,
		updates:    updates,
		value:      value,
		valueJSON:  valueJSON,
		list:       list,
	}
}

// RegisterRoutes регистрирует маршруты метрик: POST /update, POST /updates,
// POST /update/{type}/{name}/{value}, POST /value, GET /value/{type}/{name}, GET /.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/update", h.updateJSON.UpdateJSON)
	r.Post("/updates", h.updates.Updates)
	r.Post("/update/{type}/{name}/{value}", h.update.Update)

	r.Post("/value", h.valueJSON.ValueJSON)
	r.Get("/value/{type}/{name}", h.value.Value)

	r.Get("/", h.list.List)
}
