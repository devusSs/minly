package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func saveAssetToFile(ctx context.Context, asset releaseAsset) (string, error) {
	if ctx == nil {
		return "", errors.New("context cannot be nil")
	}

	if asset.name == "" {
		return "", errors.New("asset name cannot be empty")
	}

	if asset.downloadURL == "" {
		return "", errors.New("asset download URL cannot be empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for asset %s: %w", asset.name, err)
	}

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download asset %s: %w", asset.name, err)
	}
	defer resp.Body.Close()

	var dir string
	dir, err = setupUpdatesDir()
	if err != nil {
		return "", fmt.Errorf("failed to set up updates directory: %w", err)
	}

	var f *os.File
	f, err = os.Create(filepath.Join(dir, asset.name))
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file for asset %s: %w", asset.name, err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy asset bytes to file %s: %w", f.Name(), err)
	}

	return f.Name(), nil
}

func setupUpdatesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	updatesDir := filepath.Join(home, ".minly", "updates")

	err = os.MkdirAll(updatesDir, 0700)
	if err != nil {
		return "", fmt.Errorf("failed to create updates directory %s: %w", updatesDir, err)
	}

	return updatesDir, nil
}

func unpackArchive(file string) (string, error) {
	if file == "" {
		return "", errors.New("file path cannot be empty")
	}

	switch {
	case strings.HasSuffix(file, ".tar.gz"):
		return untarGzipArchive(file)
	case strings.HasSuffix(file, ".zip"):
		return unzipArchive(file)
	default:
		return "", fmt.Errorf("unsupported archive format: %s", file)
	}
}

func untarGzipArchive(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("open .tar.gz: %w", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("gzip reader: %w", err)
	}
	defer gzr.Close()

	tarReader := tar.NewReader(gzr)

	basePath := strings.TrimSuffix(filepath.Base(file), ".tar.gz")
	destPath := filepath.Join(filepath.Dir(file), basePath)

	for {
		var hdr *tar.Header
		hdr, err = tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read tar: %w", err)
		}

		name := filepath.Base(hdr.Name)
		if name == "LICENSE" || name == "README.md" || hdr.FileInfo().IsDir() {
			continue
		}

		var outFile *os.File
		outFile, err = os.Create(destPath)
		if err != nil {
			return "", fmt.Errorf("create file: %w", err)
		}
		defer outFile.Close()

		//nolint:gosec // We release the archives, this should not happen.
		_, err = io.Copy(outFile, tarReader)
		if err != nil {
			return "", fmt.Errorf("copy content: %w", err)
		}

		//nolint: gosec // We release the archives, this should not happen.
		err = outFile.Chmod(os.FileMode(hdr.Mode))
		if err != nil {
			return "", fmt.Errorf("chmod: %w", err)
		}

		return destPath, nil
	}

	return "", errors.New("no valid program file found in tar.gz archive")
}

func unzipArchive(file string) (string, error) {
	r, err := zip.OpenReader(file)
	if err != nil {
		return "", fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	basePath := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	destPath := filepath.Join(filepath.Dir(file), basePath)

	for _, f := range r.File {
		name := filepath.Base(f.Name)
		if name == "LICENSE" || name == "README.md" || f.FileInfo().IsDir() {
			continue
		}

		var rc io.ReadCloser
		rc, err = f.Open()
		if err != nil {
			return "", fmt.Errorf("open zipped file: %w", err)
		}
		defer rc.Close()

		var outFile *os.File
		outFile, err = os.Create(destPath)
		if err != nil {
			return "", fmt.Errorf("create output file: %w", err)
		}
		defer outFile.Close()

		//nolint:gosec // We release the archives, this should not happen.
		_, err = io.Copy(outFile, rc)
		if err != nil {
			return "", fmt.Errorf("copy zipped content: %w", err)
		}

		err = outFile.Chmod(f.Mode())
		if err != nil {
			return "", fmt.Errorf("chmod: %w", err)
		}

		return destPath, nil
	}

	return "", errors.New("no valid file found in zip archive")
}
