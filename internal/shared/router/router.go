package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type RouterRigister interface {
	RegisterRoutes(router chi.Router)
}

func New(registrars ...RouterRigister) http.Handler {
	r := chi.NewRouter()

	for _, rr := range registrars {
		rr.RegisterRoutes(r)
	}

	return r
}
