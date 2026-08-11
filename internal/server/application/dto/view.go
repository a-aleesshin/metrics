// Package dto содержит структуры передачи данных между
// use case-ами и слоем представления.
package dto

// MetricView — представление одной метрики для вывода:
// значение уже отформатировано строкой.
type MetricView struct {
	Type  string
	Name  string
	Value string
}

// NewMetricView создаёт представление метрики.
func NewMetricView(t string, name string, value string) *MetricView {
	return &MetricView{
		Type:  t,
		Name:  name,
		Value: value,
	}
}
