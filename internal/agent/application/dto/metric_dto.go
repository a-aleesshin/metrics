// Package dto содержит транспортно-независимые DTO прикладного слоя агента.
package dto

// MetricDTO метрика в строковом представлении для передачи между слоями агента.
type MetricDTO struct {
	Type  string
	Name  string
	Value string
}
