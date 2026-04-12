package cli

import (
	"testing"
	"time"
)

func resetEnv(t *testing.T) {
	t.Helper()
	t.Setenv("ADDRESS", "")
	t.Setenv("REPORT_INTERVAL", "")
	t.Setenv("POLL_INTERVAL", "")
}

func TestLoadConfig_LoadFlags(t *testing.T) {
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
			resetEnv(t)
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

			if cfg.ReportInterval != tt.wantReport {
				t.Fatalf("expected report interval %v, got %v", tt.wantReport, cfg.ReportInterval)
			}

			if cfg.PollInterval != tt.wantPoll {
				t.Fatalf("expected poll interval %v, got %v", tt.wantPoll, cfg.PollInterval)
			}
		})
	}
}

func TestLoadConfig_EnvOverridesFlags(t *testing.T) {
	resetEnv(t)
	t.Setenv("ADDRESS", "env-host:9999")
	t.Setenv("REPORT_INTERVAL", "15")
	t.Setenv("POLL_INTERVAL", "7")

	cfg, err := LoadConfig([]string{"-a=flag-host:8080", "-r=30", "-p=5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Address != "env-host:9999" {
		t.Fatalf("got address %q, want %q", cfg.Address, "env-host:9999")
	}

	if cfg.ReportInterval != 15*time.Second {
		t.Fatalf("got report interval %v, want %v", cfg.ReportInterval, 15*time.Second)
	}

	if cfg.PollInterval != 7*time.Second {
		t.Fatalf("got poll interval %v, want %v", cfg.PollInterval, 7*time.Second)
	}
}

func TestLoadConfig_ReturnsErrorWhenReportIntervalEnvIsInvalid(t *testing.T) {
	resetEnv(t)

	t.Setenv("REPORT_INTERVAL", "abc")

	_, err := LoadConfig([]string{"-r=10"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := `invalid value for environment variable REPORT_INTERVAL: strconv.Atoi: parsing "abc": invalid syntax`
	if err.Error() != want {
		t.Fatalf("got error %q, want %q", err.Error(), want)
	}
}

func TestLoadConfig_ReturnsErrorWhenPollIntervalEnvIsZero(t *testing.T) {
	resetEnv(t)
	t.Setenv("POLL_INTERVAL", "0")

	_, err := LoadConfig([]string{"-p=5"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := "POLL_INTERVAL must be > 0"
	if err.Error() != want {
		t.Fatalf("got error %q, want %q", err.Error(), want)
	}
}

func TestLoadConfig_UsesEnvAndFlagsTogether(t *testing.T) {
	resetEnv(t)
	t.Setenv("ADDRESS", "env-host:9999")

	cfg, err := LoadConfig([]string{"-r=30", "-p=5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Address != "env-host:9999" {
		t.Fatalf("got address %q, want %q", cfg.Address, "env-host:9999")
	}

	if cfg.ReportInterval != 30*time.Second {
		t.Fatalf("got report interval %v, want %v", cfg.ReportInterval, 30*time.Second)
	}

	if cfg.PollInterval != 5*time.Second {
		t.Fatalf("got poll interval %v, want %v", cfg.PollInterval, 5*time.Second)
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

func TestGetIntValue(t *testing.T) {
	tests := []struct {
		name       string
		envName    string
		envValue   string
		flagValue  int
		want       int
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:      "returns env value when env is valid",
			envName:   "REPORT_INTERVAL",
			envValue:  "10",
			flagValue: 20,
			want:      10,
		},
		{
			name:      "returns flag value when env is empty",
			envName:   "REPORT_INTERVAL",
			envValue:  "",
			flagValue: 20,
			want:      20,
		},
		{
			name:       "returns error when env is not a number",
			envName:    "REPORT_INTERVAL",
			envValue:   "abc",
			flagValue:  20,
			want:       0,
			wantErr:    true,
			wantErrMsg: `invalid value for environment variable REPORT_INTERVAL: strconv.Atoi: parsing "abc": invalid syntax`,
		},
		{
			name:       "returns error when env is zero",
			envName:    "REPORT_INTERVAL",
			envValue:   "0",
			flagValue:  20,
			want:       0,
			wantErr:    true,
			wantErrMsg: "REPORT_INTERVAL must be > 0",
		},
		{
			name:       "returns error when env is negative",
			envName:    "REPORT_INTERVAL",
			envValue:   "-10",
			flagValue:  20,
			want:       0,
			wantErr:    true,
			wantErrMsg: "REPORT_INTERVAL must be > 0",
		},
		{
			name:       "returns error when flag is negative and env is empty",
			envName:    "REPORT_INTERVAL",
			envValue:   "",
			flagValue:  -10,
			want:       0,
			wantErr:    true,
			wantErrMsg: "REPORT_INTERVAL must be > 0",
		},
		{
			name:       "returns error when flag is zero and env is empty",
			envName:    "REPORT_INTERVAL",
			envValue:   "",
			flagValue:  0,
			want:       0,
			wantErr:    true,
			wantErrMsg: "REPORT_INTERVAL must be > 0",
		},
		{
			name:      "returns env value when env is set even if flag is invalid",
			envName:   "REPORT_INTERVAL",
			envValue:  "10",
			flagValue: 0,
			want:      10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			resetEnv(t)
			t.Setenv(tt.envName, tt.envValue)

			// Act
			got, err := getIntValue(&tt.flagValue, tt.envName)

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
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestGetIntValue_ReturnsErrorWhenFlagIsNilAndEnvIsEmpty(t *testing.T) {
	resetEnv(t)
	envName := "REPORT_INTERVAL"
	t.Setenv(envName, "")

	got, err := getIntValue(nil, envName)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if got != 0 {
		t.Fatalf("got %d, want 0", got)
	}
}
