package cli

import (
	"flag"
	"fmt"
	"os"
)

type ServerConfig struct {
	Address string `env:"ADDRESS"`
}

var (
	addressDefault = "localhost:8080"
)

func LoadConfig(args []string) (*ServerConfig, error) {
	var cfg ServerConfig

	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	var address string

	fs.StringVar(&address, "a", addressDefault, "HTTP server address")

	err := fs.Parse(args)

	if err != nil {
		return nil, fmt.Errorf("failed to parse command line arguments: %w", err)
	}

	valueAddress, err := getStringValue(&address, "ADDRESS")

	if err != nil {
		return nil, err
	}

	cfg.Address = valueAddress

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
