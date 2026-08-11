package usecase

import (
	"context"
	"sort"
	"strconv"

	"github.com/a-aleesshin/metrics/internal/server/application/dto"
	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
)

// ListMetricUseCase — use case получения списка всех метрик.
type ListMetricUseCase struct {
	repo repository.MetricQueryRepository
}

// NewListMetricUseCase создаёт use case получения списка метрик.
func NewListMetricUseCase(repo repository.MetricQueryRepository) *ListMetricUseCase {
	return &ListMetricUseCase{
		repo: repo,
	}
}

// Execute возвращает все gauge и counter, отсортированные по типу и имени.
func (u *ListMetricUseCase) Execute(ctx context.Context) (dto.ListMetricsResult, error) {
	gauges, err := u.repo.ListGauges(ctx)

	if err != nil {
		return dto.ListMetricsResult{}, err
	}

	counters, err := u.repo.ListCounters(ctx)

	if err != nil {
		return dto.ListMetricsResult{}, err
	}

	items := make([]dto.MetricView, 0, len(counters)+len(gauges))

	for _, gauge := range gauges {
		items = append(items, dto.MetricView{
			Type:  "gauge",
			Name:  gauge.Name,
			Value: strconv.FormatFloat(gauge.Value, 'f', -1, 64),
		})
	}

	for _, counter := range counters {
		items = append(items, dto.MetricView{
			Type:  "counter",
			Name:  counter.Name,
			Value: strconv.FormatInt(counter.Delta, 10),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Type == items[j].Type {
			return items[i].Name < items[j].Name
		}

		return items[i].Type < items[j].Type
	})

	return dto.ListMetricsResult{Items: items}, nil
}
