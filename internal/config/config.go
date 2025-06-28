package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

// Errors related to the config package.
var (
	ErrMissingProjectName = errors.New("project name is required")
	ErrFileNotExists      = errors.New("config file does not exist")
)

// GetSaved retrieves all saved configurations.
// Returns a slice of Config objects or an error if any operation fails.
//
// If no configurations are found, it returns an empty slice.
func GetSaved() ([]*Config, error) {
	configDir, err := getDir()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory %s: %w", configDir, err)
	}

	configs := make([]*Config, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			return nil, fmt.Errorf(
				"unexpected directory found in config directory: %s",
				file.Name(),
			)
		}

		projectName := strings.Split(file.Name(), filepath.Ext(file.Name()))[0]

		var cfg *Config
		cfg, err = LoadConfig(projectName)
		if err != nil {
			return nil, fmt.Errorf("failed to load config for project %s: %w", projectName, err)
		}

		configs = append(configs, cfg)
	}

	return configs, nil
}

// Config represents the configuration for a minly project.
type Config struct {
	ProjectName            string    `json:"project_name"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
	LogsDirectory          string    `json:"logs_directory"`
	StoragesDirectory      string    `json:"storages_directory"`
	SecretsServiceName     string    `json:"secrets_service_name"`
	MinioPublicBucketName  string    `json:"minio_public_bucket_name"`
	MinioPrivateBucketName string    `json:"minio_private_bucket_name"`
	YOURLSDescription      string    `json:"yourls_description"`
}

// String implements the Stringer interface for Config.
func (c *Config) String() string {
	return fmt.Sprintf("%+v", *c)
}

// NewConfig creates a new Config object with the given parameters.
func NewConfig(projectName string) (*Config, error) {
	if projectName == "" {
		return nil, ErrMissingProjectName
	}

	logsDir, err := getDefaultLogsDir(projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get default logs directory: %w", err)
	}

	var storagesDir string
	storagesDir, err = getDefaultStoragesDir(projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get default storages directory: %w", err)
	}

	c := &Config{
		ProjectName:            projectName,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		LogsDirectory:          logsDir,
		StoragesDirectory:      storagesDir,
		SecretsServiceName:     getDefaultSecretsServiceName(),
		MinioPublicBucketName:  fmt.Sprintf("%s-public", projectName),
		MinioPrivateBucketName: fmt.Sprintf("%s-private", projectName),
		YOURLSDescription: fmt.Sprintf(
			"Uploaded using github.com/devusSs/minly for project %s",
			projectName,
		),
	}

	err = validateConfig(c)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return c, nil
}

// WriteConfig writes the given Config to a file.
//
// Returns an error if the config cannot be written or the file cannot be created.
func WriteConfig(c *Config) error {
	if c == nil {
		return errors.New("config cannot be nil")
	}

	err := validateConfig(c)
	if err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	var configsDir string
	configsDir, err = getDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	configFilePath := filepath.Join(configsDir, fmt.Sprintf("%s.json", c.ProjectName))

	var f *os.File
	f, err = os.Create(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to create config file %s: %w", configFilePath, err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(c)
	if err != nil {
		return fmt.Errorf("failed to write config to file %s: %w", configFilePath, err)
	}

	return nil
}

// LoadConfig loads the configuration for a given project name from the config file.
// Returns a Config object or an error if the file cannot be read or parsed.
func LoadConfig(projectName string) (*Config, error) {
	if projectName == "" {
		return nil, ErrMissingProjectName
	}

	configsDir, err := getDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	f, err := getFile(configsDir, projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get config file for project %s: %w", projectName, err)
	}
	defer f.Close()

	c := &Config{}
	err = json.NewDecoder(f).Decode(c)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file %s: %w", f.Name(), err)
	}

	err = validateConfig(c)
	if err != nil {
		return nil, fmt.Errorf("failed to validate loaded config: %w", err)
	}

	return c, nil
}

const (
	minlyDirectory   = ".minly"
	configsDirectory = "configs"
)

// getDir retrieves the directory where the minly configuration files are stored.
//
// Returns the directory path or an error if the directory cannot be accessed.
// If the directory does not exist, it will be created with appropriate permissions.
func getDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(home, minlyDirectory, configsDirectory)

	_, err = os.Stat(configDir)
	if err != nil {
		// If the directory does not exist, create it.
		if os.IsNotExist(err) {
			err = os.MkdirAll(configDir, 0700)
			if err != nil {
				return "", fmt.Errorf("failed to create config directory %s: %w", configDir, err)
			}
		} else {
			return "", fmt.Errorf("failed to access config directory %s: %w", configDir, err)
		}
	}

	return configDir, nil
}

// getFile retrieves the file for a given project name in the specified directory.
// Returns the file handle or an error if the file does not exist.
func getFile(dir string, projectName string) (*os.File, error) {
	configFilePath := filepath.Join(dir, fmt.Sprintf("%s.json", projectName))

	_, err := os.Stat(configFilePath)
	if err != nil {
		// Any other error may be disregarded as the file not existing.
		return nil, fmt.Errorf("%s: %w", configFilePath, ErrFileNotExists)
	}

	f, err := os.Open(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %s: %w", configFilePath, err)
	}

	return f, nil
}

// getDefaultLogsDir retrieves the default logs directory for a given project name.
// Returns the logs directory path or an error if any operation fails.
//
// If the directory does not exist, it will be created with appropriate permissions.
func getDefaultLogsDir(projectName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	logsDir := filepath.Join(home, minlyDirectory, "logs", projectName)

	err = os.MkdirAll(logsDir, 0700)
	if err != nil {
		return "", fmt.Errorf("failed to create logs directory %s: %w", logsDir, err)
	}

	return logsDir, nil
}

// getDefaultStoragesDir retrieves the default storage directory for a given project name.
// Returns the storage directory path or an error if any operation fails.
//
// If the directory does not exist, it will be created with appropriate permissions.
func getDefaultStoragesDir(projectName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	storageDir := filepath.Join(home, minlyDirectory, "storages", projectName)

	err = os.MkdirAll(storageDir, 0700)
	if err != nil {
		return "", fmt.Errorf("failed to create storage directory %s: %w", storageDir, err)
	}

	return storageDir, nil
}

const defaultSecretsServiceName = "minly"

func getDefaultSecretsServiceName() string {
	return defaultSecretsServiceName
}

const (
	projectNameMinLength = 3
	projectMaxNameLength = 16
)

// validateConfig validates the given Config object.
// It checks for required fields and sets default values if necessary.
// Returns an error if the config is invalid.
func validateConfig(c *Config) error {
	if c == nil {
		return errors.New("config cannot be nil")
	}

	if c.ProjectName == "" {
		return ErrMissingProjectName
	}

	err := validateLowercaseString(c.ProjectName, projectNameMinLength, projectMaxNameLength)
	if err != nil {
		return fmt.Errorf("invalid project name '%s': %w", c.ProjectName, err)
	}

	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}

	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = time.Now()
	}

	c.LogsDirectory, err = getDefaultLogsDir(c.ProjectName)
	if err != nil {
		return fmt.Errorf("failed to get default logs directory: %w", err)
	}

	c.StoragesDirectory, err = getDefaultStoragesDir(c.ProjectName)
	if err != nil {
		return fmt.Errorf("failed to get default storages directory: %w", err)
	}

	c.SecretsServiceName = getDefaultSecretsServiceName()

	c.MinioPublicBucketName = fmt.Sprintf("%s-public", c.ProjectName)
	c.MinioPrivateBucketName = fmt.Sprintf("%s-private", c.ProjectName)

	c.YOURLSDescription = fmt.Sprintf(
		"Uploaded using github.com/devusSs/minly for project %s",
		c.ProjectName,
	)

	return err
}

func validateLowercaseString(s string, minLength int, maxLength int) error {
	if len(s) < minLength {
		return fmt.Errorf("invalid length: %d is less than minimum length %d", len(s), minLength)
	}

	if len(s) > maxLength {
		return fmt.Errorf("invalid length: %d is greater than maximum length %d", len(s), maxLength)
	}

	for _, r := range s {
		if !unicode.IsLetter(r) {
			return fmt.Errorf("invalid character: %q is not a letter", r)
		}

		if !unicode.IsLower(r) {
			return fmt.Errorf("invalid character: %q is not lowercase", r)
		}
	}

	return nil
}
