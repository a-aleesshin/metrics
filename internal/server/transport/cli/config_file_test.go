package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	return path
}

func TestLoadConfig_FromFile(t *testing.T) {
	resetEnv(t)

	path := writeConfigFile(t, `{
		"address": "localhost:9090",
		"restore": false,
		"store_interval": "5s",
		"store_file": "/tmp/file.db",
		"database_dsn": "",
		"crypto_key": "/tmp/key.pem",
		"key": "file-key",
		"grpc_address": "localhost:3200",
		"trusted_subnet": "10.0.0.0/8"
	}`)

	cfg, err := LoadConfig([]string{"-c", path})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Address != "localhost:9090" {
		t.Fatalf("expected address from file, got %s", cfg.Address)
	}

	if cfg.Restore {
		t.Fatal("expected restore=false from file")
	}

	if cfg.StoreInterval != 5*time.Second {
		t.Fatalf("expected store interval 5s, got %s", cfg.StoreInterval)
	}

	if cfg.FileStoragePath != "/tmp/file.db" {
		t.Fatalf("expected store file from file, got %s", cfg.FileStoragePath)
	}

	if cfg.StorageType != StorageTypeFile {
		t.Fatalf("expected file storage type, got %s", cfg.StorageType)
	}

	if cfg.CryptoKey != "/tmp/key.pem" {
		t.Fatalf("expected crypto key from file, got %s", cfg.CryptoKey)
	}

	if cfg.KeySignature != "file-key" {
		t.Fatalf("expected key from file, got %s", cfg.KeySignature)
	}

	if cfg.GRPCAddress != "localhost:3200" {
		t.Fatalf("expected grpc address from file, got %s", cfg.GRPCAddress)
	}

	if cfg.TrustedSubnet != "10.0.0.0/8" {
		t.Fatalf("expected trusted subnet from file, got %s", cfg.TrustedSubnet)
	}
}

func TestLoadConfig_GRPCFlagsAndEnv(t *testing.T) {
	resetEnv(t)

	cfg, err := LoadConfig([]string{"-grpc-address", "localhost:3200", "-t", "192.168.0.0/24"})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.GRPCAddress != "localhost:3200" || cfg.TrustedSubnet != "192.168.0.0/24" {
		t.Fatalf("expected values from flags, got %+v", cfg)
	}

	t.Setenv("GRPC_ADDRESS", "localhost:3300")
	t.Setenv("TRUSTED_SUBNET", "10.0.0.0/8")

	cfg, err = LoadConfig([]string{"-grpc-address", "localhost:3200", "-t", "192.168.0.0/24"})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.GRPCAddress != "localhost:3300" || cfg.TrustedSubnet != "10.0.0.0/8" {
		t.Fatalf("expected env to beat flags, got %+v", cfg)
	}
}

func TestLoadConfig_FlagBeatsFile(t *testing.T) {
	resetEnv(t)

	path := writeConfigFile(t, `{"address": "localhost:9090", "store_interval": "5s"}`)

	cfg, err := LoadConfig([]string{"-c", path, "-a", "localhost:7070"})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Address != "localhost:7070" {
		t.Fatalf("expected flag to beat file, got %s", cfg.Address)
	}

	if cfg.StoreInterval != 5*time.Second {
		t.Fatalf("expected store interval from file, got %s", cfg.StoreInterval)
	}
}

func TestLoadConfig_EnvBeatsFileAndFlag(t *testing.T) {
	resetEnv(t)

	path := writeConfigFile(t, `{"address": "localhost:9090"}`)

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
	resetEnv(t)

	path := writeConfigFile(t, `{"address": "localhost:9090"}`)

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
	resetEnv(t)

	if _, err := LoadConfig([]string{"-config", "/nonexistent/config.json"}); err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestLoadConfig_FractionalStoreIntervalRejected(t *testing.T) {
	resetEnv(t)

	for _, body := range []string{
		`{"store_interval": "500ms"}`,
		`{"store_interval": "1500ms"}`,
	} {
		path := writeConfigFile(t, body)

		if _, err := LoadConfig([]string{"-c", path}); err == nil {
			t.Fatalf("expected error for %s, got nil", body)
		}
	}
}
