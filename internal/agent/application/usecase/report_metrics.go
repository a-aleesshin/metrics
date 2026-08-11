package usecase

import (
	"strconv"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
	"github.com/a-aleesshin/metrics/internal/agent/application/port/repository"
)

// MetricSender отправляет метрики на сервер поштучно или батчем.
type MetricSender interface {
	// Send отправляет одну метрику.
	Send(dto dto.MetricDTO) error
	// SendBatch отправляет пачку метрик одним запросом.
	SendBatch(metrics []dto.MetricDTO) error
}

// ReportMetricsUseCase формирует отчёт из репозитория и отправляет его через MetricSender.
type ReportMetricsUseCase struct {
	repo   repository.MetricRepository
	sender MetricSender
}

// NewReportMetricsUseCase создаёт сценарий отправки метрик.
func NewReportMetricsUseCase(repo repository.MetricRepository, sender MetricSender) *ReportMetricsUseCase {
	return &ReportMetricsUseCase{repo: repo, sender: sender}
}

// BuildMetrics читает снимок метрик из репозитория и конвертирует его в список DTO.
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

// Execute строит отчёт и отправляет его на сервер.
func (usecase *ReportMetricsUseCase) Execute() error {
	batch, err := usecase.BuildMetrics()
	if err != nil {
		return err
	}

	return usecase.SendMetrics(batch)
}

// SendMetrics отправляет готовый батч; пустой батч не отправляется.
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
