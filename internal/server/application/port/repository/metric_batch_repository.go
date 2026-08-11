package repository

import (
	"context"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

// MetricBatch — набор метрик для батч-обновления.
type MetricBatch struct {
	Gauges   []*metric.Gauge
	Counters []*metric.Counter
}

// MetricBatchRepository — порт атомарного батч-обновления метрик:
// gauge перезаписываются, counter накапливаются.
type MetricBatchRepository interface {
	UpdateBatch(ctx context.Context, batch MetricBatch) error
}
