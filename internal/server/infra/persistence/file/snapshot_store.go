package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
)

type SnapshotStore struct {
	path string
	mu   sync.Mutex
}

func NewSnapshotStore(path string) (*SnapshotStore, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("snapshot path is empty")
	}

	return &SnapshotStore{
		path: path,
	}, nil
}

func (ss *SnapshotStore) Save(metrics []repository.MetricSnapshot) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if metrics == nil {
		metrics = make([]repository.MetricSnapshot, 0)
	}

	for i, m := range metrics {
		if err := validateMetricSnapshot(m); err != nil {
			return fmt.Errorf("invalid snapshot at index %d: %w\\", i, err)
		}
	}

	dir := filepath.Dir(ss.path)

	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	tmp, err := os.CreateTemp(dir, ".metrics-*.tmp")

	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	tmpName := tmp.Name()

	enc := json.NewEncoder(tmp)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(metrics); err != nil {
		return fmt.Errorf("failed to encode metrics: %w", err)
	}

	if err = tmp.Sync(); err != nil {
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	if err := os.Rename(tmpName, ss.path); err != nil {
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

func (ss *SnapshotStore) Load() ([]repository.MetricSnapshot, error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	file, err := os.Open(ss.path)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []repository.MetricSnapshot{}, nil
		}

		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	dec := json.NewDecoder(file)
	dec.DisallowUnknownFields()

	var metrics []repository.MetricSnapshot

	if err := dec.Decode(&metrics); err != nil {
		if errors.Is(err, io.EOF) {
			return []repository.MetricSnapshot{}, nil
		}

		return nil, fmt.Errorf("failed to decode metrics: %w", err)
	}

	if metrics == nil {
		return []repository.MetricSnapshot{}, nil
	}

	for _, m := range metrics {
		if err := validateMetricSnapshot(m); err != nil {
			return nil, fmt.Errorf("invalid metric snapshot: %w", err)
		}
	}

	return metrics, nil
}

func validateMetricSnapshot(m repository.MetricSnapshot) error {
	if m.ID == "" {
		return errors.New("id is empty")
	}

	switch m.Type {
	case "gauge":
		if m.Value == nil || m.Delta != nil {
			return errors.New("gauge requires value and must not have delta")
		}
	case "counter":
		if m.Delta == nil || m.Value != nil {
			return errors.New("counter requires delta and must not have value")
		}
	default:
		return fmt.Errorf("unknown metric type %q", m.Type)
	}

	return nil
}
