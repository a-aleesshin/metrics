package reader

type RuntimeMetric struct {
	Name  string
	Value float64
}

type RuntimeReader interface {
	Read() []RuntimeMetric
}
