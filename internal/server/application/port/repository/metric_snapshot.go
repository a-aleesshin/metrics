package repository

import "context"

// MetricSnapshot — сериализуемое представление метрики:
// для gauge заполнен Value, для counter — Delta.
type MetricSnapshot struct {
	ID    string   `json:"id"`
	Type  string   `json:"type"`
	Value *float64 `json:"value,omitempty"`
	Delta *int64   `json:"delta,omitempty"`
}

// MetricSnapshotStore — порт сохранения и загрузки набора snapshot-ов метрик.
type MetricSnapshotStore interface {
	Save(ctx context.Context, metrics []MetricSnapshot) error
	Load(ctx context.Context) ([]MetricSnapshot, error)
}
