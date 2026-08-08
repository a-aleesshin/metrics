// Package usecase содержит сценарии агента: сбор runtime- и системных метрик и их отправку.
package usecase

import (
	"math"

	"github.com/a-aleesshin/metrics/internal/agent/application/port/generator"
	"github.com/a-aleesshin/metrics/internal/agent/application/port/reader"
	"github.com/a-aleesshin/metrics/internal/agent/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/agent/domain/metric"
)

// CollectMetricsUseCase собирает runtime-метрики, RandomValue и PollCount в репозиторий.
type CollectMetricsUseCase struct {
	runtimeRider reader.RuntimeReader
	repository   repository.MetricRepository
	randomValue  generator.RandomValueProvider
}

// NewCollectMetricsUseCase создаёт сценарий сбора runtime-метрик.
func NewCollectMetricsUseCase(runtimeRider reader.RuntimeReader, repository repository.MetricRepository, randomValue generator.RandomValueProvider) *CollectMetricsUseCase {
	return &CollectMetricsUseCase{
		runtimeRider: runtimeRider,
		repository:   repository,
		randomValue:  randomValue,
	}
}

// Execute читает runtime-метрики и сохраняет их как gauge; NaN/Inf заменяются нулём,
// PollCount увеличивается на 1.
func (usecase *CollectMetricsUseCase) Execute() error {
	metrics := usecase.runtimeRider.Read()

	for _, metricReader := range metrics {
		name, err := metric.NewName(metricReader.Name)

		if err != nil {
			return err
		}

		value := metricReader.Value

		if math.IsNaN(value) || math.IsInf(value, 0) {
			value = 0
		}

		err = usecase.repository.SetGauge(metric.NewGauge(name, value))

		if err != nil {
			return err
		}
	}

	randomValueName, err := metric.NewName("RandomValue")

	if err != nil {
		return err
	}

	rv := usecase.randomValue.GenerateFloat64()
	if math.IsNaN(rv) || math.IsInf(rv, 0) {
		rv = 0
	}

	err = usecase.repository.SetGauge(metric.NewGauge(randomValueName, rv))

	if err != nil {
		return err
	}

	pollCountName, err := metric.NewName("PollCount")

	if err != nil {
		return err
	}

	return usecase.repository.AddCounter(metric.NewCounter(pollCountName, 1))
}
