package dto

type ListMetricsResult struct {
	Items []MetricView
}

func NewListMetricsResult(items []MetricView) *ListMetricsResult {
	return &ListMetricsResult{Items: items}
}
