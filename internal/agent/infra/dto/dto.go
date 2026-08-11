// Package dto содержит JSON-структуры протокола обмена агента с сервером.
package dto

// MetricsSend — тело запроса обновления метрики в JSON API сервера.
type MetricsSend struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
