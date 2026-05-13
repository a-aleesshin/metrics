package healths

import "github.com/go-chi/chi/v5"

type Handler struct {
	ping *PingHandler
}

func NewHandler(ping *PingHandler) *Handler {
	return &Handler{ping: ping}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/ping", h.ping.Ping)
}
