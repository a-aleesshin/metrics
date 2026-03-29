package metric

type Gauge struct {
	name  Name
	value float64
}

func NewGauge(name Name, value float64) *Gauge {
	return &Gauge{name: name, value: value}
}

func (g *Gauge) Name() Name {
	return g.name
}

func (g *Gauge) Value() float64 {
	return g.value
}

type Counter struct {
	name  Name
	value int64
}

func NewCounter(name Name, value int64) *Counter {
	return &Counter{name: name, value: value}
}

func (g *Counter) Name() Name {
	return g.name
}

func (g *Counter) Value() int64 {
	return g.value
}
