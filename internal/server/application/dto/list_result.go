package dto

// ListMetricsResult — результат use case списка метрик.
type ListMetricsResult struct {
	Items []MetricView
}

// NewListMetricsResult создаёт результат списка метрик.
func NewListMetricsResult(items []MetricView) *ListMetricsResult {
	return &ListMetricsResult{Items: items}
}
