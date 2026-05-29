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
	sent        []dto.MetricDTO
	err         error
	sendCalled  bool
	batchCalled bool
}

func (s *metricSenderSpy) Send(metric dto.MetricDTO) error {
	s.sendCalled = true

	if s.err != nil {
		return s.err
	}

	s.sent = append(s.sent, metric)
	return nil
}

func (s *metricSenderSpy) SendBatch(metrics []dto.MetricDTO) error {
	s.batchCalled = true

	if s.err != nil {
		return s.err
	}

	s.sent = append(s.sent, metrics...)
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
		name            string
		repository      *metricRepositoryStub
		sender          *metricSenderSpy
		wantErr         bool
		wantMetrics     []dto.MetricDTO
		wantSendCalled  bool
		wantBatchCalled bool
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
			wantBatchCalled: true,
		},
		{
			name: "repository error",
			repository: &metricRepositoryStub{
				err: errors.New("repository failed"),
			},
			sender:          &metricSenderSpy{},
			wantErr:         true,
			wantBatchCalled: false,
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
			wantErr:         true,
			wantBatchCalled: true,
		},
		{
			name: "empty metrics does not send batch",
			repository: &metricRepositoryStub{
				state: repository.MetricsState{},
			},
			sender:          &metricSenderSpy{},
			wantMetrics:     nil,
			wantBatchCalled: false,
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

			if tt.sender.sendCalled {
				t.Fatal("expected Send not to be called")
			}

			if tt.sender.batchCalled != tt.wantBatchCalled {
				t.Fatalf("expected SendBatch called %v, got %v", tt.wantBatchCalled, tt.sender.batchCalled)
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

func TestReportMetricsUseCase_BuildMetrics(t *testing.T) {
	gaugeName, _ := metric.NewName("Alloc")
	counterName, _ := metric.NewName("PollCount")

	tests := []struct {
		name        string
		repository  *metricRepositoryStub
		wantErr     bool
		wantMetrics []dto.MetricDTO
	}{
		{
			name: "builds all metrics",
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
			name: "returns empty slice for empty repository",
			repository: &metricRepositoryStub{
				state: repository.MetricsState{},
			},
			wantMetrics: []dto.MetricDTO{},
		},
		{
			name: "returns repository error",
			repository: &metricRepositoryStub{
				err: errors.New("repository failed"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			uc := NewReportMetricsUseCase(tt.repository, &metricSenderSpy{})

			// Act
			gotMetrics, err := uc.BuildMetrics()

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

			got := metricsToMap(gotMetrics)
			want := metricsToMap(tt.wantMetrics)

			if len(got) != len(want) {
				t.Fatalf("expected %d metrics, got %d", len(want), len(got))
			}

			for key, wantMetric := range want {
				gotMetric, ok := got[key]
				if !ok {
					t.Fatalf("expected metric %q", key)
				}
				if gotMetric != wantMetric {
					t.Fatalf("metric %q: got %+v, want %+v", key, gotMetric, wantMetric)
				}
			}
		})
	}
}
