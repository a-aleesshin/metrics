// Package memory реализует хранилище метрик в памяти процесса.
package memory

import (
	"context"
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

// MemStorage — потокобезопасное in-memory хранилище метрик;
// реализует порты репозиториев чтения, записи, батча и состояния.
type MemStorage struct {
	mu      sync.Mutex
	gauges  map[string]gaugeRecord
	counter map[string]counterRecord
}

// NewMemStorage создаёт пустое in-memory хранилище метрик.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:  make(map[string]gaugeRecord),
		counter: make(map[string]counterRecord),
	}
}

// GetGaugeByName возвращает gauge по имени или nil, если он не найден.
func (m *MemStorage) GetGaugeByName(ctx context.Context, name metric.Name) (*metric.Gauge, error) {
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

// SaveGauge сохраняет gauge, перезаписывая запись с тем же именем.
func (m *MemStorage) SaveGauge(ctx context.Context, gauge *metric.Gauge) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gauges[gauge.Name().String()] = gaugeRecord{
		ID:    gauge.Id().String(),
		Name:  gauge.Name().String(),
		Value: gauge.Value(),
	}

	return nil
}

// GetCounterByName возвращает counter по имени или nil, если он не найден.
func (m *MemStorage) GetCounterByName(ctx context.Context, name metric.Name) (*metric.Counter, error) {
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

// SaveCounter сохраняет counter, перезаписывая запись с тем же именем.
func (m *MemStorage) SaveCounter(ctx context.Context, counter *metric.Counter) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter[counter.Name().String()] = counterRecord{
		ID:    counter.Id().String(),
		Name:  counter.Name().String(),
		Delta: counter.Delta(),
	}

	return nil
}

// ListCounters возвращает все counter-ы без гарантии порядка.
func (m *MemStorage) ListCounters(ctx context.Context) ([]repository.CounterSnapshot, error) {
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

// ListGauges возвращает все gauge без гарантии порядка.
func (m *MemStorage) ListGauges(ctx context.Context) ([]repository.GaugeSnapshot, error) {
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

// FindGaugeByName возвращает значение gauge и флаг его наличия.
func (m *MemStorage) FindGaugeByName(ctx context.Context, name metric.Name) (value float64, found bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.gauges[name.String()]
	if !ok {
		return 0, false, nil
	}

	return rec.Value, true, nil
}

// FindCounterByName возвращает значение counter и флаг его наличия.
func (m *MemStorage) FindCounterByName(ctx context.Context, name metric.Name) (delta int64, found bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.counter[name.String()]
	if !ok {
		return 0, false, nil
	}

	return rec.Delta, true, nil
}

// GetAllMetrics возвращает полное состояние хранилища как доменные объекты.
func (m *MemStorage) GetAllMetrics(ctx context.Context) (repository.MetricsState, error) {
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

// UpdateBatch применяет батч под одной блокировкой:
// gauge перезаписываются, counter накапливаются.
func (m *MemStorage) UpdateBatch(ctx context.Context, batch repository.MetricBatch) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, gauge := range batch.Gauges {
		m.gauges[gauge.Name().String()] = gaugeRecord{
			ID:    gauge.Id().String(),
			Name:  gauge.Name().String(),
			Value: gauge.Value(),
		}
	}

	for _, counter := range batch.Counters {
		name := counter.Name().String()

		rec, ok := m.counter[name]
		if !ok {
			m.counter[name] = counterRecord{
				ID:    counter.Id().String(),
				Name:  name,
				Delta: counter.Delta(),
			}
			continue
		}

		rec.Delta += counter.Delta()
		m.counter[name] = rec
	}

	return nil
}
