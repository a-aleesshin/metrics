// Package id реализует генерацию идентификаторов метрик на основе UUID v7.
package id

import (
	"fmt"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
	"github.com/google/uuid"
)

// UUIDV7Generator генерирует идентификаторы метрик в формате UUID v7.
type UUIDV7Generator struct{}

// NewUUIDV7Generator создаёт генератор UUID v7.
func NewUUIDV7Generator() *UUIDV7Generator {
	return &UUIDV7Generator{}
}

// NewID генерирует новый UUID v7 и оборачивает его в metric.ID.
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
