// Package healths содержит HTTP-хендлеры проверки состояния сервиса (GET /ping).
package healths

import "github.com/go-chi/chi/v5"

// Handler объединяет хендлеры проверок состояния и регистрирует их маршруты.
type Handler struct {
	ping *PingHandler
}

// NewHandler создаёт составной хендлер проверок состояния.
func NewHandler(ping *PingHandler) *Handler {
	return &Handler{ping: ping}
}

// RegisterRoutes регистрирует маршрут GET /ping.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/ping", h.ping.Ping)
}
