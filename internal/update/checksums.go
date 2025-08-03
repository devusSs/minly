package update

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

func getChecksumForFile(file string) ([]byte, error) {
	if file == "" {
		return nil, errors.New("file cannot be empty")
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	hasher := sha256.New()

	_, err = io.Copy(hasher, f)
	if err != nil {
		return nil, err
	}

	return hasher.Sum(nil), nil
}

const partsLength = 2

func readChecksumForAssetFromFile(checksumsFile string) ([]byte, error) {
	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH
	if targetArch == "amd64" {
		targetArch = "x86_64"
	}

	b, err := os.ReadFile(checksumsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read checksums file %s: %w", checksumsFile, err)
	}

	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != partsLength {
			return nil, fmt.Errorf("invalid checksum line: %s", line)
		}

		checksum := parts[0]
		fileName := strings.ToLower(parts[1])

		if strings.Contains(fileName, targetOS) && strings.Contains(fileName, targetArch) {
			checksumBytes, decodeErr := hex.DecodeString(checksum)
			if decodeErr != nil {
				return nil, fmt.Errorf("failed to decode hex checksum %s: %w", checksum, decodeErr)
			}
			return checksumBytes, nil
		}
	}

	return nil, fmt.Errorf(
		"no checksum found for OS %s and arch %s in checksums file %s",
		targetOS,
		targetArch,
		checksumsFile,
	)
}
