// Package error содержит ошибки прикладного слоя сервиса метрик.
package error

import "errors"

var (
	// ErrMetricNotFound — запрошенная метрика отсутствует в хранилище.
	ErrMetricNotFound = errors.New("metric not found")
)
