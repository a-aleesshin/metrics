//go:build integration

package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPool_Integration_OK(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := NewConfigFromString(
		"postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
	)
	require.NoError(t, err)

	pool, err := NewPool(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, pool)

	t.Cleanup(pool.Close)

	require.NoError(t, pool.Ping(ctx))
}

func TestPool_Integration_Fail(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := NewConfigFromString(
		"postgres://unknown:unknown@unknown:5432/unknown?sslmode=disable",
	)

	if err != nil {
		t.Fatalf("invalid config: %v", err)
	}

	_, err = NewPool(ctx, cfg)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to connect to `user=unknown database=unknown`: hostname resolving error: lookup unknown: no such host") {
		t.Fatalf("unexpected error: %v", err)
	}
}
