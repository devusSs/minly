package lastrun

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type LastRun struct {
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error"`
}

func (l *LastRun) String() string {
	return fmt.Sprintf("%+v", *l)
}

func Write(runErr error) error {
	file, err := createLastRunFile()
	if err != nil {
		return err
	}
	defer file.Close()

	lastRun := LastRun{
		Timestamp: time.Now(),
		Error:     "",
	}

	if runErr != nil {
		lastRun.Error = runErr.Error()
	}

	err = json.NewEncoder(file).Encode(lastRun)
	if err != nil {
		return fmt.Errorf("failed to encode last run data: %w", err)
	}

	return nil
}

func Read() (*LastRun, error) {
	file, err := openLastRunFile()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lastRun LastRun
	err = json.NewDecoder(file).Decode(&lastRun)
	if err != nil {
		return nil, fmt.Errorf("failed to decode last run file: %w", err)
	}

	return &lastRun, nil
}

func createLastRunFile() (*os.File, error) {
	lastRunDir, err := setupLastRunDir()
	if err != nil {
		return nil, err
	}

	lastRunFilePath := filepath.Join(lastRunDir, "last_run.json")

	file, err := os.Create(lastRunFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create last run file: %w", err)
	}

	return file, nil
}

func openLastRunFile() (*os.File, error) {
	lastRunDir, err := setupLastRunDir()
	if err != nil {
		return nil, err
	}

	lastRunFilePath := filepath.Join(lastRunDir, "last_run.json")

	file, err := os.Open(lastRunFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open last run file: %w", err)
	}

	return file, nil
}

func setupLastRunDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	lastRunDir := filepath.Join(home, ".minly", "lastrun")

	err = os.MkdirAll(lastRunDir, 0700)
	if err != nil {
		return "", fmt.Errorf("failed to create last run directory: %w", err)
	}

	return lastRunDir, nil
}
