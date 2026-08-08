// Package repository определяет порт хранилища собранных метрик агента.
package repository

import (
	"github.com/a-aleesshin/metrics/internal/agent/domain/metric"
)

// MetricsState — снимок всех накопленных метрик: счётчики и gauge.
type MetricsState struct {
	Counters []metric.Counter
	Gauges   []metric.Gauge
}

// MetricRepository — хранилище метрик агента между циклами сбора и отправки.
type MetricRepository interface {
	// SetGauge сохраняет значение gauge, заменяя предыдущее.
	SetGauge(gauge *metric.Gauge) error
	// AddCounter прибавляет значение к накопленному счётчику.
	AddCounter(counter *metric.Counter) error

	// GetMetrics возвращает снимок всех сохранённых метрик.
	GetMetrics() (MetricsState, error)
}
