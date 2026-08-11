// Package http содержит общие HTTP-утилиты: роутер chi, JSON-хелперы и определение IP клиента.
package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// RouterRigister регистрирует свои маршруты в chi-роутере.
type RouterRigister interface {
	// RegisterRoutes добавляет обработчики в переданный роутер.
	RegisterRoutes(router chi.Router)
}

// New собирает chi-роутер: подключает middleware и регистрирует маршруты registrars.
func New(middlewares []func(http.Handler) http.Handler, registrars ...RouterRigister) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.StripSlashes)

	for _, mw := range middlewares {
		r.Use(mw)
	}

	for _, rr := range registrars {
		rr.RegisterRoutes(r)
	}

	return r
}
