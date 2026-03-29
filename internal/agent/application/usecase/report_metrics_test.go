package usecase

import (
	"errors"
	"strconv"
	"testing"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
	"github.com/a-aleesshin/metrics/internal/agent/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/agent/domain/metric"
)

type metricSenderSpy struct {
	sent []dto.MetricDTO
	err  error
}

func (s *metricSenderSpy) Send(metric dto.MetricDTO) error {
	if s.err != nil {
		return s.err
	}

	s.sent = append(s.sent, metric)
	return nil
}

type metricRepositoryStub struct {
	state repository.MetricsState
	err   error
}

func (s *metricRepositoryStub) SetGauge(_ *metric.Gauge) error {
	return nil
}

func (s *metricRepositoryStub) AddCounter(_ *metric.Counter) error {
	return nil
}

func (s *metricRepositoryStub) GetMetrics() (repository.MetricsState, error) {
	return s.state, s.err
}

func metricsToMap(metrics []dto.MetricDTO) map[string]dto.MetricDTO {
	out := make(map[string]dto.MetricDTO, len(metrics))
	for _, metricDTO := range metrics {
		key := metricDTO.Type + ":" + metricDTO.Name
		out[key] = metricDTO
	}
	return out
}

func TestReportMetricsUseCase_Execute(t *testing.T) {
	gaugeName, _ := metric.NewName("Alloc")
	counterName, _ := metric.NewName("PollCount")

	tests := []struct {
		name        string
		repository  *metricRepositoryStub
		sender      *metricSenderSpy
		wantErr     bool
		wantMetrics []dto.MetricDTO
	}{
		{
			name: "send all metrics",
			repository: &metricRepositoryStub{
				state: repository.MetricsState{
					Gauges: []metric.Gauge{
						*metric.NewGauge(gaugeName, 123.45),
					},
					Counters: []metric.Counter{
						*metric.NewCounter(counterName, 7),
					},
				},
			},
			sender: &metricSenderSpy{},
			wantMetrics: []dto.MetricDTO{
				{
					Type:  "gauge",
					Name:  "Alloc",
					Value: strconv.FormatFloat(123.45, 'f', -1, 64),
				},
				{
					Type:  "counter",
					Name:  "PollCount",
					Value: strconv.FormatInt(7, 10),
				},
			},
		},
		{
			name: "repository error",
			repository: &metricRepositoryStub{
				err: errors.New("repository failed"),
			},
			sender:  &metricSenderSpy{},
			wantErr: true,
		},
		{
			name: "sender error",
			repository: &metricRepositoryStub{
				state: repository.MetricsState{
					Gauges: []metric.Gauge{
						*metric.NewGauge(gaugeName, 10),
					},
				},
			},
			sender: &metricSenderSpy{
				err: errors.New("sender failed"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			uc := NewReportMetricsUseCase(tt.repository, tt.sender)

			// Act
			err := uc.Execute()

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotMetrics := metricsToMap(tt.sender.sent)
			wantMetrics := metricsToMap(tt.wantMetrics)

			if len(gotMetrics) != len(wantMetrics) {
				t.Fatalf("expected %d sent metrics, got %d", len(wantMetrics), len(gotMetrics))
			}

			for key, wantMetric := range wantMetrics {
				gotMetric, ok := gotMetrics[key]
				if !ok {
					t.Fatalf("expected metric %q to be sent", key)
				}

				if gotMetric != wantMetric {
					t.Fatalf("metric %q: got %+v, want %+v", key, gotMetric, wantMetric)
				}
			}
		})
	}
}
