// Package repository описывает порты хранения метрик:
// репозитории чтения/записи, батч-обновление, состояние и snapshot-хранилище.
package repository

import (
	"context"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

// MetricRepository — порт чтения и сохранения отдельных метрик;
// Get-методы возвращают nil без ошибки, если метрика не найдена.
type MetricRepository interface {
	GetGaugeByName(ctx context.Context, name metric.Name) (*metric.Gauge, error)
	GetCounterByName(ctx context.Context, name metric.Name) (*metric.Counter, error)

	SaveGauge(ctx context.Context, gauge *metric.Gauge) error
	SaveCounter(ctx context.Context, counter *metric.Counter) error
}
