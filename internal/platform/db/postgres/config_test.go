package postgres

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_ConnectionString(t *testing.T) {
	// Arrange
	cfg := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "postgres",
		SSLMode:  "disable",
	}
	want := "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"

	// Act
	connectionString := cfg.ConnectionString()

	// Assert
	t.Run("connection_string", func(t *testing.T) {
		t.Log(connectionString)
		t.Logf("Expected: %s", want)
		t.Logf("Actual: %s", connectionString)

		require.Equal(t, want, connectionString)
	})
}

func TestConfig_NewConfigFromString(t *testing.T) {
	tests := []struct {
		name        string
		dsnString   string
		want        Config
		errContains string
	}{
		{
			name:      "valid_config",
			dsnString: "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
			want: Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				DBName:   "postgres",
				SSLMode:  "disable",
			},
		},
		{
			name:        "dsn_string_invalid_config_missing_db",
			dsnString:   "postgres://postgres:postgres@localhost:/",
			errContains: "db name cannot be empty",
		},
		{
			name:        "params_string_invalid_config_missing_db",
			dsnString:   "host=localhost port=5432 user=postgres password=postgres sslmode=disable",
			errContains: "db name cannot be empty",
		},
		{
			name:        "empty_string_invalid",
			dsnString:   "",
			errContains: "DSN cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfigFromString(tt.dsnString)

			if tt.errContains != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.errContains)
				}

				if !strings.Contains(err.Error(), tt.errContains) {
					fmt.Println(err.Error())
					t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			require.Equal(t, tt.want, *got)
		})
	}
}
