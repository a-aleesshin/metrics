//go:build integration

package postgres

import (
	"context"
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

func TestNewConfigFromString_InvalidSSLMode(t *testing.T) {
	cfg, err := NewConfigFromString(
		"postgres://unknown:unknown@unknown:5432/unknown?sslmode=wrong",
	)

	require.Error(t, err)
	require.Nil(t, cfg)
	require.Contains(t, err.Error(), "parse dsn: cannot parse `postgres://unknown:xxxxx@unknown:5432/unknown?sslmode=wrong`: failed to configure TLS (sslmode is invalid)")
}

func TestPool_Integration_Fail(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := NewConfigFromString(
		"postgres://unknown:unknown@localhost:5432/unknown?sslmode=disable",
	)

	if err != nil {
		t.Fatalf("invalid config: %v", err)
	}

	pool, err := NewPool(ctx, cfg)

	require.NoError(t, err)
	require.NotNil(t, pool)
	t.Cleanup(pool.Close)

	err = pool.Ping(ctx)
	require.Error(t, err)
}
