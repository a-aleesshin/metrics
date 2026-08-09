// Package cli загружает конфигурацию сервера метрик из флагов командной строки
// и переменных окружения (env имеет приоритет над флагами, флаги — над значениями по умолчанию).
package cli

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/a-aleesshin/metrics/internal/platform/db/postgres"
	"github.com/ilyakaznacheev/cleanenv"
)

// ValueSource обозначает источник значения параметра конфигурации: default, flag или env.
type ValueSource string

// StorageTypeFile, StorageTypePostgres и StorageTypeMemory — типы хранилища метрик
// (ServerConfig.StorageType); ValueSourceDefault, ValueSourceFlag и ValueSourceEnv —
// возможные источники значения параметра конфигурации.
const (
	StorageTypeFile     = "file"
	StorageTypePostgres = "postgres"
	StorageTypeMemory   = "memory"

	ValueSourceDefault ValueSource = "default"
	ValueSourceFlag    ValueSource = "flag"
	ValueSourceEnv     ValueSource = "env"
)

// ServerConfig — итоговая конфигурация сервера метрик: адрес, параметры хранилища,
// ключ подписи HMAC-SHA256 и настройки аудита.
type ServerConfig struct {
	Address         string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
	Postgres        *postgres.Config
	StorageType     string
	KeySignature    string
	AuditFile       string
	AuditURL        string
	CryptoKey       string
}

type rawServerConfig struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	Postgres        string `env:"DATABASE_DSN"`
	StorageType     string `env:"STORAGE_TYPE"`
	KeySignature    string `env:"KEY"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
	CryptoKey       string `env:"CRYPTO_KEY"`
}

type rawServerConfigSource struct {
	Address         ValueSource
	StoreInterval   ValueSource
	FileStoragePath ValueSource
	Restore         ValueSource
	Postgres        ValueSource
	KeySignature    ValueSource
}

func defaultRawServerConfig() (*rawServerConfig, *rawServerConfigSource) {
	return &rawServerConfig{
			Address:         "localhost:8080",
			StoreInterval:   300,
			FileStoragePath: "./metrics-db.json",
			Restore:         true,
			Postgres:        "",
			StorageType:     "memory",
			KeySignature:    "",
		},
		&rawServerConfigSource{
			Address:         ValueSourceDefault,
			StoreInterval:   ValueSourceDefault,
			FileStoragePath: ValueSourceDefault,
			Restore:         ValueSourceDefault,
			Postgres:        ValueSourceDefault,
			KeySignature:    ValueSourceDefault,
		}
}

// LoadConfig собирает конфигурацию сервера из args (флаги) и переменных окружения;
// env переопределяет флаги. Тип хранилища выбирается по заданным DSN/пути к файлу.
func LoadConfig(args []string) (*ServerConfig, error) {
	raw, source := defaultRawServerConfig()
	err := parseServerFlags(raw, source, args)

	if err != nil {
		return nil, fmt.Errorf("failed to parse command line arguments: %w", err)
	}

	if err := cleanenv.ReadEnv(raw); err != nil {
		return nil, fmt.Errorf("read env config: %w", err)
	}

	markEnvSources(source)

	cfg, err := buildServerConfig(raw, source)

	if err != nil {
		return nil, fmt.Errorf("build server config: %w", err)
	}

	return cfg, nil
}

func parseServerFlags(raw *rawServerConfig, rawSource *rawServerConfigSource, args []string) error {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	fs.StringVar(&raw.Address, "a", raw.Address, "HTTP server address")
	fs.IntVar(&raw.StoreInterval, "i", raw.StoreInterval, "store interval in seconds")
	fs.StringVar(&raw.FileStoragePath, "f", raw.FileStoragePath, "file storage path")
	fs.StringVar(&raw.Postgres, "d", raw.Postgres, "database DSN")
	fs.BoolVar(&raw.Restore, "r", raw.Restore, "restore metrics from file on startup")
	fs.StringVar(&raw.KeySignature, "k", raw.KeySignature, "key signature")
	fs.StringVar(&raw.AuditFile, "audit-file", raw.AuditFile, "path to audit log file (audit disabled if empty)")
	fs.StringVar(&raw.AuditURL, "audit-url", raw.AuditURL, "URL to send audit events to (audit disabled if empty)")
	fs.StringVar(&raw.CryptoKey, "crypto-key", raw.CryptoKey, "path to RSA private key PEM file (decryption disabled if empty)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse command line arguments: %w", err)
	}

	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			rawSource.Address = ValueSourceFlag
		case "i":
			rawSource.StoreInterval = ValueSourceFlag
		case "f":
			rawSource.FileStoragePath = ValueSourceFlag
		case "d":
			rawSource.Postgres = ValueSourceFlag
		case "r":
			rawSource.Restore = ValueSourceFlag
		case "k":
			rawSource.KeySignature = ValueSourceFlag
		}
	})

	return nil
}

func markEnvSources(sources *rawServerConfigSource) {
	if _, ok := os.LookupEnv("ADDRESS"); ok {
		sources.Address = ValueSourceEnv
	}

	if _, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		sources.StoreInterval = ValueSourceEnv
	}

	if value, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok && value != "" {
		sources.FileStoragePath = ValueSourceEnv
	}

	if _, ok := os.LookupEnv("RESTORE"); ok {
		sources.Restore = ValueSourceEnv
	}

	if value, ok := os.LookupEnv("DATABASE_DSN"); ok && value != "" {
		sources.Postgres = ValueSourceEnv
	}

	if _, ok := os.LookupEnv("KEY"); ok {
		sources.KeySignature = ValueSourceEnv
	}
}

func buildServerConfig(raw *rawServerConfig, source *rawServerConfigSource) (*ServerConfig, error) {
	if raw.StoreInterval < 0 {
		return nil, fmt.Errorf("store interval must be >= 0")
	}

	var postgresConfig *postgres.Config
	var typeStorage = StorageTypeMemory

	hasPostgres := raw.Postgres != "" && source.Postgres != ValueSourceDefault
	hasFile := raw.FileStoragePath != "" && source.FileStoragePath != ValueSourceDefault

	switch {
	case hasPostgres:
		var err error
		postgresConfig, err = postgres.NewConfigFromString(raw.Postgres)

		if err != nil {
			return nil, fmt.Errorf("failed to create postgres config from DSN: %w", err)
		}

		typeStorage = StorageTypePostgres
	case hasFile:
		typeStorage = StorageTypeFile
	default:
		typeStorage = StorageTypeMemory
	}

	return &ServerConfig{
		Address:         raw.Address,
		StoreInterval:   time.Duration(raw.StoreInterval) * time.Second,
		FileStoragePath: raw.FileStoragePath,
		Restore:         raw.Restore,
		Postgres:        postgresConfig,
		StorageType:     typeStorage,
		KeySignature:    raw.KeySignature,
		AuditFile:       raw.AuditFile,
		AuditURL:        raw.AuditURL,
		CryptoKey:       raw.CryptoKey,
	}, nil
}
