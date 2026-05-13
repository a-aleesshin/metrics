package repository

import (
	"context"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type MetricsState struct {
	Counters []*metric.Counter
	Gauges   []*metric.Gauge
}

type MetricStateRepository interface {
	GetAllMetrics(ctx context.Context) (MetricsState, error)
}
