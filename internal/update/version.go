package update

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

func newVersionAvailable(currentVersion string, latestVersion string) (bool, error) {
	current, err := semver.NewVersion(currentVersion)
	if err != nil {
		return false, fmt.Errorf("invalid current version: %w", err)
	}

	var latest *semver.Version
	latest, err = semver.NewVersion(latestVersion)
	if err != nil {
		return false, fmt.Errorf("invalid latest version: %w", err)
	}

	return latest.GreaterThan(current), nil
}
