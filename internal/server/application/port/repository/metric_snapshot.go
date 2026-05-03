package repository

type MetricSnapshot struct {
	ID    string   `json:"id"`
	Type  string   `json:"type"`
	Value *float64 `json:"value,omitempty"`
	Delta *int64   `json:"delta,omitempty"`
}

type MetricSnapshotStore interface {
	Save(metrics []MetricSnapshot) error
	Load() ([]MetricSnapshot, error)
}
