package usecase

import (
	"strconv"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
	"github.com/a-aleesshin/metrics/internal/agent/application/port/repository"
)

type MetricSender interface {
	Send(dto dto.MetricDTO) error
	SendBatch(metrics []dto.MetricDTO) error
}

type ReportMetricsUseCase struct {
	repo   repository.MetricRepository
	sender MetricSender
}

func NewReportMetricsUseCase(repo repository.MetricRepository, sender MetricSender) *ReportMetricsUseCase {
	return &ReportMetricsUseCase{repo: repo, sender: sender}
}

func (usecase *ReportMetricsUseCase) BuildMetrics() ([]dto.MetricDTO, error) {
	metrics, err := usecase.repo.GetMetrics()

	if err != nil {
		return nil, err
	}

	batch := make([]dto.MetricDTO, 0, len(metrics.Gauges)+len(metrics.Counters))

	for _, metric := range metrics.Gauges {
		metricDTO := dto.MetricDTO{
			Type:  "gauge",
			Name:  metric.Name().String(),
			Value: strconv.FormatFloat(metric.Value(), 'f', -1, 64),
		}

		batch = append(batch, metricDTO)
	}

	for _, metric := range metrics.Counters {
		metricDTO := dto.MetricDTO{
			Type:  "counter",
			Name:  metric.Name().String(),
			Value: strconv.FormatInt(metric.Value(), 10),
		}

		batch = append(batch, metricDTO)
	}

	return batch, nil
}

func (usecase *ReportMetricsUseCase) Execute() error {
	batch, err := usecase.BuildMetrics()
	if err != nil {
		return err
	}

	return usecase.SendMetrics(batch)
}

func (usecase *ReportMetricsUseCase) SendMetrics(batch []dto.MetricDTO) error {
	if len(batch) == 0 {
		return nil
	}

	err := usecase.sender.SendBatch(batch)

	if err != nil {
		return err
	}

	return nil
}
