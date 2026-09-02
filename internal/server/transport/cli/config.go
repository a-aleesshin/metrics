// Package cli загружает конфигурацию сервера метрик из флагов командной строки
// и переменных окружения (env имеет приоритет над флагами, флаги — над значениями по умолчанию).
package cli

import (
	"encoding/json"
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
	ValueSourceFile    ValueSource = "file"
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
	GRPCAddress     string
	TrustedSubnet   string
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
	GRPCAddress     string `env:"GRPC_ADDRESS"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET"`
}

type rawServerConfigSource struct {
	Address         ValueSource
	StoreInterval   ValueSource
	FileStoragePath ValueSource
	Restore         ValueSource
	Postgres        ValueSource
	KeySignature    ValueSource
	AuditFile       ValueSource
	AuditURL        ValueSource
	CryptoKey       ValueSource
	GRPCAddress     ValueSource
	TrustedSubnet   ValueSource
}

// fileServerConfig — формат JSON-файла конфигурации сервера.
// Интервалы задаются строками длительности ("1s", "5m").
type fileServerConfig struct {
	Address       string `json:"address"`
	Restore       *bool  `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
	GRPCAddress   string `json:"grpc_address"`
	TrustedSubnet string `json:"trusted_subnet"`
	Key           string `json:"key"`
	AuditFile     string `json:"audit_file"`
	AuditURL      string `json:"audit_url"`
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
			AuditFile:       ValueSourceDefault,
			AuditURL:        ValueSourceDefault,
			CryptoKey:       ValueSourceDefault,
			GRPCAddress:     ValueSourceDefault,
			TrustedSubnet:   ValueSourceDefault,
		}
}

// LoadConfig собирает конфигурацию сервера из args (флаги), переменных окружения
// и JSON-файла (флаг -c/-config или переменная CONFIG). Приоритет источников:
// env > флаги > файл > значения по умолчанию. Тип хранилища выбирается
// по заданным DSN/пути к файлу.
func LoadConfig(args []string) (*ServerConfig, error) {
	raw, source := defaultRawServerConfig()
	configPath, err := parseServerFlags(raw, source, args)

	if err != nil {
		return nil, fmt.Errorf("failed to parse command line arguments: %w", err)
	}

	if err := cleanenv.ReadEnv(raw); err != nil {
		return nil, fmt.Errorf("read env config: %w", err)
	}

	markEnvSources(source)

	if envPath := os.Getenv("CONFIG"); envPath != "" {
		configPath = envPath
	}

	if err := applyFileConfig(configPath, raw, source); err != nil {
		return nil, err
	}

	cfg, err := buildServerConfig(raw, source)

	if err != nil {
		return nil, fmt.Errorf("build server config: %w", err)
	}

	return cfg, nil
}

// applyFileConfig подставляет значения из JSON-файла для параметров,
// не заданных ни флагом, ни переменной окружения.
func applyFileConfig(path string, raw *rawServerConfig, source *rawServerConfigSource) error {
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}

	var file fileServerConfig
	if err := json.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("parse config file %s: %w", path, err)
	}

	if source.StoreInterval == ValueSourceDefault && file.StoreInterval != "" {
		seconds, err := parseStoreInterval(path, file.StoreInterval)
		if err != nil {
			return err
		}

		raw.StoreInterval = seconds
		source.StoreInterval = ValueSourceFile
	}

	applyFileString(&source.Address, &raw.Address, file.Address)
	applyFileString(&source.FileStoragePath, &raw.FileStoragePath, file.StoreFile)
	applyFileString(&source.Postgres, &raw.Postgres, file.DatabaseDSN)
	applyFileString(&source.KeySignature, &raw.KeySignature, file.Key)
	applyFileString(&source.AuditFile, &raw.AuditFile, file.AuditFile)
	applyFileString(&source.AuditURL, &raw.AuditURL, file.AuditURL)
	applyFileString(&source.CryptoKey, &raw.CryptoKey, file.CryptoKey)
	applyFileString(&source.GRPCAddress, &raw.GRPCAddress, file.GRPCAddress)
	applyFileString(&source.TrustedSubnet, &raw.TrustedSubnet, file.TrustedSubnet)
	applyFileBool(&source.Restore, &raw.Restore, file.Restore)

	return nil
}

func applyFileString(source *ValueSource, dst *string, value string) {
	if *source != ValueSourceDefault || value == "" {
		return
	}

	*dst = value
	*source = ValueSourceFile
}

func applyFileBool(source *ValueSource, dst *bool, value *bool) {
	if *source != ValueSourceDefault || value == nil {
		return
	}

	*dst = *value
	*source = ValueSourceFile
}

func parseStoreInterval(path, value string) (int, error) {
	interval, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("config file %s: parse store_interval: %w", path, err)
	}

	if interval < 0 || interval%time.Second != 0 {
		return 0, fmt.Errorf("config file %s: store_interval must be a whole number of seconds, got %q", path, value)
	}

	return int(interval / time.Second), nil
}

func parseServerFlags(raw *rawServerConfig, rawSource *rawServerConfigSource, args []string) (string, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	var configPath string

	fs.StringVar(&raw.Address, "a", raw.Address, "HTTP server address")
	fs.IntVar(&raw.StoreInterval, "i", raw.StoreInterval, "store interval in seconds")
	fs.StringVar(&raw.FileStoragePath, "f", raw.FileStoragePath, "file storage path")
	fs.StringVar(&raw.Postgres, "d", raw.Postgres, "database DSN")
	fs.BoolVar(&raw.Restore, "r", raw.Restore, "restore metrics from file on startup")
	fs.StringVar(&raw.KeySignature, "k", raw.KeySignature, "key signature")
	fs.StringVar(&raw.AuditFile, "audit-file", raw.AuditFile, "path to audit log file (audit disabled if empty)")
	fs.StringVar(&raw.AuditURL, "audit-url", raw.AuditURL, "URL to send audit events to (audit disabled if empty)")
	fs.StringVar(&raw.CryptoKey, "crypto-key", raw.CryptoKey, "path to RSA private key PEM file (decryption disabled if empty)")
	fs.StringVar(&raw.GRPCAddress, "grpc-address", raw.GRPCAddress, "gRPC server address (gRPC disabled if empty)")
	fs.StringVar(&raw.TrustedSubnet, "t", raw.TrustedSubnet, "trusted subnet in CIDR notation (check disabled if empty)")
	fs.StringVar(&configPath, "c", configPath, "path to JSON config file")
	fs.StringVar(&configPath, "config", configPath, "path to JSON config file")

	if err := fs.Parse(args); err != nil {
		return "", fmt.Errorf("failed to parse command line arguments: %w", err)
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
		case "audit-file":
			rawSource.AuditFile = ValueSourceFlag
		case "audit-url":
			rawSource.AuditURL = ValueSourceFlag
		case "crypto-key":
			rawSource.CryptoKey = ValueSourceFlag
		case "grpc-address":
			rawSource.GRPCAddress = ValueSourceFlag
		case "t":
			rawSource.TrustedSubnet = ValueSourceFlag
		}
	})

	return configPath, nil
}

// markEnvSources помечает источником env параметры, заданные переменными
// окружения. Правило единое для всех переменных: пустое значение считается
// не заданным и источник не меняет.
func markEnvSources(sources *rawServerConfigSource) {
	markEnv := func(name string, source *ValueSource) {
		if value, ok := os.LookupEnv(name); ok && value != "" {
			*source = ValueSourceEnv
		}
	}

	markEnv("ADDRESS", &sources.Address)
	markEnv("STORE_INTERVAL", &sources.StoreInterval)
	markEnv("FILE_STORAGE_PATH", &sources.FileStoragePath)
	markEnv("RESTORE", &sources.Restore)
	markEnv("DATABASE_DSN", &sources.Postgres)
	markEnv("KEY", &sources.KeySignature)
	markEnv("AUDIT_FILE", &sources.AuditFile)
	markEnv("AUDIT_URL", &sources.AuditURL)
	markEnv("CRYPTO_KEY", &sources.CryptoKey)
	markEnv("GRPC_ADDRESS", &sources.GRPCAddress)
	markEnv("TRUSTED_SUBNET", &sources.TrustedSubnet)
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
		GRPCAddress:     raw.GRPCAddress,
		TrustedSubnet:   raw.TrustedSubnet,
	}, nil
}
