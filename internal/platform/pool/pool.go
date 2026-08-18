// Package pool предоставляет типобезопасную обёртку над sync.Pool для
// повторного использования «тяжёлых» объектов. Generic-параметр ограничен
// типами с методом Reset() (например, сгенерированным утилитой cmd/reset),
// поэтому объект всегда возвращается в пул со сброшенным состоянием.
package pool

import "sync"

// Resettable — ограничение generic-параметра Pool: тип должен уметь
// сбрасывать своё состояние к начальным значениям.
type Resettable interface {
	Reset()
}

// Pool — типобезопасный пул объектов типа T поверх sync.Pool.
// Put сбрасывает объект методом Reset() перед возвратом в пул,
// поэтому Get всегда отдаёт объект в начальном состоянии.
type Pool[T Resettable] struct {
	pool sync.Pool
}

// New создаёт пул для типа T. Фабрика опциональна, если factory задана,
// она вызывается, когда пул пуст и нужен новый объект, при factory = nil
// пустой пул возвращает из Get нулевое значение типа T.
func New[T Resettable](factory func() T) *Pool[T] {
	p := &Pool[T]{}

	if factory != nil {
		p.pool.New = func() any { return factory() }
	}

	return p
}

// Get возвращает объект из пула, если пул пуст то новый объект от factory,
// а при не заданной factory, нулевое значение типа T.
func (p *Pool[T]) Get() T {
	value, ok := p.pool.Get().(T)
	if !ok {
		var zero T
		return zero
	}

	return value
}

// Put сбрасывает состояние объекта вызовом Reset() и возвращает его в пул.
// После Put объект использовать нельзя — им владеет пул.
func (p *Pool[T]) Put(value T) {
	value.Reset()
	p.pool.Put(value)
}
