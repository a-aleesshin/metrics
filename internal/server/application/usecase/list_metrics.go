package usecase

import (
	"sort"
	"strconv"

	"github.com/a-aleesshin/metrics/internal/server/application/dto"
	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
)

type ListMetricUseCase struct {
	repo repository.MetricQueryRepository
}

func NewListMetricUseCase(repo repository.MetricQueryRepository) *ListMetricUseCase {
	return &ListMetricUseCase{
		repo: repo,
	}
}

func (u *ListMetricUseCase) Execute() (dto.ListMetricsResult, error) {
	gauges, err := u.repo.ListGauges()

	if err != nil {
		return dto.ListMetricsResult{}, err
	}

	counters, err := u.repo.ListCounters()

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
