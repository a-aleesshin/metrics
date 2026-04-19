package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type RouterRigister interface {
	RegisterRoutes(router chi.Router)
}

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
