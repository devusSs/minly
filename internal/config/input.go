package config

import (
	"bufio"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

func FromInput() (*Config, error) {
	projectName, err := getProjectNameFromInput()
	if err != nil {
		return nil, fmt.Errorf("failed to get project name: %w", err)
	}

	var minioEndpoint string
	minioEndpoint, err = getMinioEndpointFromInput()
	if err != nil {
		return nil, fmt.Errorf("failed to get MinIO endpoint: %w", err)
	}

	var minioUseSSL bool
	minioUseSSL, err = getMinioUseSSLFromInput()
	if err != nil {
		return nil, fmt.Errorf("failed to get MinIO use SSL: %w", err)
	}

	var minioBucketName string
	minioBucketName, err = getMinioBucketNameFromInput()
	if err != nil {
		return nil, fmt.Errorf("failed to get MinIO bucket name: %w", err)
	}

	var minioRegion string
	minioRegion, err = getMinioRegionFromInput()
	if err != nil {
		return nil, fmt.Errorf("failed to get MinIO region: %w", err)
	}

	var minioLinkExpiry time.Duration
	minioLinkExpiry, err = getMinioLinkExpiryFromInput()
	if err != nil {
		return nil, fmt.Errorf("failed to get MinIO link expiry: %w", err)
	}

	var yourlsEndpoint *url.URL
	yourlsEndpoint, err = getYOURLSEndpointFromInput()
	if err != nil {
		return nil, fmt.Errorf("failed to get YOURLS endpoint: %w", err)
	}

	cfg := newDefaultConfig()

	cfg.ProjectName = projectName
	cfg.CreatedAt = time.Now()
	cfg.UpdatedAt = time.Now()
	cfg.MinioEndpoint = minioEndpoint
	cfg.MinioUseSSL = minioUseSSL
	cfg.MinioBucketName = minioBucketName
	cfg.MinioRegion = minioRegion
	cfg.MinioLinkExpiry = minioLinkExpiry
	cfg.YOURLSEndpoint = yourlsEndpoint

	return cfg, nil
}

func getProjectNameFromInput() (string, error) {
	generated, err := generateProjectName()
	if err != nil {
		return "", fmt.Errorf("failed to generate project name: %w", err)
	}

	var projectName string
	projectName, err = getInput("Enter project name", generated)
	if err != nil {
		return "", fmt.Errorf("failed to get project name from input: %w", err)
	}

	return projectName, nil
}

func generateProjectName() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz"

	lengthRange := maxProjectNameLength - minProjectNameLength + 1
	randomLengthOffset, err := rand.Int(rand.Reader, big.NewInt(int64(lengthRange)))
	if err != nil {
		return "", fmt.Errorf("failed to generate random length: %w", err)
	}

	length := minProjectNameLength + int(randomLengthOffset.Int64())

	result := make([]byte, length)

	for i := range result {
		var randomIndex *big.Int
		randomIndex, err = rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random character: %w", err)
		}

		result[i] = charset[randomIndex.Int64()]
	}

	return string(result), nil
}

func getMinioEndpointFromInput() (string, error) {
	endpoint, err := getInput("Enter MinIO endpoint", "localhost:9000")
	if err != nil {
		return "", fmt.Errorf("failed to get minio endpoint from input: %w", err)
	}

	if endpoint == "" {
		return "", errors.New("minio endpoint cannot be empty")
	}

	return endpoint, nil
}

func getMinioUseSSLFromInput() (bool, error) {
	useSSLStr, err := getInput("Use SSL for MinIO (true/false)", "false")
	if err != nil {
		return false, fmt.Errorf("failed to get minio use ssl from input: %w", err)
	}

	var useSSL bool
	useSSL, err = strconv.ParseBool(useSSLStr)
	if err != nil {
		return false, fmt.Errorf("failed to parse minio use ssl value: %w", err)
	}

	return useSSL, nil
}

func getMinioBucketNameFromInput() (string, error) {
	bucketName, err := getInput("Enter MinIO bucket name", "minly")
	if err != nil {
		return "", fmt.Errorf("failed to get minio bucket name from input: %w", err)
	}

	return bucketName, nil
}

func getMinioRegionFromInput() (string, error) {
	region, err := getInput("Enter MinIO region", "us-east-1")
	if err != nil {
		return "", fmt.Errorf("failed to get minio region from input: %w", err)
	}

	return region, nil
}

func getMinioLinkExpiryFromInput() (time.Duration, error) {
	linkExpiryStr, err := getInput(
		"Enter MinIO link expiry duration (e.g., 24h, 30m)",
		"24h",
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get minio link expiry from input: %w", err)
	}

	var linkExpiry time.Duration
	linkExpiry, err = time.ParseDuration(linkExpiryStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse minio link expiry duration: %w", err)
	}

	if linkExpiry <= 0 {
		return 0, errors.New("minio link expiry must be a positive duration")
	}

	return linkExpiry, nil
}

func getYOURLSEndpointFromInput() (*url.URL, error) {
	endpoint, err := getInput("Enter YOURLS endpoint", "http://localhost:80/yourls-api.php")
	if err != nil {
		return nil, fmt.Errorf("failed to get yourls endpoint from input: %w", err)
	}

	if endpoint == "" {
		return nil, errors.New("yourls endpoint cannot be empty")
	}

	var u *url.URL
	u, err = url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse yourls endpoint URL: %w", err)
	}

	return u, nil
}

func getInput(prompt string, def string) (string, error) {
	if !isTerminal() {
		return "", errors.New("stdin is not a readable terminal")
	}

	if prompt == "" {
		return "", errors.New("prompt cannot be empty")
	}

	prompt = prompt + " (default: " + def + ")"

	if prompt[len(prompt)-1] != ':' {
		prompt += ":"
	}

	if prompt[len(prompt)-1] != ' ' {
		prompt += " "
	}

	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read from stdin: %w", err)
	}

	trimmed := strings.TrimSpace(text)

	if trimmed == "" {
		trimmed = def
	}

	return trimmed, nil
}

func isTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
