package repository

import "context"

type MetricSnapshot struct {
	ID    string   `json:"id"`
	Type  string   `json:"type"`
	Value *float64 `json:"value,omitempty"`
	Delta *int64   `json:"delta,omitempty"`
}

type MetricSnapshotStore interface {
	Save(ctx context.Context, metrics []MetricSnapshot) error
	Load(ctx context.Context) ([]MetricSnapshot, error)
}
