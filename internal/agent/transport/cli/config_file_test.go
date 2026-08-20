package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeAgentConfigFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	return path
}

func clearAgentEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{"ADDRESS", "REPORT_INTERVAL", "POLL_INTERVAL", "KEY", "RATE_LIMIT", "CRYPTO_KEY", "CONFIG"} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset env %s: %v", key, err)
		}
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	clearAgentEnv(t)

	path := writeAgentConfigFile(t, `{
		"address": "localhost:9090",
		"report_interval": "20s",
		"poll_interval": "3s",
		"crypto_key": "/tmp/pub.pem",
		"key": "file-key",
		"rate_limit": 4
	}`)

	cfg, err := LoadConfig([]string{"-c", path})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Address != "localhost:9090" {
		t.Fatalf("expected address from file, got %s", cfg.Address)
	}

	if cfg.ReportInterval != 20*time.Second {
		t.Fatalf("expected report interval 20s, got %s", cfg.ReportInterval)
	}

	if cfg.PollInterval != 3*time.Second {
		t.Fatalf("expected poll interval 3s, got %s", cfg.PollInterval)
	}

	if cfg.CryptoKey != "/tmp/pub.pem" {
		t.Fatalf("expected crypto key from file, got %s", cfg.CryptoKey)
	}

	if cfg.KeySignature != "file-key" {
		t.Fatalf("expected key from file, got %s", cfg.KeySignature)
	}

	if cfg.RateLimit != 4 {
		t.Fatalf("expected rate limit from file, got %d", cfg.RateLimit)
	}
}

func TestLoadConfig_FlagBeatsFile(t *testing.T) {
	clearAgentEnv(t)

	path := writeAgentConfigFile(t, `{"address": "localhost:9090", "poll_interval": "3s"}`)

	cfg, err := LoadConfig([]string{"-c", path, "-a", "localhost:7070"})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Address != "localhost:7070" {
		t.Fatalf("expected flag to beat file, got %s", cfg.Address)
	}

	if cfg.PollInterval != 3*time.Second {
		t.Fatalf("expected poll interval from file, got %s", cfg.PollInterval)
	}
}

func TestLoadConfig_EnvBeatsFileAndFlag(t *testing.T) {
	clearAgentEnv(t)

	path := writeAgentConfigFile(t, `{"address": "localhost:9090"}`)

	t.Setenv("ADDRESS", "localhost:6060")

	cfg, err := LoadConfig([]string{"-c", path, "-a", "localhost:7070"})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Address != "localhost:6060" {
		t.Fatalf("expected env to beat flag and file, got %s", cfg.Address)
	}
}

func TestLoadConfig_ConfigPathFromEnv(t *testing.T) {
	clearAgentEnv(t)

	path := writeAgentConfigFile(t, `{"address": "localhost:9090"}`)

	t.Setenv("CONFIG", path)

	cfg, err := LoadConfig([]string{})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Address != "localhost:9090" {
		t.Fatalf("expected address from CONFIG file, got %s", cfg.Address)
	}
}

func TestLoadConfig_MissingConfigFile(t *testing.T) {
	clearAgentEnv(t)

	if _, err := LoadConfig([]string{"-config", "/nonexistent/config.json"}); err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestLoadConfig_FractionalIntervalRejected(t *testing.T) {
	clearAgentEnv(t)

	tests := []struct {
		name string
		body string
	}{
		{name: "sub-second report interval", body: `{"report_interval": "500ms"}`},
		{name: "fractional report interval", body: `{"report_interval": "1500ms"}`},
		{name: "fractional poll interval", body: `{"poll_interval": "1.5s"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeAgentConfigFile(t, tt.body)

			if _, err := LoadConfig([]string{"-c", path}); err == nil {
				t.Fatalf("expected error for %s, got nil", tt.body)
			}
		})
	}
}
