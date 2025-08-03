package update

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/selfupdate"
)

type Update struct {
	Updated   bool      `json:"updated"`
	Version   string    `json:"version"`
	Date      time.Time `json:"date"`
	Changelog string    `json:"changelog"`
}

func (u *Update) String() string {
	return fmt.Sprintf("%+v", *u)
}

//nolint:funlen // This function is designed to be an all-in-one update handler, so it may be this long.
func DoUpdate(ctx context.Context, currentVersion string) (*Update, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}

	if currentVersion == "" {
		return nil, errors.New("current version cannot be empty")
	}

	if currentVersion == "development" || currentVersion == "dev" {
		return nil, errors.New("cannot update from development version")
	}

	release, err := getLatestRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest release: %w", err)
	}

	var newAvailable bool
	newAvailable, err = newVersionAvailable(currentVersion, release.TagName)
	if err != nil {
		return nil, fmt.Errorf("failed to check for new version: %w", err)
	}

	if !newAvailable {
		return &Update{
			Updated:   false,
			Version:   release.TagName,
			Date:      release.PublishedAt,
			Changelog: release.Body,
		}, nil
	}

	var checksumsAsset *asset
	checksumsAsset, err = getChecksumsAsset(release)
	if err != nil {
		return nil, fmt.Errorf("failed to get checksums asset: %w", err)
	}

	var matchingAsset *asset
	matchingAsset, err = getMatchingAsset(release)
	if err != nil {
		return nil, fmt.Errorf("failed to get matching asset: %w", err)
	}

	var checksumsFilePath string
	checksumsFilePath, err = saveAssetToFile(ctx, checksumsAsset)
	if err != nil {
		return nil, fmt.Errorf("failed to save checksums asset: %w", err)
	}

	defer os.RemoveAll(filepath.Dir(checksumsFilePath))

	var matchingFilePath string
	matchingFilePath, err = saveAssetToFile(ctx, matchingAsset)
	if err != nil {
		return nil, fmt.Errorf("failed to save matching asset: %w", err)
	}

	var gotChecksum []byte
	gotChecksum, err = getChecksumForFile(matchingFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get checksum for asset file: %w", err)
	}

	var expectedChecksum []byte
	expectedChecksum, err = readChecksumForAssetFromFile(checksumsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read checksum for asset: %w", err)
	}

	if !bytes.Equal(gotChecksum, expectedChecksum) {
		return nil, fmt.Errorf(
			"checksum mismatch for asset %s: got %x, expected %x",
			matchingAsset,
			gotChecksum,
			expectedChecksum,
		)
	}

	var newExecutablePath string
	newExecutablePath, err = unpackArchive(matchingFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack archive: %w", err)
	}

	var f *os.File
	f, err = os.Open(newExecutablePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open new executable: %w", err)
	}
	defer f.Close()

	err = selfupdate.Apply(f, selfupdate.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to apply update: %w", err)
	}

	return &Update{
		Updated:   true,
		Version:   release.TagName,
		Date:      release.PublishedAt,
		Changelog: release.Body,
	}, nil
}
