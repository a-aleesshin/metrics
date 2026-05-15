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

func (usecase *ReportMetricsUseCase) Execute() error {
	metrics, err := usecase.repo.GetMetrics()

	if err != nil {
		return err
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

	if len(batch) == 0 {
		return nil
	}

	err = usecase.sender.SendBatch(batch)

	if err != nil {
		return err
	}

	return nil
}
