package repository

import (
	"context"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type GaugeSnapshot struct {
	Name  string
	Value float64
}

type CounterSnapshot struct {
	Name  string
	Delta int64
}

type MetricQueryRepository interface {
	ListGauges(ctx context.Context) ([]GaugeSnapshot, error)
	ListCounters(ctx context.Context) ([]CounterSnapshot, error)

	FindGaugeByName(ctx context.Context, name metric.Name) (value float64, found bool, err error)
	FindCounterByName(ctx context.Context, name metric.Name) (delta int64, found bool, err error)
}
