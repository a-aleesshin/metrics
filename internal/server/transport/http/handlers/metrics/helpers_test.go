package metrics

import (
	"context"

	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
)

type updateMetricUseCaseSpy struct {
	command usecase.UpdateMetricCommand
	err     error
	called  bool
}

type valueMetricsUseCaseSpy struct {
	command usecase.ValueMetricCommand
	result  string
	err     error
	called  bool
}

func (s *valueMetricsUseCaseSpy) Execute(ctx context.Context, cmd usecase.ValueMetricCommand) (string, error) {
	s.called = true
	s.command = cmd
	if s.err != nil {
		return "", s.err
	}
	return s.result, nil
}

func updatesCommandsEqual(a, b usecase.UpdatesMetricsCommand) bool {
	if len(a.Metrics) != len(b.Metrics) {
		return false
	}

	for i := range a.Metrics {
		if a.Metrics[i].Name != b.Metrics[i].Name {
			return false
		}
		if a.Metrics[i].MType != b.Metrics[i].MType {
			return false
		}

		if !float64PtrEqual(a.Metrics[i].Value, b.Metrics[i].Value) {
			return false
		}

		if !int64PtrEqual(a.Metrics[i].Delta, b.Metrics[i].Delta) {
			return false
		}
	}

	return true
}

func float64PtrEqual(a, b *float64) bool {
	if a == nil || b == nil {
		return a == b
	}

	return *a == *b
}

func int64PtrEqual(a, b *int64) bool {
	if a == nil || b == nil {
		return a == b
	}

	return *a == *b
}
