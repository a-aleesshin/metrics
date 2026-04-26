package usecase

import (
	"math"

	"github.com/a-aleesshin/metrics/internal/agent/application/port/generator"
	"github.com/a-aleesshin/metrics/internal/agent/application/port/reader"
	"github.com/a-aleesshin/metrics/internal/agent/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/agent/domain/metric"
)

type CollectMetricsUseCase struct {
	runtimeRider reader.RuntimeReader
	repository   repository.MetricRepository
	randomValue  generator.RandomValueProvider
}

func NewCollectMetricsUseCase(runtimeRider reader.RuntimeReader, repository repository.MetricRepository, randomValue generator.RandomValueProvider) *CollectMetricsUseCase {
	return &CollectMetricsUseCase{
		runtimeRider: runtimeRider,
		repository:   repository,
		randomValue:  randomValue,
	}
}

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
