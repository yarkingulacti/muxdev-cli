package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/yarkingulacti/muxdev-cli/internal/version"
)

type Release struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Prerelease  bool   `json:"prerelease"`
	Draft       bool   `json:"draft"`
	HTMLURL     string `json:"html_url"`
	Assets      []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Result struct {
	Current         string
	Latest          string
	UpdateAvailable bool
	Release         Release
	InstallMethod   InstallMethod
	ManifestURL     string
}

type CheckOptions struct {
	Channel     Channel
	Target      string
	ManifestURL string
}

func Check(ctx context.Context, opts CheckOptions) (Result, error) {
	exe, err := CurrentExecutable()
	if err != nil {
		return Result{}, err
	}

	method := Detect(exe, version.InstallMethod)
	current := version.Version
	if current == "dev" {
		current = "v0.0.0-dev"
	} else if !strings.HasPrefix(current, "v") {
		current = "v" + current
	}

	release, err := fetchRelease(ctx, opts)
	if err != nil {
		return Result{}, err
	}

	latest := release.TagName
	if !strings.HasPrefix(latest, "v") {
		latest = "v" + latest
	}

	manifestURL := ""
	base := strings.TrimSpace(opts.ManifestURL)
	if base == "" {
		base = manifestURLFromEnv()
	}
	if base != "" {
		manifestURL = resolveManifestURL(base, opts.Target)
	}

	return Result{
		Current:         current,
		Latest:          latest,
		UpdateAvailable: semver.Compare(latest, current) > 0,
		Release:         release,
		InstallMethod:   method,
		ManifestURL:     manifestURL,
	}, nil
}

func fetchRelease(ctx context.Context, opts CheckOptions) (Release, error) {
	manifestBase := strings.TrimSpace(opts.ManifestURL)
	if manifestBase == "" {
		manifestBase = manifestURLFromEnv()
	}
	if manifestBase != "" {
		if opts.Channel == ChannelPrerelease {
			return Release{}, fmt.Errorf("prerelease channel is not supported with manifest updates")
		}
		manifestURL := resolveManifestURL(manifestBase, opts.Target)
		manifest, err := fetchManifest(ctx, manifestURL)
		if err != nil {
			return Release{}, err
		}
		release := manifest.toRelease()
		return release, nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", RepoOwner, RepoName, strings.TrimPrefix(opts.Target, "v"))
	if opts.Target == "" {
		if opts.Channel == ChannelPrerelease {
			url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", RepoOwner, RepoName)
			return fetchLatestFromList(ctx, url, false)
		}
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", RepoOwner, RepoName)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	setRequestHeaders(req, version.Short())

	resp, err := httpClient.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return Release{}, fmt.Errorf("github api: %w", httpStatusError(resp.StatusCode, string(body)))
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return Release{}, fmt.Errorf("decode release: %w", err)
	}
	return release, nil
}

func fetchLatestFromList(ctx context.Context, url string, stableOnly bool) (Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	setRequestHeaders(req, version.Short())

	resp, err := httpClient.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return Release{}, err
	}

	for _, release := range releases {
		if release.Draft {
			continue
		}
		if stableOnly && release.Prerelease {
			continue
		}
		return release, nil
	}
	return Release{}, fmt.Errorf("no matching release found")
}

func FindAsset(release Release, assetName string) (Asset, error) {
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			return asset, nil
		}
	}
	return Asset{}, fmt.Errorf("asset %q not found in release %s", assetName, release.TagName)
}
