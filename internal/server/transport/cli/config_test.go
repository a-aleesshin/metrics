package cli

import (
	"testing"
	"time"
)

func resetEnv(t *testing.T) {
	t.Helper()
	t.Setenv("ADDRESS", "")
	t.Setenv("STORE_INTERVAL", "")
	t.Setenv("FILE_STORAGE_PATH", "")
	t.Setenv("RESTORE", "")
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		env          map[string]string
		wantAddress  string
		wantInterval time.Duration
		wantFilePath string
		wantRestore  bool
		wantErr      bool
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
			name: "env overrides flags",
			args: []string{"-a=127.0.0.1:9090", "-i=10", "-f=/tmp/from-flag.json", "-r=false"},
			env: map[string]string{
				"ADDRESS":           "env-host:7777",
				"STORE_INTERVAL":    "42",
				"FILE_STORAGE_PATH": "/tmp/from-env.json",
				"RESTORE":           "true",
			},
			wantAddress:  "env-host:7777",
			wantInterval: 42 * time.Second,
			wantFilePath: "/tmp/from-env.json",
			wantRestore:  true,
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
		})
	}
}

func TestGetIntValueAllowZero(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		flagValue int
		want      int
		wantErr   bool
	}{
		{
			name:      "env has positive value",
			envValue:  "10",
			flagValue: 5,
			want:      10,
		},
		{
			name:      "env has zero value",
			envValue:  "0",
			flagValue: 5,
			want:      0,
		},
		{
			name:      "flag used when env empty",
			envValue:  "",
			flagValue: 7,
			want:      7,
		},
		{
			name:      "invalid env value",
			envValue:  "abc",
			flagValue: 7,
			wantErr:   true,
		},
		{
			name:      "negative env value",
			envValue:  "-1",
			flagValue: 7,
			wantErr:   true,
		},
		{
			name:      "negative flag value",
			envValue:  "",
			flagValue: -1,
			wantErr:   true,
		},
		{
			name:      "nil flag pointer and empty env",
			envValue:  "",
			flagValue: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			resetEnv(t)
			if tt.envValue != "" {
				t.Setenv("STORE_INTERVAL", tt.envValue)
			}

			var flagPtr *int
			if tt.name == "nil flag pointer and empty env" {
				flagPtr = nil
			} else {
				v := tt.flagValue
				flagPtr = &v
			}

			// Act
			got, err := getIntValueAllowZero(flagPtr, "STORE_INTERVAL")

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
			if got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestGetBoolValue(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		flagValue bool
		want      bool
		wantErr   bool
		nilFlag   bool
	}{
		{
			name:      "env true overrides flag",
			envValue:  "true",
			flagValue: false,
			want:      true,
		},
		{
			name:      "env false overrides flag",
			envValue:  "false",
			flagValue: true,
			want:      false,
		},
		{
			name:      "flag used when env empty",
			envValue:  "",
			flagValue: true,
			want:      true,
		},
		{
			name:      "invalid env bool",
			envValue:  "not-bool",
			flagValue: true,
			wantErr:   true,
		},
		{
			name:     "nil flag pointer and empty env",
			envValue: "",
			nilFlag:  true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			resetEnv(t)
			if tt.envValue != "" {
				t.Setenv("RESTORE", tt.envValue)
			}

			var flagPtr *bool
			if tt.nilFlag {
				flagPtr = nil
			} else {
				v := tt.flagValue
				flagPtr = &v
			}

			// Act
			got, err := getBoolValue(flagPtr, "RESTORE")

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
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
