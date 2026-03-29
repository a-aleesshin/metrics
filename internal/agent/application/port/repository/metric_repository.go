package repository

import (
	"github.com/a-aleesshin/metrics/internal/agent/domain/metric"
)

type MetricsState struct {
	Counters []metric.Counter
	Gauges   []metric.Gauge
}

type MetricRepository interface {
	SetGauge(gauge *metric.Gauge) error
	AddCounter(counter *metric.Counter) error

	GetMetrics() (MetricsState, error)
}
