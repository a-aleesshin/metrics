package usecase

import (
	"strconv"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
	"github.com/a-aleesshin/metrics/internal/agent/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/agent/application/port/sender"
)

type ReportMetricsUseCase struct {
	repo   repository.MetricRepository
	sender sender.MetricSender
}

func NewReportMetricsUseCase(repo repository.MetricRepository, sender sender.MetricSender) *ReportMetricsUseCase {
	return &ReportMetricsUseCase{repo: repo, sender: sender}
}

func (usecase *ReportMetricsUseCase) Execute() error {
	metrics, err := usecase.repo.GetMetrics()

	if err != nil {
		return err
	}

	for _, metric := range metrics.Gauges {
		metricDTO := dto.MetricDTO{
			Type:  "gauge",
			Name:  metric.Name().String(),
			Value: strconv.FormatFloat(metric.Value(), 'f', -1, 64),
		}

		err = usecase.sender.Send(metricDTO)

		if err != nil {
			return err
		}
	}

	for _, metric := range metrics.Counters {
		metricDTO := dto.MetricDTO{
			Type:  "counter",
			Name:  metric.Name().String(),
			Value: strconv.FormatInt(metric.Value(), 10),
		}

		err = usecase.sender.Send(metricDTO)

		if err != nil {
			return err
		}
	}

	return nil
}
