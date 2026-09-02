// Package cli загружает конфигурацию агента из флагов командной строки,
// переменных окружения и JSON-файла (флаг -c/-config или переменная CONFIG).
// Приоритет источников: env > флаги > файл > значения по умолчанию.
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

// fileAgentConfig — формат JSON-файла конфигурации агента.
// Интервалы задаются строками длительности ("1s", "5m").
type fileAgentConfig struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
	GRPCAddress    string `json:"grpc_address"`
	Key            string `json:"key"`
	RateLimit      int    `json:"rate_limit"`
}

// AgentConfig — параметры запуска агента: адрес сервера, интервалы, ключ подписи,
// лимит воркеров и путь к публичному ключу шифрования.
type AgentConfig struct {
	Address        string `env:"ADDRESS"`
	ReportInterval time.Duration
	PollInterval   time.Duration
	KeySignature   string
	RateLimit      int
	CryptoKey      string
	GRPCAddress    string
}

var (
	addressDefault        = "localhost:8080"
	reportIntervalDefault = 10
	pollIntervalDefault   = 2
	keySignatureDefault   = ""
	rateLimitDefault      = 1
)

// LoadConfig разбирает флаги args и переменные окружения;
// значения из окружения имеют приоритет над флагами.
func LoadConfig(args []string) (*AgentConfig, error) {
	var config AgentConfig

	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	var address string
	var keySignature string
	var cryptoKey string
	var grpcAddress string
	var configPath string
	var reportInterval int
	var pollInterval int
	var rateLimit int

	fs.StringVar(&address, "a", addressDefault, "HTTP server address")
	fs.IntVar(&reportInterval, "r", reportIntervalDefault, "report interval in seconds")
	fs.IntVar(&pollInterval, "p", pollIntervalDefault, "poll interval in seconds")
	fs.StringVar(&keySignature, "k", keySignatureDefault, "key signature")
	fs.IntVar(&rateLimit, "l", rateLimitDefault, "outgoing requests rate limit")
	fs.StringVar(&cryptoKey, "crypto-key", "", "path to RSA public key PEM file (encryption disabled if empty)")
	fs.StringVar(&grpcAddress, "grpc-address", "", "gRPC server address (metrics are sent over gRPC when set)")
	fs.StringVar(&configPath, "c", "", "path to JSON config file")
	fs.StringVar(&configPath, "config", "", "path to JSON config file")

	err := fs.Parse(args)

	if err != nil {
		return nil, fmt.Errorf("failed to parse command line arguments: %w", err)
	}

	setFlags := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) { setFlags[f.Name] = true })

	if envPath := os.Getenv("CONFIG"); envPath != "" {
		configPath = envPath
	}

	// Значения из файла подставляются вместо значений по умолчанию только
	// для флагов, не заданных явно; env-приоритет обрабатывается ниже.
	if configPath != "" {
		file, err := readFileConfig(configPath)
		if err != nil {
			return nil, err
		}

		if !setFlags["a"] && file.Address != "" {
			address = file.Address
		}

		if !setFlags["r"] && file.ReportInterval != "" {
			seconds, err := durationSeconds(file.ReportInterval, "report_interval")
			if err != nil {
				return nil, err
			}
			reportInterval = seconds
		}

		if !setFlags["p"] && file.PollInterval != "" {
			seconds, err := durationSeconds(file.PollInterval, "poll_interval")
			if err != nil {
				return nil, err
			}
			pollInterval = seconds
		}

		if !setFlags["k"] && file.Key != "" {
			keySignature = file.Key
		}

		if !setFlags["l"] && file.RateLimit > 0 {
			rateLimit = file.RateLimit
		}

		if !setFlags["crypto-key"] && file.CryptoKey != "" {
			cryptoKey = file.CryptoKey
		}

		if !setFlags["grpc-address"] && file.GRPCAddress != "" {
			grpcAddress = file.GRPCAddress
		}
	}

	valueAddress, err := getStringValue(&address, "ADDRESS")

	if err != nil {
		return nil, err
	}

	config.Address = valueAddress

	valueReportInterval, err := getIntValue(&reportInterval, "REPORT_INTERVAL")

	if err != nil {
		return nil, err
	}

	config.ReportInterval = time.Duration(valueReportInterval) * time.Second

	valuePollInterval, err := getIntValue(&pollInterval, "POLL_INTERVAL")

	if err != nil {
		return nil, err
	}

	config.PollInterval = time.Duration(valuePollInterval) * time.Second

	config.KeySignature = getOptionalStringValue(&keySignature, "KEY")
	config.CryptoKey = getOptionalStringValue(&cryptoKey, "CRYPTO_KEY")
	config.GRPCAddress = getOptionalStringValue(&grpcAddress, "GRPC_ADDRESS")

	valueRateLimit, err := getIntValue(&rateLimit, "RATE_LIMIT")

	if err != nil {
		return nil, err
	}

	config.RateLimit = valueRateLimit

	return &config, nil
}

func readFileConfig(path string) (*fileAgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var file fileAgentConfig
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse config file %s: %w", path, err)
	}

	return &file, nil
}

func durationSeconds(value, field string) (int, error) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("config file: parse %s: %w", field, err)
	}

	if duration <= 0 || duration%time.Second != 0 {
		return 0, fmt.Errorf("config file: %s must be a positive whole number of seconds, got %q", field, value)
	}

	return int(duration / time.Second), nil
}

func getOptionalStringValue(flagValue *string, envName string) string {
	envValue := os.Getenv(envName)
	if envValue != "" {
		return envValue
	}

	if flagValue == nil {
		return ""
	}

	return *flagValue
}

func getStringValue(flagValue *string, envName string) (string, error) {
	envValue := os.Getenv(envName)
	if envValue != "" {
		return envValue, nil
	}

	if flagValue == nil || *flagValue == "" {
		return "", fmt.Errorf("%s must be set", envName)
	}

	return *flagValue, nil
}

func getIntValue(flagValue *int, envName string) (int, error) {
	envValue := os.Getenv(envName)

	if envValue != "" {
		val, err := strconv.Atoi(envValue)

		if err != nil {
			return 0, fmt.Errorf("invalid value for environment variable %s: %w", envName, err)
		}

		if val <= 0 {
			return 0, fmt.Errorf("%s must be > 0", envName)
		}

		return val, nil
	}

	if flagValue == nil || *flagValue <= 0 {
		return 0, fmt.Errorf("%s must be > 0", envName)
	}

	return *flagValue, nil
}
