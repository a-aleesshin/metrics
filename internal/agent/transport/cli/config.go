package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type AgentConfig struct {
	Address        string `env:"ADDRESS"`
	ReportInterval time.Duration
	PollInterval   time.Duration
}

var (
	addressDefault        = "localhost:8080"
	reportIntervalDefault = 10
	pollIntervalDefault   = 2
)

func LoadConfig(args []string) (*AgentConfig, error) {
	var config AgentConfig

	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	var address string
	var reportInterval int
	var pollInterval int

	fs.StringVar(&address, "a", addressDefault, "HTTP server address")
	fs.IntVar(&reportInterval, "r", reportIntervalDefault, "report interval in seconds")
	fs.IntVar(&pollInterval, "p", pollIntervalDefault, "poll interval in seconds")

	err := fs.Parse(args)

	if err != nil {
		return nil, fmt.Errorf("failed to parse command line arguments: %w", err)
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

	return &config, nil
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
