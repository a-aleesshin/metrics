// Package randomadapter реализует порт генерации случайных значений через math/rand.
package randomadapter

import "math/rand"

// RandomValueAdapter — реализация RandomValueProvider на основе math/rand.
type RandomValueAdapter struct{}

// NewRandomValueAdapter создаёт адаптер случайных значений.
func NewRandomValueAdapter() *RandomValueAdapter {
	return &RandomValueAdapter{}
}

// GenerateFloat64 возвращает псевдослучайное число в диапазоне [0.0, 1.0).
func (a *RandomValueAdapter) GenerateFloat64() float64 {
	return rand.Float64()
}
