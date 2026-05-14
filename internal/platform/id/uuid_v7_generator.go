package id

import (
	"fmt"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
	"github.com/google/uuid"
)

type UUIDV7Generator struct{}

func NewUUIDV7Generator() *UUIDV7Generator {
	return &UUIDV7Generator{}
}

func (g *UUIDV7Generator) NewID() (metric.ID, error) {
	raw, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate uuid v7: %w", err)
	}

	id, err := metric.NewID(raw.String())
	if err != nil {
		return "", fmt.Errorf("create metric id: %w", err)
	}

	return id, nil
}
