package cli

import (
	"testing"
	"time"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantAddress string
		wantReport  time.Duration
		wantPoll    time.Duration
		wantErr     bool
	}{
		{
			name:        "defaults",
			args:        []string{},
			wantAddress: "localhost:8080",
			wantReport:  10 * time.Second,
			wantPoll:    2 * time.Second,
		},
		{
			name:        "new all flags",
			args:        []string{"-a=127.0.0.1:9000", "-r=30", "-p=5"},
			wantAddress: "127.0.0.1:9000",
			wantReport:  30 * time.Second,
			wantPoll:    5 * time.Second,
		},
		{
			name:    "unknown flag",
			args:    []string{"-x=1"},
			wantErr: true,
		},
		{
			name:    "invalid report interval",
			args:    []string{"-r=abc"},
			wantErr: true,
		},
		{
			name:    "invalid poll interval zero",
			args:    []string{"-p=0"},
			wantErr: true,
		},
		{
			name:    "invalid poll interval negative",
			args:    []string{"-p=-1"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			cfg, err := ParseConfig(tt.args)

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

			if cfg.ReportInterval != tt.wantReport {
				t.Fatalf("expected report interval %v, got %v", tt.wantReport, cfg.ReportInterval)
			}

			if cfg.PollInterval != tt.wantPoll {
				t.Fatalf("expected poll interval %v, got %v", tt.wantPoll, cfg.PollInterval)
			}
		})
	}
}
