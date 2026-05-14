package generator

import "github.com/a-aleesshin/metrics/internal/server/domain/metric"

type IDGenerator interface {
	NewID() (metric.ID, error)
}
