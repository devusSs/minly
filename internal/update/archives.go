package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func unpackArchive(archivePath string) (string, error) {
	ext := filepath.Ext(archivePath)

	switch ext {
	case ".gz":
		return untarGzArchive(archivePath)
	case ".zip":
		return unzipArchive(archivePath)
	default:
		return "", fmt.Errorf("unsupported archive format: %s", ext)
	}
}

//nolint:gocognit // Might be changed in the future but it works for now.
func untarGzArchive(archivePath string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tarReader := tar.NewReader(gzr)
	destDir := filepath.Dir(archivePath)
	var appPath string

	for {
		var header *tar.Header
		header, err = tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar entry: %w", err)
		}

		if header.Typeflag != tar.TypeReg || !strings.Contains(header.Name, "minly") {
			continue
		}

		var outPath string
		if isExecutable(header.Name) {
			outPath = filepath.Join(destDir, "minly")
		} else {
			outPath = filepath.Join(destDir, filepath.Base(header.Name))
		}

		var uval uint32
		uval, err = safeInt64ToUint32(header.Mode)
		if err != nil {
			return "", fmt.Errorf("invalid file size for %s: %w", header.Name, err)
		}

		var outFile *os.File
		outFile, err = os.OpenFile(
			outPath,
			os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
			os.FileMode(uval),
		)
		if err != nil {
			return "", fmt.Errorf("failed to create output file: %w", err)
		}
		defer outFile.Close()

		//nolint:gosec // Copying file content is safe here.
		_, err = io.Copy(outFile, tarReader)
		if err != nil {
			return "", fmt.Errorf("failed to copy tar content: %w", err)
		}

		if isExecutable(header.Name) {
			appPath = outPath
		}
	}

	if appPath == "" {
		return "", errors.New("minly application not found in tar.gz")
	}

	return appPath, nil
}

func unzipArchive(archivePath string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	destDir := filepath.Dir(archivePath)
	var appPath string

	for _, f := range r.File {
		if !strings.Contains(f.Name, "minly") || f.FileInfo().IsDir() {
			continue
		}

		isApp := isExecutable(f.Name)

		var outPath string
		if isApp {
			outPath = filepath.Join(destDir, "minly")
		} else {
			outPath = filepath.Join(destDir, filepath.Base(f.Name))
		}

		var srcFile io.ReadCloser
		srcFile, err = f.Open()
		if err != nil {
			return "", fmt.Errorf("failed to open file in zip: %w", err)
		}
		defer srcFile.Close()

		var dstFile *os.File
		dstFile, err = os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", fmt.Errorf("failed to create destination file: %w", err)
		}
		defer dstFile.Close()

		//nolint:gosec // Copying file content is safe here.
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return "", fmt.Errorf("failed to copy file content: %w", err)
		}

		if isApp {
			appPath = outPath
		}
	}

	if appPath == "" {
		return "", errors.New("minly application not found in archive")
	}

	return appPath, nil
}

func isExecutable(name string) bool {
	base := filepath.Base(name)
	return strings.Contains(base, "minly") &&
		(strings.HasSuffix(base, ".exe") || !strings.Contains(base, "."))
}

func safeInt64ToUint32(v int64) (uint32, error) {
	if v < 0 || v > int64(^uint32(0)) {
		return 0, fmt.Errorf("value %d out of range for uint32", v)
	}
	return uint32(v), nil
}
