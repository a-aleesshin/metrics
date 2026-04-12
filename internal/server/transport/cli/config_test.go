package cli

import (
	"testing"
)

func resetEnv(t *testing.T) {
	t.Helper()
	t.Setenv("ADDRESS", "")
}

func TestLoadConfig_LoadFlags(t *testing.T) {
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
			cfg, err := LoadConfig(tt.args)

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

func TestLoadConfig_EnvOverridesFlags(t *testing.T) {
	resetEnv(t)
	t.Setenv("ADDRESS", "env-host:9999")

	cfg, err := LoadConfig([]string{"-a=flag-host:8080"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Address != "env-host:9999" {
		t.Fatalf("got address %q, want %q", cfg.Address, "env-host:9999")
	}
}

func TestGetStringValue(t *testing.T) {
	tests := []struct {
		name       string
		envName    string
		envValue   string
		flagValue  string
		want       string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:      "test method getStringValue return env value",
			envName:   "ADDRESS",
			envValue:  "localhost:9080",
			flagValue: "localhost:8080",
			want:      "localhost:9080",
			wantErr:   false,
		},
		{
			name:      "test method getStringValue return flag value",
			envName:   "ADDRESS",
			envValue:  "",
			flagValue: "localhost:1080",
			want:      "localhost:1080",
			wantErr:   false,
		},
		{
			name:       "test method getStringValue return error",
			envName:    "ADDRESS",
			envValue:   "",
			flagValue:  "",
			want:       "localhost:8080",
			wantErr:    true,
			wantErrMsg: "ADDRESS must be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			resetEnv(t)
			t.Setenv(tt.envName, tt.envValue)

			// Act
			got, err := getStringValue(&tt.flagValue, tt.envName)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if err.Error() != tt.wantErrMsg {
					t.Fatalf("got error %q, want %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestGetStringValue_NilFlagValue(t *testing.T) {
	resetEnv(t)
	envName := "ADDRESS"
	t.Setenv(envName, "")

	got, err := getStringValue(nil, envName)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if got != "" {
		t.Fatalf("got %q, want empty string", got)
	}
}
