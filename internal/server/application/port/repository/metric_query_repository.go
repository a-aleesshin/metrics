package repository

import "github.com/a-aleesshin/metrics/internal/server/domain/metric"

type GaugeSnapshot struct {
	Name  string
	Value float64
}

type CounterSnapshot struct {
	Name  string
	Delta int64
}

type MetricQueryRepository interface {
	ListGauges() ([]GaugeSnapshot, error)
	ListCounters() ([]CounterSnapshot, error)

	FindGaugeByName(name metric.Name) (value float64, found bool, err error)
	FindCounterByName(name metric.Name) (delta int64, found bool, err error)
}
