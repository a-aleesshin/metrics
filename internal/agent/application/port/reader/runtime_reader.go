// Package reader определяет порты чтения runtime и системных метрик.
package reader

// RuntimeMetric — метрика Go runtime: имя и значение gauge.
type RuntimeMetric struct {
	Name  string
	Value float64
}

// SystemMetric — системная метрика (память, CPU): имя и значение gauge.
type SystemMetric struct {
	Name  string
	Value float64
}

// RuntimeReader читает набор метрик Go runtime.
type RuntimeReader interface {
	// Read возвращает текущие значения runtime-метрик.
	Read() []RuntimeMetric
}

// SystemReader читает набор системных метрик.
type SystemReader interface {
	// Read возвращает текущие значения системных метрик.
	Read() ([]SystemMetric, error)
}
