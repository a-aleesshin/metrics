package mapper

import (
	"errors"
	"fmt"

	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

type MetricSnapshotMapper struct {
}

func NewMetricSnapshotMapper() *MetricSnapshotMapper {
	return &MetricSnapshotMapper{}
}

func (m *MetricSnapshotMapper) SnapshotToDomain(snapshot repository.MetricSnapshot) (g *metric.Gauge, c *metric.Counter, err error) {
	switch snapshot.Type {
	case TypeGauge:
		if snapshot.Value == nil || snapshot.Delta != nil {
			return nil, nil, errors.New("invalid gauge snapshot: value required, delta must be nil")
		}

		gauge, err := metric.RestoreGauge(snapshot.ID, snapshot.ID, *snapshot.Value)

		if err != nil {
			return nil, nil, fmt.Errorf("restore gauge: %w", err)
		}

		return gauge, nil, nil
	case TypeCounter:
		if snapshot.Delta == nil || snapshot.Value != nil {
			return nil, nil, errors.New("invalid counter snapshot: delta required, value must be nil")
		}

		counter, err := metric.RestoreCounter(snapshot.ID, snapshot.ID, *snapshot.Delta)

		if err != nil {
			return nil, nil, fmt.Errorf("restore counter: %w", err)
		}

		return nil, counter, nil

	default:
		return nil, nil, fmt.Errorf("unknown metric type %q", snapshot.Type)
	}
}

func (m *MetricSnapshotMapper) GaugeToSnapshot(gauge *metric.Gauge) (repository.MetricSnapshot, error) {
	if gauge == nil {
		return repository.MetricSnapshot{}, errors.New("gauge is nil")
	}

	v := gauge.Value()

	return repository.MetricSnapshot{
		ID:    gauge.Name().String(),
		Type:  TypeGauge,
		Value: &v,
		Delta: nil,
	}, nil
}

func (m *MetricSnapshotMapper) CounterToSnapshot(counter *metric.Counter) (repository.MetricSnapshot, error) {
	if counter == nil {
		return repository.MetricSnapshot{}, errors.New("counter is nil")
	}

	d := counter.Delta()

	return repository.MetricSnapshot{
		ID:    counter.Name().String(),
		Type:  TypeCounter,
		Value: nil,
		Delta: &d,
	}, nil
}
