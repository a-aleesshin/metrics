package repository

import (
	"context"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

// GaugeSnapshot — значение gauge для запросов чтения.
type GaugeSnapshot struct {
	Name  string
	Value float64
}

// CounterSnapshot — значение counter для запросов чтения.
type CounterSnapshot struct {
	Name  string
	Delta int64
}

// MetricQueryRepository — порт запросов чтения метрик: списки и поиск по имени;
// Find-методы сигнализируют отсутствие метрики флагом found.
type MetricQueryRepository interface {
	ListGauges(ctx context.Context) ([]GaugeSnapshot, error)
	ListCounters(ctx context.Context) ([]CounterSnapshot, error)

	FindGaugeByName(ctx context.Context, name metric.Name) (value float64, found bool, err error)
	FindCounterByName(ctx context.Context, name metric.Name) (delta int64, found bool, err error)
}
