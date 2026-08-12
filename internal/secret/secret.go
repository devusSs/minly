package secret

import (
	"errors"
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

func GetInput(prompt string) (string, error) {
	if !isTerminal() {
		return "", errors.New("stdin is not a readable terminal")
	}

	if len(prompt) > 0 && prompt[len(prompt)-1] != ':' {
		prompt += ":"
		prompt += " "
	}

	_, err := fmt.Fprint(os.Stdout, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to write prompt: %w", err)
	}

	byteInput, err := term.ReadPassword(int(os.Stdin.Fd()))
	_, newlineErr := fmt.Fprintln(os.Stdout)
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	if newlineErr != nil {
		return "", fmt.Errorf("failed to write newline: %w", newlineErr)
	}

	return string(byteInput), nil
}

type Key string

const (
	MinioAccessKey    Key = "minio_access_key"
	MinioAccessSecret Key = "minio_access_secret"
	YOURLSignature    Key = "yourl_signature"
)

func Exists(key Key) (bool, error) {
	_, err := keyring.Get("minly", string(key))
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("failed to check if key exists: %w", err)
	}

	return true, nil
}

func Load(key Key) (string, error) {
	value, err := keyring.Get("minly", string(key))
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", fmt.Errorf("key not found: %s", key)
		}

		return "", fmt.Errorf("failed to load key: %w", err)
	}

	return value, nil
}

func Save(key Key, value string) error {
	if value == "" {
		return errors.New("value cannot be empty")
	}

	err := keyring.Set("minly", string(key), value)
	if err != nil {
		return fmt.Errorf("failed to save key: %w", err)
	}

	return nil
}

func DeleteAll() error {
	err := keyring.DeleteAll("minly")
	if err != nil {
		return fmt.Errorf("failed to delete all keys: %w", err)
	}

	return nil
}

func isTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
