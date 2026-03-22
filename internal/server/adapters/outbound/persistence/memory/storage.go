package memory

import (
	"sync"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type gaugeRecord struct {
	ID    string
	Name  string
	Value float64
}

type counterRecord struct {
	ID    string
	Name  string
	Delta int64
}

type MemStorage struct {
	mu      sync.RWMutex
	gauges  map[string]gaugeRecord
	counter map[string]counterRecord
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:  make(map[string]gaugeRecord),
		counter: make(map[string]counterRecord),
	}
}

func (m *MemStorage) GetGaugeByName(name metric.Name) (*metric.Gauge, error) {
	m.mu.RLock()
	record, ok := m.gauges[string(name)]
	m.mu.RUnlock()

	if !ok {
		return nil, nil
	}

	gauge, err := metric.RestoreGauge(record.ID, record.Name, record.Value)

	if err != nil {
		return nil, err
	}

	return gauge, nil
}

func (m *MemStorage) SaveGauge(gauge *metric.Gauge) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gauges[gauge.Name().String()] = gaugeRecord{
		ID:    gauge.Id().String(),
		Name:  gauge.Name().String(),
		Value: gauge.Value(),
	}

	return nil
}

func (m *MemStorage) GetCounterByName(name metric.Name) (*metric.Counter, error) {
	m.mu.RLock()
	record, ok := m.counter[string(name)]
	m.mu.RUnlock()

	if !ok {
		return nil, nil
	}

	counter, err := metric.RestoreCounter(record.ID, record.Name, record.Delta)

	if err != nil {
		return nil, err
	}

	return counter, nil
}

func (m *MemStorage) SaveCounter(counter *metric.Counter) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter[counter.Name().String()] = counterRecord{
		ID:    counter.Id().String(),
		Name:  counter.Name().String(),
		Delta: counter.Delta(),
	}

	return nil
}
