package postgres

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

type Config struct {
	Host     string
	Port     uint16
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func newConfig(host string, port uint16, user string, password string, dbName string, sslMode string) (*Config, error) {
	if port == 0 {
		port = 5432
	}

	if sslMode == "" {
		sslMode = "disable"
	}

	return validatedConfig(&Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
		SSLMode:  sslMode,
	})
}

func validatedConfig(cfg *Config) (*Config, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("host cannot be empty")
	}

	if cfg.Port == 0 {
		return nil, fmt.Errorf("port cannot be empty")
	}

	if cfg.User == "" {
		return nil, fmt.Errorf("user cannot be empty")
	}

	if cfg.DBName == "" {
		return nil, fmt.Errorf("db name cannot be empty")
	}

	switch cfg.SSLMode {
	case "disable", "require", "verify-ca", "verify-full", "no-verify":
		return cfg, nil
	default:
		return nil, fmt.Errorf("invalid ssl mode: %s", cfg.SSLMode)
	}
}

func NewConfigFromString(dsn string) (*Config, error) {
	// patterns
	// postgres://user:password@host:5432/dbname?sslmode=disable
	// host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable
	if dsn == "" {
		return nil, fmt.Errorf("DSN cannot be empty")
	}

	urlParse, err := url.Parse(dsn)
	sslMode := "disable"

	if err == nil && strings.HasPrefix(dsn, "postgres") {
		if urlParse.Host == "" || urlParse.Path == "" {
			return nil, fmt.Errorf("dsn must contain host and db name")
		}

		if sslmode := urlParse.Query().Get("sslmode"); sslmode != "" {
			sslMode = sslmode
		}
	} else if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	pc, err := pgconn.ParseConfig(dsn)

	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	return newConfig(pc.Host, pc.Port, pc.User, pc.Password, pc.Database, sslMode)
}

func (c *Config) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}
