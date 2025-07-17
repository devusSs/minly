package update

import (
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

func DoUpdate(ctx context.Context, currentVersion string) (*Update, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}

	if currentVersion == "" {
		return nil, errors.New("current version cannot be empty")
	}

	r, err := findLatestRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find latest release: %w", err)
	}

	var available bool
	available, err = r.isGreaterThan(currentVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to compare versions: %w", err)
	}

	if !available {
		return &Update{
			Updated:   false,
			Version:   currentVersion,
			Date:      time.Now(),
			Changelog: "No updates available",
		}, nil
	}

	var asset releaseAsset
	asset, err = findMatchingAsset(r.assets)
	if err != nil {
		return nil, fmt.Errorf("failed to find matching asset: %w", err)
	}

	var file string
	file, err = saveAssetToFile(ctx, asset)
	if err != nil {
		return nil, fmt.Errorf("failed to save asset to file: %w", err)
	}

	dir := filepath.Dir(file)
	defer os.RemoveAll(dir)

	var unpackedArchive string
	unpackedArchive, err = unpackArchive(file)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack archive: %w", err)
	}

	var f *os.File
	f, err = os.Open(unpackedArchive)
	if err != nil {
		return nil, fmt.Errorf("failed to open unpacked archive: %w", err)
	}
	defer f.Close()

	// TODO: we might want to integrate checksum checking
	err = selfupdate.Apply(f, selfupdate.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to apply update: %w", err)
	}

	return &Update{
		Updated:   true,
		Version:   r.version,
		Date:      r.createdAt,
		Changelog: r.changelog,
	}, nil
}
