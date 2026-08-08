// Package generator определяет порт генерации случайных значений.
package generator

// RandomValueProvider — источник случайных значений для метрики RandomValue.
type RandomValueProvider interface {
	// GenerateFloat64 возвращает случайное число float64.
	GenerateFloat64() float64
}
