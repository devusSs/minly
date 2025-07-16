package log

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func Setup() error {
	if file != nil {
		return errors.New("log file already set up")
	}

	err := createFile()
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	output := io.MultiWriter(os.Stderr, file)

	if noConsole() {
		output = file
	}

	logger := slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: getLevel(),
	}))

	slog.SetDefault(logger)

	return nil
}

func Flush() error {
	if file == nil {
		return errors.New("no log file to flush")
	}

	err := file.Sync()
	if err != nil {
		return fmt.Errorf("failed to flush log file: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}

	file = nil

	return nil
}

func getLevel() slog.Level {
	s := os.Getenv("MINLY_LOG_LEVEL")
	s = strings.ToLower(s)

	switch s {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func noConsole() bool {
	s := os.Getenv("MINLY_LOG_NO_CONSOLE")
	b, err := strconv.ParseBool(s)
	return err == nil && b
}

var (
	fileTimestamp = time.Now().Format("2006-01-02_15-04-05")
	file          *os.File
)

func createFile() error {
	logsDir, err := setupLogsDir()
	if err != nil {
		return fmt.Errorf("failed to setup logs directory: %w", err)
	}

	fileName := fmt.Sprintf("minly_%s.log.json", fileTimestamp)

	filePath := filepath.Join(logsDir, fileName)

	file, err = os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	return nil
}

func setupLogsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	logsDir := filepath.Join(home, ".minly", "logs")

	err = os.MkdirAll(logsDir, 0700)
	if err != nil {
		return "", fmt.Errorf("failed to create logs directory: %w", err)
	}

	return logsDir, nil
}
