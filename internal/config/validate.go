package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"
)

func (c *Config) validate() error {
	err := validateProjectName(c.ProjectName)
	if err != nil {
		return fmt.Errorf("invalid project name: %w", err)
	}

	err = validateTimestamps(c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("invalid timestamps: %w", err)
	}

	err = validateMinioEndpoint(c.MinioEndpoint)
	if err != nil {
		return fmt.Errorf("invalid minio endpoint: %w", err)
	}

	err = validateMinioBucketName(c.MinioBucketName)
	if err != nil {
		return fmt.Errorf("invalid minio bucket name: %w", err)
	}

	err = validateMinioLinkExpiry(c.MinioLinkExpiry)
	if err != nil {
		return fmt.Errorf("invalid minio link expiry: %w", err)
	}

	err = validateYOURLSEndpoint(c.YOURLSEndpoint)
	if err != nil {
		return fmt.Errorf("invalid yourls endpoint: %w", err)
	}

	return nil
}

const (
	minProjectNameLength = 4
	maxProjectNameLength = 16
)

func validateProjectName(name string) error {
	if name == "" {
		return errors.New("project name cannot be empty")
	}

	if len(name) < minProjectNameLength || len(name) > maxProjectNameLength {
		return fmt.Errorf(
			"project name must be between %d and %d characters long, got %d",
			minProjectNameLength,
			maxProjectNameLength,
			len(name),
		)
	}

	for _, char := range name {
		if !unicode.IsLetter(char) {
			return fmt.Errorf("project name must only contain letters, got '%c'", char)
		}

		if !unicode.IsLower(char) {
			return fmt.Errorf("project name must only contain lowercase letters, got '%c'", char)
		}
	}

	return nil
}

func validateTimestamps(createdAt time.Time, updatedAt time.Time) error {
	if createdAt.IsZero() {
		return errors.New("created_at cannot be zero")
	}

	if updatedAt.IsZero() {
		return errors.New("updated_at cannot be zero")
	}

	if updatedAt.Before(createdAt) {
		return errors.New("updated_at cannot be before created_at")
	}

	return nil
}

func validateMinioEndpoint(endpoint string) error {
	if endpoint == "" {
		return errors.New("minio_endpoint cannot be empty")
	}

	if strings.Contains(endpoint, "://") {
		return errors.New("minio_endpoint should not contain a scheme")
	}

	return nil
}

const (
	minMinioBucketNameLength = 4
	maxMinioBucketNameLength = 36
)

func validateMinioBucketName(bucketName string) error {
	if bucketName == "" {
		return errors.New("minio_bucket_name cannot be empty")
	}

	if len(bucketName) < minMinioBucketNameLength || len(bucketName) > maxMinioBucketNameLength {
		return fmt.Errorf(
			"minio_bucket_name must be between %d and %d characters long, got %d",
			minMinioBucketNameLength,
			maxMinioBucketNameLength,
			len(bucketName),
		)
	}

	for _, char := range bucketName {
		if !unicode.IsLetter(char) {
			return fmt.Errorf("minio_bucket_name must only contain letters, got '%c'", char)
		}

		if !unicode.IsLower(char) {
			return fmt.Errorf(
				"minio_bucket_name must only contain lowercase letters, got '%c'",
				char,
			)
		}
	}

	return nil
}

const (
	minMinioLinkExpiry = 1 * time.Hour
	maxMinioLinkExpiry = 7 * 24 * time.Hour
)

func validateMinioLinkExpiry(expiry time.Duration) error {
	if expiry < minMinioLinkExpiry || expiry > maxMinioLinkExpiry {
		return fmt.Errorf(
			"minio_link_expiry must be between %s and %s, got %s",
			minMinioLinkExpiry,
			maxMinioLinkExpiry,
			expiry,
		)
	}

	return nil
}

func validateYOURLSEndpoint(endpoint *url.URL) error {
	if endpoint == nil {
		return errors.New("yourls_endpoint cannot be nil")
	}

	if endpoint.Scheme != "http" && endpoint.Scheme != "https" {
		return fmt.Errorf("yourls_endpoint must use http or https, got %s", endpoint.Scheme)
	}

	return nil
}
