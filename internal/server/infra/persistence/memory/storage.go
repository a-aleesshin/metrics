package memory

import (
	"sync"

	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
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
	mu      sync.Mutex
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
	m.mu.Lock()
	record, ok := m.gauges[string(name)]
	defer m.mu.Unlock()

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
	m.mu.Lock()
	record, ok := m.counter[string(name)]
	defer m.mu.Unlock()

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

func (m *MemStorage) ListCounters() ([]repository.CounterSnapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]repository.CounterSnapshot, 0, len(m.counter))

	for _, cointer := range m.counter {
		out = append(out, repository.CounterSnapshot{
			Name:  cointer.Name,
			Delta: cointer.Delta,
		})
	}

	return out, nil
}

func (m *MemStorage) ListGauges() ([]repository.GaugeSnapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]repository.GaugeSnapshot, 0, len(m.gauges))

	for _, gauge := range m.gauges {
		out = append(out, repository.GaugeSnapshot{
			Name:  gauge.Name,
			Value: gauge.Value,
		})
	}

	return out, nil
}

func (m *MemStorage) FindGaugeByName(name metric.Name) (value float64, found bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.gauges[name.String()]
	if !ok {
		return 0, false, nil
	}

	return rec.Value, true, nil
}

func (m *MemStorage) FindCounterByName(name metric.Name) (delta int64, found bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.counter[name.String()]
	if !ok {
		return 0, false, nil
	}

	return rec.Delta, true, nil
}

func (m *MemStorage) GetAllMetrics() (repository.MetricsState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := repository.MetricsState{
		Gauges:   make([]*metric.Gauge, 0, len(m.gauges)),
		Counters: make([]*metric.Counter, 0, len(m.counter)),
	}

	for _, rec := range m.gauges {
		g, err := metric.RestoreGauge(rec.ID, rec.Name, rec.Value)

		if err != nil {
			return repository.MetricsState{}, err
		}

		state.Gauges = append(state.Gauges, g)
	}

	for _, rec := range m.counter {
		c, err := metric.RestoreCounter(rec.ID, rec.Name, rec.Delta)

		if err != nil {
			return repository.MetricsState{}, err
		}

		state.Counters = append(state.Counters, c)
	}

	return state, nil
}
