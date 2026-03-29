package memory

import (
	"sync"

	"github.com/a-aleesshin/metrics/internal/agent/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/agent/domain/metric"
)

type MemMetricRepository struct {
	mu       sync.RWMutex
	gauges   map[metric.Name]float64
	counters map[metric.Name]int64
}

func NewMemMetricRepository() *MemMetricRepository {
	return &MemMetricRepository{
		gauges:   make(map[metric.Name]float64),
		counters: make(map[metric.Name]int64),
	}
}

func (m *MemMetricRepository) SetGauge(gauge *metric.Gauge) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gauges[gauge.Name()] = gauge.Value()
	return nil
}

func (m *MemMetricRepository) AddCounter(counter *metric.Counter) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters[counter.Name()] += counter.Value()
	return nil
}

func (m *MemMetricRepository) GetMetrics() (repository.MetricsState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := repository.MetricsState{
		Counters: make([]metric.Counter, 0, len(m.counters)),
		Gauges:   make([]metric.Gauge, 0, len(m.gauges)),
	}

	for name, value := range m.gauges {
		out.Gauges = append(out.Gauges, *metric.NewGauge(name, value))
	}

	for name, value := range m.counters {
		out.Counters = append(out.Counters, *metric.NewCounter(name, value))
	}

	return out, nil
}
