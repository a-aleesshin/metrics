// Package metric содержит доменную модель метрик: сущности Gauge и Counter,
// value object-ы ID и Name и доменные ошибки.
package metric

// Gauge — метрика типа gauge: хранит последнее значение float64, перезаписывается.
type Gauge struct {
	id   ID
	name Name
	val  float64
}

// NewGauge создаёт gauge, валидируя id и имя.
func NewGauge(id string, name string, value float64) (*Gauge, error) {
	metricID, err := NewID(id)

	if err != nil {
		return nil, err
	}

	metricName, err := NewName(name)
	if err != nil {
		return nil, err
	}

	return &Gauge{id: metricID, name: metricName, val: value}, nil
}

// RestoreGauge восстанавливает gauge из сохранённого состояния.
func RestoreGauge(id string, name string, value float64) (*Gauge, error) {
	return NewGauge(id, name, value)
}

// Rename меняет имя gauge; пустое имя — ошибка.
func (g *Gauge) Rename(name string) error {
	metricName, err := NewName(name)

	if err != nil {
		return err
	}

	g.name = metricName

	return nil
}

// UpdateValue перезаписывает значение gauge.
func (g *Gauge) UpdateValue(value float64) {
	g.val = value
}

// Id возвращает идентификатор gauge.
func (g *Gauge) ID() ID {
	return g.id
}

// Name возвращает имя gauge.
func (g *Gauge) Name() Name {
	return g.name
}

// Value возвращает текущее значение gauge.
func (g *Gauge) Value() float64 {
	return g.val
}

// Counter — метрика типа counter: накапливает int64-дельты.
type Counter struct {
	id    ID
	name  Name
	delta int64
}

// NewCounter создаёт counter, валидируя id и имя.
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

// RestoreCounter восстанавливает counter из сохранённого состояния.
func RestoreCounter(id string, name string, delta int64) (*Counter, error) {
	return NewCounter(id, name, delta)
}

// Rename меняет имя counter; пустое имя — ошибка.
func (c *Counter) Rename(name string) error {
	if name == "" {
		return ErrNameEmpty
	}

	c.name = Name(name)

	return nil
}

// Add прибавляет дельту к накопленному значению counter.
func (c *Counter) Add(delta int64) {
	c.delta += delta
}

// Id возвращает идентификатор counter.
func (c *Counter) ID() ID {
	return c.id
}

// Name возвращает имя counter.
func (c *Counter) Name() Name {
	return c.name
}

// Delta возвращает накопленное значение counter.
func (c *Counter) Delta() int64 {
	return c.delta
}
