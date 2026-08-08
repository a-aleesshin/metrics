// Package metric содержит доменную модель метрик агента: gauge, counter и их имена.
package metric

// Gauge — метрика типа gauge: имя и текущее значение float64.
type Gauge struct {
	name  Name
	value float64
}

// NewGauge создаёт gauge с заданными именем и значением.
func NewGauge(name Name, value float64) *Gauge {
	return &Gauge{name: name, value: value}
}

// Name возвращает имя метрики.
func (g *Gauge) Name() Name {
	return g.name
}

// Value возвращает значение метрики.
func (g *Gauge) Value() float64 {
	return g.value
}

// Counter — метрика типа counter: имя и накапливаемое значение int64.
type Counter struct {
	name  Name
	value int64
}

// NewCounter создаёт counter с заданными именем и значением.
func NewCounter(name Name, value int64) *Counter {
	return &Counter{name: name, value: value}
}

// Name возвращает имя метрики.
func (g *Counter) Name() Name {
	return g.name
}

// Value возвращает значение метрики.
func (g *Counter) Value() int64 {
	return g.value
}
