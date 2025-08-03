package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

func saveAssetToFile(ctx context.Context, a *asset) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for asset %s: %w", a.url, err)
	}

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download asset %s: %w", a.url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download asset %s: %s", a.url, resp.Status)
	}

	var u *url.URL
	u, err = url.Parse(a.url)
	if err != nil {
		return "", fmt.Errorf("failed to parse asset URL %s: %w", a.url, err)
	}

	name := path.Base(u.Path)

	var updatesDir string
	updatesDir, err = setupUpdatesDir()
	if err != nil {
		return "", fmt.Errorf("failed to setup updates directory: %w", err)
	}

	filePath := filepath.Join(updatesDir, name)

	var f *os.File
	f, err = os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save asset %s to file: %w", a.url, err)
	}

	return filePath, nil
}

func setupUpdatesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	updatesDir := filepath.Join(home, ".minly", "updates")

	err = os.MkdirAll(updatesDir, 0700)
	if err != nil {
		return "", fmt.Errorf("failed to create updates directory: %w", err)
	}

	return updatesDir, nil
}
