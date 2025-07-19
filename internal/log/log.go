package log

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func Setup() error {
	if file != nil || logger != nil {
		return errors.New("log already initialized")
	}

	//nolint:reassign // We intentionally reassign this for performance reasons.
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	console := zerolog.ConsoleWriter{
		Out:     os.Stderr,
		NoColor: noColor(),
	}

	var err error
	file, err = createFile()
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	multi := zerolog.MultiLevelWriter(console, file)
	newLogger := zerolog.New(multi).With().Timestamp().Logger()

	if noConsole() {
		newLogger = newLogger.Output(file)
	}

	leveled := newLogger.Level(level())
	logger = &leveled

	return nil
}

func Flush() error {
	if file == nil || logger == nil {
		return errors.New("log not initialized")
	}

	err := file.Sync()
	if err != nil {
		return fmt.Errorf("failed to flush log file: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}

	return nil
}

func Logger() *zerolog.Logger {
	// Return a no-op logger if logging is suppressed.
	if suppressLog {
		l := zerolog.Nop()
		return &l
	}

	// This should never happen.
	if logger == nil {
		panic("log not initialized")
	}

	return logger
}

func CleanOld() error {
	if file == nil || logger == nil {
		return errors.New("log not initialized")
	}

	logsDir, err := setupLogsDir()
	if err != nil {
		return fmt.Errorf("failed to setup logs directory: %w", err)
	}

	var files []os.DirEntry
	files, err = os.ReadDir(logsDir)
	if err != nil {
		return fmt.Errorf("failed to read logs directory: %w", err)
	}

	fileErrors := make([]error, 0)
	for _, file := range files {
		if file.IsDir() {
			fileErrors = append(fileErrors, fmt.Errorf("unexpected directory %s", file.Name()))
			continue
		}

		filePath := filepath.Join(logsDir, file.Name())

		var info os.FileInfo
		info, err = file.Info()
		if err != nil {
			fileErrors = append(
				fileErrors,
				fmt.Errorf("failed to get info for file %s: %w", file.Name(), err),
			)
			continue
		}

		if time.Since(info.ModTime()) > logFileRetention {
			err = os.Remove(filePath)
			if err != nil {
				fileErrors = append(
					fileErrors,
					fmt.Errorf("failed to remove old log file %s: %w", file.Name(), err),
				)
			}
		}
	}

	if len(fileErrors) > 0 {
		var combinedError strings.Builder
		combinedError.WriteString("failed to clean old log files:\n")
		for _, err := range fileErrors {
			combinedError.WriteString(fmt.Sprintf("- %s\n", err.Error()))
		}

		return errors.New(combinedError.String())
	}

	return nil
}

func Suppress() {
	suppressLog = true
}

func Enable() {
	suppressLog = false
}

const logFileRetention = 7 * 24 * time.Hour

var (
	file   *os.File
	logger *zerolog.Logger
)

func noColor() bool {
	s := strings.ToLower(os.Getenv("MINLY_LOG_NO_COLOR"))
	if s == "" {
		return false
	}

	b, err := strconv.ParseBool(s)
	return err == nil && b
}

var fileTimestamp = time.Now().Format("2006-01-02_15-04-05")

func createFile() (*os.File, error) {
	logsDir, err := setupLogsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to setup logs directory: %w", err)
	}

	fileName := fmt.Sprintf("minly_%s.log.json", fileTimestamp)

	filePath := filepath.Join(logsDir, fileName)

	var f *os.File
	f, err = os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return f, nil
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

func noConsole() bool {
	s := strings.ToLower(os.Getenv("MINLY_LOG_NO_CONSOLE"))
	if s == "" {
		return false
	}

	b, err := strconv.ParseBool(s)
	return err == nil && b
}

func level() zerolog.Level {
	s := strings.ToLower(os.Getenv("MINLY_LOG_LEVEL"))

	switch s {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

var suppressLog = false
