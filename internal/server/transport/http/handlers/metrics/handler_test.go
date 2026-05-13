package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHandler_RegisterRoutes_Smoke(t *testing.T) {
	updateUC := updateUseCaseNoop{}
	valueUC := valueUseCaseNoop{}
	listUC := listUseCaseNoop{}
	updatesUC := updatesUseCaseNoop{}

	h := NewHandler(
		NewUpdateHandler(updateUC),
		NewUpdateJsonHandler(updateUC),
		NewUpdatesHandler(updatesUC),
		NewValueHandler(valueUC),
		NewValueJsonHandler(valueUC),
		NewListMetricsHandler(listUC),
	)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/update/gauge/Alloc/1"},
		{http.MethodPost, "/update"},
		{http.MethodPost, "/updates"},
		{http.MethodPost, "/value"},
		{http.MethodGet, "/value/gauge/Alloc"},
		{http.MethodGet, "/"},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code == http.StatusNotFound {
			t.Fatalf("%s %s should be registered, got 404", tc.method, tc.path)
		}
	}
}
