package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID               string    `json:"id"`
	Timestamp        time.Time `json:"timestamp"`
	MinioLink        string    `json:"minio_link"`
	MinioLinkExpires time.Time `json:"minio_link_expires"`
	YOURLSLink       string    `json:"yourls_link"`
}

func (f *File) String() string {
	return fmt.Sprintf("%+v", *f)
}

func (f *File) validate() error {
	if f.ID == "" {
		return errors.New("id is required")
	}

	if f.Timestamp.IsZero() {
		return errors.New("timestamp is required")
	}

	if f.MinioLink == "" {
		return errors.New("minio_link is required")
	}

	if f.MinioLinkExpires.IsZero() {
		return errors.New("minio_link_expires is required")
	}

	if f.YOURLSLink == "" {
		return errors.New("yourls_link is required")
	}

	return nil
}

func NewFile(minioLink string, minioLinkExpires time.Time, yourlsLink string) *File {
	return &File{
		ID:               uuid.NewString(),
		Timestamp:        time.Now(),
		MinioLink:        minioLink,
		MinioLinkExpires: minioLinkExpires,
		YOURLSLink:       yourlsLink,
	}
}

type FileStore struct {
	dir string
	mu  sync.Mutex
}

func (fs *FileStore) String() string {
	return fmt.Sprintf("FileStore{dir: %s}", fs.dir)
}

func NewFileStore() (*FileStore, error) {
	dir, err := getStorageDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage directory: %w", err)
	}

	return &FileStore{dir: dir}, nil
}

func (fs *FileStore) Save(file *File) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if file == nil {
		return errors.New("file cannot be nil")
	}

	err := file.validate()
	if err != nil {
		return fmt.Errorf("file validation failed: %w", err)
	}

	filename := filepath.Join(fs.dir, file.Timestamp.Format("2006-01")+".jsonl")

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(file)
	if err != nil {
		return fmt.Errorf("failed to encode file %s: %w", filename, err)
	}

	return nil
}

func (fs *FileStore) LoadAll() ([]File, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	files, err := os.ReadDir(fs.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", fs.dir, err)
	}

	var result []File

	for _, entry := range files {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			return nil, fmt.Errorf("invalid file %s in storage directory", entry.Name())
		}

		fullPath := filepath.Join(fs.dir, entry.Name())

		var f *os.File
		f, err = os.Open(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", fullPath, err)
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var fobj File
			err = json.Unmarshal(scanner.Bytes(), &fobj)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal file %s: %w", fullPath, err)
			}

			if fobj.ID == "" {
				fobj.ID = uuid.NewString()
			}

			err = fobj.validate()
			if err != nil {
				return nil, fmt.Errorf("file validation failed for %s: %w", fullPath, err)
			}

			result = append(result, fobj)
		}

		err = f.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close file %s: %w", fullPath, err)
		}
	}

	return result, nil
}

//nolint:gocognit // This was vibe-coded and might be changed in the future.
func (fs *FileStore) CleanOldFiles() (int, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	files, err := os.ReadDir(fs.dir)
	if err != nil {
		return 0, fmt.Errorf("failed to read storage dir: %w", err)
	}

	now := time.Now()
	totalDeleted := 0

	for _, entry := range files {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}

		fullPath := filepath.Join(fs.dir, entry.Name())

		var f *os.File
		f, err = os.Open(fullPath)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to open file %s: %w", fullPath, err)
		}

		var keep []File
		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			var fobj File
			err = json.Unmarshal(scanner.Bytes(), &fobj)
			if err != nil {
				_ = f.Close()
				return totalDeleted, fmt.Errorf("failed to unmarshal file %s: %w", fullPath, err)
			}

			if fobj.MinioLinkExpires.After(now) {
				keep = append(keep, fobj)
			} else {
				totalDeleted++
			}
		}
		err = f.Close()
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to close file %s: %w", fullPath, err)
		}

		if len(keep) < 1 {
			err = os.Remove(fullPath)
			if err != nil {
				return totalDeleted, fmt.Errorf(
					"failed to remove expired file %s: %w",
					fullPath,
					err,
				)
			}
			continue
		}

		tmpPath := fullPath + ".tmp"

		var tmpFile *os.File
		tmpFile, err = os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to create temp file for %s: %w", fullPath, err)
		}

		enc := json.NewEncoder(tmpFile)
		for _, f := range keep {
			err = enc.Encode(f)
			if err != nil {
				_ = tmpFile.Close()
				return totalDeleted, fmt.Errorf("failed to encode retained file: %w", err)
			}
		}
		_ = tmpFile.Close()

		err = os.Rename(tmpPath, fullPath)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to replace original file %s: %w", fullPath, err)
		}
	}

	return totalDeleted, nil
}

func getStorageDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	storageDir := filepath.Join(home, ".minly", "storage")

	err = os.MkdirAll(storageDir, 0700)
	if err != nil {
		return "", fmt.Errorf("failed to create storage directory %s: %w", storageDir, err)
	}

	return storageDir, nil
}
