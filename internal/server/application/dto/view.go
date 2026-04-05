package dto

type MetricView struct {
	Type  string
	Name  string
	Value string
}

func NewMetricView(t string, name string, value string) *MetricView {
	return &MetricView{
		Type:  t,
		Name:  name,
		Value: value,
	}
}
