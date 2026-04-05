package cli

import (
	"testing"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantAddress string
		wantErr     bool
	}{
		{
			name:        "defaults",
			args:        []string{},
			wantAddress: "localhost:8080",
		},
		{
			name:        "new all flags",
			args:        []string{"-a=127.0.0.1:9000"},
			wantAddress: "127.0.0.1:9000",
		},
		{
			name:    "unknown flag",
			args:    []string{"-x=1"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			cfg, err := ParseConfig(tt.args)

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
		})
	}
}
