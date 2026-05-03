package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type ServerConfig struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
}

var (
	addressDefault         = "localhost:8080"
	storeIntervalDefault   = 300
	fileStoragePathDefault = "./metrics-db.json"
	restoreDefault         = true
)

func LoadConfig(args []string) (*ServerConfig, error) {
	var cfg ServerConfig

	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	var address string
	var storeInterval int
	var filePath string
	var restore bool

	fs.StringVar(&address, "a", addressDefault, "HTTP server address")
	fs.IntVar(&storeInterval, "i", storeIntervalDefault, "store interval in seconds")
	fs.StringVar(&filePath, "f", fileStoragePathDefault, "file storage path")
	fs.BoolVar(&restore, "r", restoreDefault, "restore metrics from file on startup")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse command line arguments: %w", err)
	}

	valueAddress, err := getStringValue(&address, "ADDRESS")
	if err != nil {
		return nil, err
	}
	cfg.Address = valueAddress

	valueStoreInterval, err := getIntValueAllowZero(&storeInterval, "STORE_INTERVAL")
	if err != nil {
		return nil, err
	}
	cfg.StoreInterval = time.Duration(valueStoreInterval) * time.Second

	valueFilePath, err := getStringValue(&filePath, "FILE_STORAGE_PATH")
	if err != nil {
		return nil, err
	}
	cfg.FileStoragePath = valueFilePath

	valueRestore, err := getBoolValue(&restore, "RESTORE")
	if err != nil {
		return nil, err
	}
	cfg.Restore = valueRestore

	return &cfg, nil
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

func getIntValueAllowZero(flagValue *int, envName string) (int, error) {
	envValue := os.Getenv(envName)

	if envValue != "" {
		val, err := strconv.Atoi(envValue)
		if err != nil {
			return 0, fmt.Errorf("invalid value for environment variable %s: %w", envName, err)
		}
		if val < 0 {
			return 0, fmt.Errorf("%s must be >= 0", envName)
		}
		return val, nil
	}

	if flagValue == nil || *flagValue < 0 {
		return 0, fmt.Errorf("%s must be >= 0", envName)
	}

	return *flagValue, nil
}

func getBoolValue(flagValue *bool, envName string) (bool, error) {
	envValue := os.Getenv(envName)
	if envValue != "" {
		val, err := strconv.ParseBool(envValue)
		if err != nil {
			return false, fmt.Errorf("invalid value for environment variable %s: %w", envName, err)
		}
		return val, nil
	}

	if flagValue == nil {
		return false, fmt.Errorf("%s must be set", envName)
	}

	return *flagValue, nil
}
