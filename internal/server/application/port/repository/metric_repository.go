package repository

import (
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type MetricRepository interface {
	GetGaugeByName(name metric.Name) (*metric.Gauge, error)
	GetCounterByName(name metric.Name) (*metric.Counter, error)

	SaveGauge(gauge *metric.Gauge) error
	SaveCounter(counter *metric.Counter) error
}
