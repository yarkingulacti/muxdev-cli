package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/yarkingulacti/muxdev-cli/internal/version"
)

// Manifest describes a release served from a static artifact store (Nexus, S3, etc.).
type Manifest struct {
	Version   string            `json:"version"`
	Tag       string            `json:"tag"`
	BaseURL   string            `json:"base_url"`
	Checksums string            `json:"checksums"`
	Assets    map[string]string `json:"assets"`
}

func manifestURLFromEnv() string {
	return strings.TrimSpace(os.Getenv("MUXDEV_UPDATE_URL"))
}

func resolveManifestURL(baseURL, target string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return ""
	}
	if strings.TrimSpace(target) == "" {
		return baseURL
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	tag := strings.TrimSpace(target)
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}

	dir := path.Dir(u.Path)
	u.Path = path.Join(dir, tag, "manifest.json")
	return u.String()
}

func fetchManifest(ctx context.Context, manifestURL string) (Manifest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		return Manifest{}, err
	}
	req.Header.Set("Accept", "application/json")
	setRequestHeaders(req, version.Short())

	resp, err := httpClient.Do(req)
	if err != nil {
		return Manifest{}, fmt.Errorf("fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return Manifest{}, fmt.Errorf("fetch manifest %q: %w", manifestURL, httpStatusError(resp.StatusCode, string(body)))
	}

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	rewriteManifestBaseURL(manifestURL, &manifest)
	if err := manifest.validate(); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

// rewriteManifestBaseURL aligns base_url with the URL used to fetch the manifest.
// Uploads may record an internal NEXUS_URL while clients fetch via a public proxy.
func rewriteManifestBaseURL(manifestURL string, m *Manifest) {
	manifestURL = strings.TrimSpace(manifestURL)
	if manifestURL == "" {
		return
	}
	mu, err := url.Parse(manifestURL)
	if err != nil || mu.Scheme == "" || mu.Host == "" {
		return
	}
	publishDir := path.Dir(mu.Path)
	if path.Base(mu.Path) == "manifest.json" {
		mu.Path = publishDir
	} else {
		tag := m.tagName()
		if tag == "" {
			return
		}
		mu.Path = path.Join(publishDir, tag)
	}
	mu.RawQuery = ""
	mu.Fragment = ""
	base := strings.TrimSuffix(mu.String(), "/")
	m.BaseURL = base
	m.Checksums = base + "/checksums.txt"
}

func (m Manifest) validate() error {
	if strings.TrimSpace(m.Tag) == "" && strings.TrimSpace(m.Version) == "" {
		return fmt.Errorf("manifest missing version/tag")
	}
	if strings.TrimSpace(m.BaseURL) == "" {
		return fmt.Errorf("manifest missing base_url")
	}
	if len(m.Assets) == 0 {
		return fmt.Errorf("manifest missing assets")
	}
	return nil
}

func (m Manifest) tagName() string {
	tag := strings.TrimSpace(m.Tag)
	if tag != "" {
		if !strings.HasPrefix(tag, "v") {
			return "v" + tag
		}
		return tag
	}
	v := strings.TrimPrefix(strings.TrimSpace(m.Version), "v")
	return "v" + v
}

func (m Manifest) toRelease() Release {
	seen := make(map[string]struct{})
	var assets []Asset

	addAsset := func(name, downloadURL string) {
		if name == "" || downloadURL == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		assets = append(assets, Asset{
			Name:               name,
			BrowserDownloadURL: downloadURL,
		})
	}

	for _, name := range m.Assets {
		addAsset(name, joinURL(m.BaseURL, name))
	}

	checksumsURL := strings.TrimSpace(m.Checksums)
	if checksumsURL == "" {
		checksumsURL = joinURL(m.BaseURL, "checksums.txt")
	}
	addAsset("checksums.txt", checksumsURL)

	return Release{
		TagName: m.tagName(),
		Assets:  assets,
	}
}

func joinURL(base, name string) string {
	base = strings.TrimRight(base, "/")
	return base + "/" + strings.TrimPrefix(name, "/")
}
