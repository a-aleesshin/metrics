package repository

import (
	"context"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type MetricBatch struct {
	Gauges   []*metric.Gauge
	Counters []*metric.Counter
}

type MetricBatchRepository interface {
	UpdateBatch(ctx context.Context, batch MetricBatch) error
}
