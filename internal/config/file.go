package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func SetFile(f string) {
	file = f
}

func FromFile() (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %s: %w", file, err)
	}
	defer f.Close()

	cfg := newDefaultConfig()
	err = json.NewDecoder(f).Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file %s: %w", file, err)
	}

	cfg.CreatedAt = time.Now()
	cfg.UpdatedAt = time.Now()

	return cfg, nil
}

var file string
