package repository

import "github.com/a-aleesshin/metrics/internal/server/domain/metric"

type MetricsState struct {
	Counters []*metric.Counter
	Gauges   []*metric.Gauge
}

type MetricStateRepository interface {
	GetAllMetrics() (MetricsState, error)
}
