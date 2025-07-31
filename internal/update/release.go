package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

type release struct {
	version   string
	createdAt time.Time
	changelog string
	assets    []releaseAsset
}

func (r *release) isGreaterThan(currentVersion string) (bool, error) {
	latest, err := semver.NewVersion(r.version)
	if err != nil {
		return false, fmt.Errorf("failed to parse latest version %s: %w", r.version, err)
	}

	var current *semver.Version
	current, err = semver.NewVersion(currentVersion)
	if err != nil {
		return false, fmt.Errorf("failed to parse current version %s: %w", currentVersion, err)
	}

	return latest.GreaterThan(current), nil
}

type releaseAsset struct {
	name        string
	downloadURL string
}

const latestReleaseURL = "https://api.github.com/repos/devusSs/minly/releases/latest"

type latestReleaseResponse struct {
	URL       string `json:"url"`
	AssetsURL string `json:"assets_url"`
	UploadURL string `json:"upload_url"`
	HTMLURL   string `json:"html_url"`
	ID        int    `json:"id"`
	Author    struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		UserViewType      string `json:"user_view_type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"author"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Immutable       bool      `json:"immutable"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []struct {
		URL      string `json:"url"`
		ID       int    `json:"id"`
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		Label    string `json:"label"`
		Uploader struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			UserViewType      string `json:"user_view_type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"uploader"`
		ContentType        string    `json:"content_type"`
		State              string    `json:"state"`
		Size               int       `json:"size"`
		Digest             any       `json:"digest"`
		DownloadCount      int       `json:"download_count"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		BrowserDownloadURL string    `json:"browser_download_url"`
	} `json:"assets"`
	TarballURL string `json:"tarball_url"`
	ZipballURL string `json:"zipball_url"`
	Body       string `json:"body"`
}

func findLatestRelease(ctx context.Context) (*release, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch latest release, status code: %d", resp.StatusCode)
	}

	var res latestReleaseResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	r := &release{
		version:   res.TagName,
		createdAt: res.CreatedAt,
		changelog: res.Body,
		assets:    make([]releaseAsset, 0, len(res.Assets)),
	}

	for _, asset := range res.Assets {
		r.assets = append(r.assets, releaseAsset{
			name:        asset.Name,
			downloadURL: asset.BrowserDownloadURL,
		})
	}

	return r, nil
}

func findMatchingAsset(assets []releaseAsset) (releaseAsset, error) {
	if len(assets) == 0 {
		return releaseAsset{}, errors.New("no assets available")
	}

	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH
	if targetArch == "amd64" {
		targetArch = "x86_64"
	}

	for _, asset := range assets {
		name := strings.ToLower(asset.name)
		if strings.Contains(name, targetOS) && strings.Contains(name, targetArch) {
			return asset, nil
		}
	}

	return releaseAsset{}, fmt.Errorf(
		"no matching asset found for OS %s and architecture %s",
		targetOS,
		targetArch,
	)
}
