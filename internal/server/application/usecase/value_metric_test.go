package usecase

import (
	"errors"
	"testing"

	applicationerror "github.com/a-aleesshin/metrics/internal/server/application/error"
	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type metricQueryRepoStub struct {
	gaugeValue   float64
	gaugeFound   bool
	gaugeErr     error
	counterValue int64
	counterFound bool
	counterErr   error
}

func (s *metricQueryRepoStub) ListGauges() ([]repository.GaugeSnapshot, error) {
	return nil, nil
}

func (s *metricQueryRepoStub) ListCounters() ([]repository.CounterSnapshot, error) {
	return nil, nil
}

func (s *metricQueryRepoStub) FindGaugeByName(name metric.Name) (float64, bool, error) {
	if s.gaugeErr != nil {
		return 0, false, s.gaugeErr
	}
	return s.gaugeValue, s.gaugeFound, nil
}

func (s *metricQueryRepoStub) FindCounterByName(name metric.Name) (int64, bool, error) {
	if s.counterErr != nil {
		return 0, false, s.counterErr
	}
	return s.counterValue, s.counterFound, nil
}

func TestGetValueMetricUseCase_Execute(t *testing.T) {
	tests := []struct {
		name      string
		cmd       ValueMetricCommand
		repo      *metricQueryRepoStub
		wantValue string
		wantErr   error
	}{
		{
			name: "user_case_get_value_gauge_found",
			cmd:  ValueMetricCommand{Type: "gauge", Name: "Alloc"},
			repo: &metricQueryRepoStub{
				gaugeValue: 123.45,
				gaugeFound: true,
			},
			wantValue: "123.45",
		},
		{
			name: "user_case_get_value_counter_found",
			cmd:  ValueMetricCommand{Type: "counter", Name: "PollCount"},
			repo: &metricQueryRepoStub{
				counterValue: 7,
				counterFound: true,
			},
			wantValue: "7",
		},
		{
			name: "user_case_get_value_gauge_not_found",
			cmd:  ValueMetricCommand{Type: "gauge", Name: "Missing"},
			repo: &metricQueryRepoStub{
				gaugeFound: false,
			},
			wantErr: applicationerror.ErrMetricNotFound,
		},
		{
			name: "user_case_get_value_counter_not_found",
			cmd:  ValueMetricCommand{Type: "counter", Name: "Missing"},
			repo: &metricQueryRepoStub{
				counterFound: false,
			},
			wantErr: applicationerror.ErrMetricNotFound,
		},
		{
			name:    "user_case_get_value_unsupported_type",
			cmd:     ValueMetricCommand{Type: "hist", Name: "Any"},
			repo:    &metricQueryRepoStub{},
			wantErr: metric.ErrUnsupportedMetricType,
		},
		{
			name: "user_case_get_value_repo_gauge_error",
			cmd:  ValueMetricCommand{Type: "gauge", Name: "Alloc"},
			repo: &metricQueryRepoStub{
				gaugeErr: errors.New("repo fail"),
			},
			wantErr: errors.New("repo fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			uc := NewGetValueMetricUseCase(tt.repo)

			// Act
			got, err := uc.Execute(tt.cmd)

			// Assert
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.wantValue {
				t.Fatalf("expected value %q, got %q", tt.wantValue, got)
			}
		})
	}
}
