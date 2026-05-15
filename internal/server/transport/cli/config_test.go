package cli

import (
	"os"
	"testing"
	"time"
)

func resetEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"ADDRESS",
		"STORE_INTERVAL",
		"FILE_STORAGE_PATH",
		"RESTORE",
		"DATABASE_DSN",
	} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset env %s: %v", key, err)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		env             map[string]string
		wantAddress     string
		wantInterval    time.Duration
		wantFilePath    string
		wantRestore     bool
		wantErr         bool
		wantDatabaseDsn string
	}{
		{
			name:         "defaults",
			args:         []string{},
			env:          map[string]string{},
			wantAddress:  "localhost:8080",
			wantInterval: 300 * time.Second,
			wantFilePath: "./metrics-db.json",
			wantRestore:  true,
		},
		{
			name:         "flags only",
			args:         []string{"-a=127.0.0.1:9090", "-i=10", "-f=/tmp/metrics.json", "-r=false"},
			env:          map[string]string{},
			wantAddress:  "127.0.0.1:9090",
			wantInterval: 10 * time.Second,
			wantFilePath: "/tmp/metrics.json",
			wantRestore:  false,
		},
		{
			name: "env_overrides_flags",
			args: []string{"-a=127.0.0.1:9090", "-i=10", "-f=/tmp/from-flag.json", "-r=false", "-d=postgres://postgres:postgres@localhost:54321/postgres?sslmode=disable"},
			env: map[string]string{
				"ADDRESS":           "env-host:7777",
				"STORE_INTERVAL":    "42",
				"FILE_STORAGE_PATH": "/tmp/from-env.json",
				"RESTORE":           "true",
				"DATABASE_DSN":      "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
			},
			wantAddress:     "env-host:7777",
			wantInterval:    42 * time.Second,
			wantFilePath:    "/tmp/from-env.json",
			wantRestore:     true,
			wantDatabaseDsn: "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		},
		{
			name:    "unknown flag",
			args:    []string{"-x=1"},
			env:     map[string]string{},
			wantErr: true,
		},
		{
			name:    "invalid store interval in env",
			args:    []string{},
			env:     map[string]string{"STORE_INTERVAL": "abc"},
			wantErr: true,
		},
		{
			name:    "negative store interval in env",
			args:    []string{},
			env:     map[string]string{"STORE_INTERVAL": "-1"},
			wantErr: true,
		},
		{
			name:    "invalid restore in env",
			args:    []string{},
			env:     map[string]string{"RESTORE": "not-bool"},
			wantErr: true,
		},
		{
			name:    "negative store interval in flag",
			args:    []string{"-i=-1"},
			env:     map[string]string{},
			wantErr: true,
		},
		{
			name:    "invalid_database_dsn",
			args:    []string{"-d=some-invalid-dsn"},
			env:     map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			resetEnv(t)
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			// Act
			cfg, err := LoadConfig(tt.args)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.Address != tt.wantAddress {
				t.Fatalf("expected address %q, got %q", tt.wantAddress, cfg.Address)
			}
			if cfg.StoreInterval != tt.wantInterval {
				t.Fatalf("expected store interval %v, got %v", tt.wantInterval, cfg.StoreInterval)
			}
			if cfg.FileStoragePath != tt.wantFilePath {
				t.Fatalf("expected file path %q, got %q", tt.wantFilePath, cfg.FileStoragePath)
			}
			if cfg.Restore != tt.wantRestore {
				t.Fatalf("expected restore %v, got %v", tt.wantRestore, cfg.Restore)
			}
			if tt.wantDatabaseDsn != "" && cfg.Postgres.ConnectionString() != tt.wantDatabaseDsn {
				t.Fatalf("expected restore %v, got %v", tt.wantDatabaseDsn, cfg.Postgres.ConnectionString())
			}
		})
	}
}
