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

// Pool типобезопасный пул объектов типа T поверх sync.Pool.
// Put сбрасывает объект методом Reset() перед возвратом в пул,
// поэтому Get всегда отдаёт объект в начальном состоянии.
type Pool[T Resettable] struct {
	pool sync.Pool
}

// New создаёт пул для типа T, factory вызывается, когда пул пуст
// и нужен новый объект.
func New[T Resettable](factory func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any { return factory() },
		},
	}
}

// Get возвращает объект из пула или, если пул пуст, создаёт новый через factory.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put сбрасывает состояние объекта вызовом Reset() и возвращает его в пул.
func (p *Pool[T]) Put(value T) {
	value.Reset()
	p.pool.Put(value)
}
