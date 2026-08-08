// Package generator описывает порт генерации идентификаторов метрик.
package generator

import "github.com/a-aleesshin/metrics/internal/server/domain/metric"

// IDGenerator — порт генерации уникальных идентификаторов метрик.
type IDGenerator interface {
	NewID() (metric.ID, error)
}
