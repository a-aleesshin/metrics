package metric

type Gauge struct {
	id   ID
	name Name
	val  float64
}

func NewGauge(id string, name string, value float64) (*Gauge, error) {
	metricName, err := NewName(name)
	if err != nil {
		return nil, err
	}

	metricID, err := NewID(id)

	if err != nil {
		return nil, err
	}

	return &Gauge{id: metricID, name: metricName, val: value}, nil
}

func RestoreGauge(id string, name string, value float64) (*Gauge, error) {
	return NewGauge(id, name, value)
}

func (g *Gauge) Rename(name string) error {
	metricName, err := NewName(name)

	if err != nil {
		return err
	}

	g.name = metricName

	return nil
}

func (g *Gauge) UpdateValue(value float64) {
	g.val = value
}

func (g *Gauge) Id() ID {
	return g.id
}

func (g *Gauge) Name() Name {
	return g.name
}

func (g *Gauge) Value() float64 {
	return g.val
}

type Counter struct {
	id    ID
	name  Name
	delta int64
}

func NewCounter(id, name string, delta int64) (*Counter, error) {
	metricName, err := NewName(name)

	if err != nil {
		return nil, err
	}

	metricID, err := NewID(id)

	if err != nil {
		return nil, err
	}

	return &Counter{id: metricID, name: metricName, delta: delta}, nil
}

func RestoreCounter(id string, name string, delta int64) (*Counter, error) {
	return NewCounter(id, name, delta)
}

func (c *Counter) Rename(name string) error {
	if name == "" {
		return ErrNameEmpty
	}

	c.name = Name(name)

	return nil
}

func (c *Counter) Add(delta int64) {
	c.delta += delta
}

func (c *Counter) Id() ID {
	return c.id
}

func (c *Counter) Name() Name {
	return c.name
}

func (c *Counter) Delta() int64 {
	return c.delta
}
