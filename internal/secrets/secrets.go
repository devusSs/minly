package secrets

import (
	"github.com/zalando/go-keyring"
)

// Keys related to secrets management.
const (
	MinioAccessKeyKey    = "MINIO_ACCESS_KEY"
	MinioAccessSecretKey = "MINIO_ACCESS_SECRET"
	YOURLSSignatureKey   = "YOURLS_SIGNATURE"
)

// Load retrieves a secret for a given service and key.
//
// It uses keyring under the hood to securely retrieve secrets.
func Load(serviceName string, key string) (string, error) {
	value, err := keyring.Get(serviceName, key)
	if err != nil {
		return "", err
	}

	return value, nil
}

// Save stores a secret for a given service and key.
//
// It uses keyring under the hood to securely store secrets.
func Save(serviceName string, key string, value string) error {
	return keyring.Set(serviceName, key, value)
}

// Exists checks if a secret exists for a given service and key.
//
// It uses keyring to check for the existence of a secret.
func Exists(serviceName string, key string) (bool, error) {
	_, err := Load(serviceName, key)
	if err != nil {
		if err == keyring.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
