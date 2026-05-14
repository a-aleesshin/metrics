package usecase

import (
	"context"

	"github.com/a-aleesshin/metrics/internal/server/application/port/generator"
	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type MetricUpdatesCommand struct {
	Name  string
	MType string
	Delta *int64
	Value *float64
}

type UpdatesMetricsCommand struct {
	Metrics []MetricUpdatesCommand
}

type UpdatesMetricsUseCase struct {
	idGenerator generator.IDGenerator
	repository  repository.MetricBatchRepository
}

func NewUpdatesMetricsUseCase(
	repository repository.MetricBatchRepository,
	idGenerator generator.IDGenerator,
) *UpdatesMetricsUseCase {
	return &UpdatesMetricsUseCase{
		repository:  repository,
		idGenerator: idGenerator,
	}
}

func (uc *UpdatesMetricsUseCase) Execute(ctx context.Context, command UpdatesMetricsCommand) error {
	batch := repository.MetricBatch{
		Gauges:   make([]*metric.Gauge, 0, len(command.Metrics)),
		Counters: make([]*metric.Counter, 0, len(command.Metrics)),
	}

	for _, metricItem := range command.Metrics {
		id, err := uc.idGenerator.NewID()

		if err != nil {
			return err
		}

		name, err := metric.NewName(metricItem.Name)
		if err != nil {
			return err
		}

		switch metricItem.MType {
		case "gauge":
			if metricItem.Value == nil {
				return metric.ErrInvalidMetricValue
			}

			gauge, err := metric.NewGauge(id.String(), name.String(), *metricItem.Value)

			if err != nil {
				return err
			}

			batch.Gauges = append(batch.Gauges, gauge)
		case "counter":
			if metricItem.Delta == nil {
				return metric.ErrInvalidMetricValue
			}
			
			counter, err := metric.NewCounter(id.String(), name.String(), *metricItem.Delta)

			if err != nil {
				return err
			}

			batch.Counters = append(batch.Counters, counter)

		default:
			return metric.ErrUnsupportedMetricType
		}
	}

	if len(batch.Counters) == 0 && len(batch.Gauges) == 0 {
		return nil
	}

	return uc.repository.UpdateBatch(ctx, batch)
}
