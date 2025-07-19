package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type File struct {
	Timestamp  time.Time `json:"timestamp"`
	MinioLink  string    `json:"minio_link"`
	YOURLSLink string    `json:"yourls_link"`
}

func (f *File) String() string {
	return fmt.Sprintf("%+v", *f)
}

func NewFile(minioLink, yourlsLink string) *File {
	return &File{
		Timestamp:  time.Now(),
		MinioLink:  minioLink,
		YOURLSLink: yourlsLink,
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
			continue
		}

		fullPath := filepath.Join(fs.dir, entry.Name())

		var f *os.File
		f, err = os.Open(fullPath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var fobj File
			err = json.Unmarshal(scanner.Bytes(), &fobj)
			if err == nil {
				result = append(result, fobj)
			}
		}

		err = f.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close file %s: %w", fullPath, err)
		}
	}

	return result, nil
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
