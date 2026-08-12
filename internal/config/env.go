package config

import (
	"errors"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

func SetEnvFile(file string) {
	envFile = file
}

func FromEnv() (*Config, error) {
	err := loadEnvFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load env file: %w", err)
	}

	cfg := newDefaultConfig()
	err = parseEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse env: %w", err)
	}

	return cfg, nil
}

var envFile string //nolint:gochecknoglobals // SetEnvFile configures the package's subsequent FromEnv call.

func loadEnvFile() error {
	if envFile == "" {
		return nil
	}

	err := godotenv.Load(envFile)
	if err != nil {
		return fmt.Errorf("failed to load %s: %w", envFile, err)
	}

	return nil
}

func parseEnv(cfg *Config) error {
	if cfg == nil {
		return errors.New("config cannot be nil")
	}

	opts := env.Options{Prefix: "MINLY_", RequiredIfNoDef: true}
	err := env.ParseWithOptions(cfg, opts)
	if err != nil {
		return fmt.Errorf("failed to parse with options: %w", err)
	}

	return nil
}
