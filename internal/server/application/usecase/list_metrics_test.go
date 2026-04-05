package usecase

import (
	"errors"
	"testing"

	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type metricQueryRepositoryStub struct {
	gauges          []repository.GaugeSnapshot
	counters        []repository.CounterSnapshot
	listGaugesErr   error
	listCountersErr error
}

func (m metricQueryRepositoryStub) ListGauges() ([]repository.GaugeSnapshot, error) {
	if m.listGaugesErr != nil {
		return nil, m.listGaugesErr
	}

	return m.gauges, nil
}

func (m metricQueryRepositoryStub) ListCounters() ([]repository.CounterSnapshot, error) {
	if m.listCountersErr != nil {
		return nil, m.listCountersErr
	}

	return m.counters, nil
}

func (m metricQueryRepositoryStub) FindGaugeByName(name metric.Name) (value float64, found bool, err error) {
	return 0, false, nil
}

func (m metricQueryRepositoryStub) FindCounterByName(name metric.Name) (delta int64, found bool, err error) {
	return 0, false, nil
}

func TestListMetrics_Execute(t *testing.T) {
	tests := []struct {
		name       string
		repo       *metricQueryRepositoryStub
		wantErr    bool
		wantItems  int
		wantFirst  string
		wantSecond string
	}{
		{
			name: "use case test 1",
			repo: &metricQueryRepositoryStub{
				gauges: []repository.GaugeSnapshot{
					{
						Name:  "HeapAlloc",
						Value: 123.45,
					},
				},
				counters: []repository.CounterSnapshot{
					{
						Name:  "PollCount",
						Delta: 7,
					},
				},
			},
			wantErr:    false,
			wantItems:  2,
			wantFirst:  "counter:PollCount=7",
			wantSecond: "gauge:HeapAlloc=123.45",
		},
		{
			name: "use case list gauges error",
			repo: &metricQueryRepositoryStub{
				listCountersErr: errors.New("gauges failed"),
			},
			wantErr: true,
		},
		{
			name: "use case list counters error",
			repo: &metricQueryRepositoryStub{
				listGaugesErr: errors.New("counters failed"),
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Arrange
			uc := NewListMetricUseCase(test.repo)

			// Act
			items, err := uc.Execute()

			// Assert
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(items.Items) != test.wantItems {
				t.Fatalf("expected %d items, got %d", test.wantItems, len(items.Items))
			}

			if test.wantItems >= 1 {
				first := items.Items[0].Type + ":" + items.Items[0].Name + "=" + items.Items[0].Value

				if first != test.wantFirst {
					t.Fatalf("expected %q, got %q", test.wantFirst, first)

				}
			}

			if test.wantItems >= 2 {
				second := items.Items[test.wantItems-1].Type + ":" + items.Items[test.wantItems-1].Name + "=" + items.Items[test.wantItems-1].Value
				if second != test.wantSecond {
					t.Fatalf("expected %q, got %q", test.wantSecond, second)
				}
			}
		})
	}
}
