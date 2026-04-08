package cli

import (
	"flag"
	"fmt"
	"io"
)

type ServerConfig struct {
	Address string
}

func ParseConfig(args []string) (ServerConfig, error) {
	cfg := ServerConfig{
		Address: "localhost:8080",
	}

	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.StringVar(&cfg.Address, "a", cfg.Address, "HTTP server address")

	if err := fs.Parse(args); err != nil {
		return ServerConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}
