package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/a-aleesshin/metrics/internal/server/audit"
)

type FileObserver struct {
	mu   sync.Mutex
	file *os.File
}

func NewFileObserver(path string) (*FileObserver, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)

	if err != nil {
		return nil, fmt.Errorf("open audit file: %w", err)
	}

	return &FileObserver{file: file}, nil
}

func (o *FileObserver) Notify(_ context.Context, event audit.Event) error {
	data, err := json.Marshal(event)

	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	data = append(data, '\n')

	o.mu.Lock()
	defer o.mu.Unlock()

	if _, err := o.file.Write(data); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}

	return nil
}

func (o *FileObserver) Close() error {
	return o.file.Close()
}
