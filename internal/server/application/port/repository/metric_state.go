package repository

import (
	"context"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

// MetricsState — полное текущее состояние метрик хранилища.
type MetricsState struct {
	Counters []*metric.Counter
	Gauges   []*metric.Gauge
}

// MetricStateRepository — порт чтения полного состояния метрик, например для snapshot.
type MetricStateRepository interface {
	GetAllMetrics(ctx context.Context) (MetricsState, error)
}
