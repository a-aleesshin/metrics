package metrics

import (
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	update     *UpdateHandler
	updateJSON *UpdateJsonHandler
	updates    *UpdatesHandler
	value      *ValueHandler
	valueJSON  *ValueJsonHandler
	list       *ListMetricsHandler
}

func NewHandler(
	update *UpdateHandler,
	updateJSON *UpdateJsonHandler,
	updates *UpdatesHandler,
	value *ValueHandler,
	valueJSON *ValueJsonHandler,
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

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/update", h.updateJSON.UpdateJSON)
	r.Post("/updates", h.updates.Updates)
	r.Post("/update/{type}/{name}/{value}", h.update.Update)

	r.Post("/value", h.valueJSON.ValueJSON)
	r.Get("/value/{type}/{name}", h.value.Value)

	r.Get("/", h.list.List)
}
