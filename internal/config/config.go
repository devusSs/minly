package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	ProjectName     string        `json:"project_name"      env:"PROJECT_NAME"      envDefault:"minly"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	MinioEndpoint   string        `json:"minio_endpoint"    env:"MINIO_ENDPOINT"    envDefault:"localhost:9000"`
	MinioUseSSL     bool          `json:"minio_use_ssl"     env:"MINIO_USE_SSL"     envDefault:"false"`
	MinioBucketName string        `json:"minio_bucket_name" env:"MINIO_BUCKET_NAME" envDefault:"minly"`
	MinioRegion     string        `json:"minio_region"      env:"MINIO_REGION"      envDefault:"us-east-1"`
	MinioLinkExpiry time.Duration `json:"minio_link_expiry" env:"MINIO_LINK_EXPIRY" envDefault:"24h"`
	YOURLSEndpoint  *url.URL      `json:"yourls_endpoint"   env:"YOURLS_ENDPOINT"   envDefault:"http://localhost:80/yourls-api.php"`

	filePath string
}

func (c *Config) String() string {
	return fmt.Sprintf("%+v", *c)
}

func (c *Config) FilePath() string {
	return c.filePath
}

func Read() (*Config, error) {
	f, err := openConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	cfg := newDefaultConfig()
	err = json.NewDecoder(f).Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	err = cfg.validate()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	cfg.filePath = f.Name()

	return cfg, nil
}

func Write(cfg *Config) error {
	if cfg == nil {
		return errors.New("config cannot be nil")
	}

	err := cfg.validate()
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	cfg.UpdatedAt = time.Now()

	var f *os.File
	f, err = createConfigFile()
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(cfg)
	if err != nil {
		return fmt.Errorf("failed to encode config file: %w", err)
	}

	return nil
}

func openConfigFile() (*os.File, error) {
	configDir, err := setupConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to setup config directory: %w", err)
	}

	configFilePath := filepath.Join(configDir, "config.json")

	var f *os.File
	f, err = os.Open(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}

	return f, nil
}

func createConfigFile() (*os.File, error) {
	configDir, err := setupConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to setup config directory: %w", err)
	}

	configFilePath := filepath.Join(configDir, "config.json")

	var f *os.File
	f, err = os.Create(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config file: %w", err)
	}

	return f, nil
}

func setupConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(home, ".minly", "config")

	err = os.MkdirAll(configDir, 0700)
	if err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configDir, nil
}

func newDefaultConfig() *Config {
	return &Config{
		ProjectName:     "minly",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		MinioEndpoint:   "localhost:9000",
		MinioUseSSL:     false,
		MinioBucketName: "minly",
		MinioRegion:     "us-east-1",
		MinioLinkExpiry: minMinioLinkExpiry,
		YOURLSEndpoint:  &url.URL{Scheme: "http", Host: "localhost:80", Path: "/yourls-api.php"},
	}
}
