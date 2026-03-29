package randomadapter

import "math/rand"

type RandomValueAdapter struct{}

func NewRandomValueAdapter() *RandomValueAdapter {
	return &RandomValueAdapter{}
}

func (a *RandomValueAdapter) GenerateFloat64() float64 {
	return rand.Float64()
}
