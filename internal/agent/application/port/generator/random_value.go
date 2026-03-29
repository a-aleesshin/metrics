package generator

type RandomValueProvider interface {
	GenerateFloat64() float64
}
