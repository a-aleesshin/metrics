package usecase

import (
	"math"

	"github.com/a-aleesshin/metrics/internal/agent/application/port/reader"
	"github.com/a-aleesshin/metrics/internal/agent/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/agent/domain/metric"
)

// CollectSystemMetricsUseCase собирает системные метрики (память, CPU) в репозиторий.
type CollectSystemMetricsUseCase struct {
	systemReader reader.SystemReader
	repository   repository.MetricRepository
}

// NewCollectSystemMetricsUseCase создаёт сценарий сбора системных метрик.
func NewCollectSystemMetricsUseCase(systemReader reader.SystemReader, repository repository.MetricRepository) *CollectSystemMetricsUseCase {
	return &CollectSystemMetricsUseCase{
		systemReader: systemReader,
		repository:   repository,
	}
}

// Execute читает системные метрики и сохраняет их как gauge; NaN/Inf заменяются нулём.
func (usecase *CollectSystemMetricsUseCase) Execute() error {
	metrics, err := usecase.systemReader.Read()
	if err != nil {
		return err
	}

	for _, metricReader := range metrics {
		name, err := metric.NewName(metricReader.Name)

		if err != nil {
			return err
		}

		value := metricReader.Value

		if math.IsNaN(value) || math.IsInf(value, 0) {
			value = 0
		}

		if err := usecase.repository.SetGauge(metric.NewGauge(name, value)); err != nil {
			return err
		}
	}

	return nil
}
