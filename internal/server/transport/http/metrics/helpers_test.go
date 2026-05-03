package metrics

import (
	"context"
	"net/http"

	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/go-chi/chi/v5"
)

type updateMetricUseCaseSpy struct {
	command usecase.UpdateMetricCommand
	err     error
	called  bool
}

func withChiParams(r *http.Request, params map[string]string) *http.Request {
	rc := chi.NewRouteContext()
	for k, v := range params {
		rc.URLParams.Add(k, v)
	}
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rc)
	return r.WithContext(ctx)
}

type valueMetricsUseCaseSpy struct {
	command usecase.ValueMetricCommand
	result  string
	err     error
	called  bool
}

func (s *valueMetricsUseCaseSpy) Execute(cmd usecase.ValueMetricCommand) (string, error) {
	s.called = true
	s.command = cmd
	if s.err != nil {
		return "", s.err
	}
	return s.result, nil
}
