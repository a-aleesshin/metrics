package cli

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"time"
)

type AgentConfig struct {
	Address        string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func ParseConfig(args []string) (AgentConfig, error) {
	cfg := AgentConfig{
		Address:        "localhost:8080",
		ReportInterval: 10 * time.Second,
		PollInterval:   2 * time.Second,
	}

	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var reportSec string
	var pollSec string

	fs.StringVar(&cfg.Address, "a", cfg.Address, "HTTP server address")
	fs.StringVar(&reportSec, "r", "10", "report interval in seconds")
	fs.StringVar(&pollSec, "p", "2", "poll interval in seconds")

	err := fs.Parse(args)

	if err != nil {
		return AgentConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

	r, err := strconv.Atoi(reportSec)

	if err != nil || r <= 0 {
		return AgentConfig{}, fmt.Errorf("failed to parse report interval: %w", err)
	}

	p, err := strconv.Atoi(pollSec)

	if err != nil || p <= 0 {
		return AgentConfig{}, fmt.Errorf("failed to parse poll interval: %w", err)

	}

	cfg.ReportInterval = time.Duration(r) * time.Second
	cfg.PollInterval = time.Duration(p) * time.Second

	return cfg, nil
}
