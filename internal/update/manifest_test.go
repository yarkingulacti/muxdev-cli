package update_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/update"
)

func TestCheckManifestFromEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if path.Base(r.URL.Path) != "latest.json" {
			http.NotFound(w, r)
			return
		}
		base := "http://" + r.Host + "/stable/v1.0.0"
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "version": "1.0.0",
  "tag": "v1.0.0",
  "base_url": "` + base + `",
  "assets": {"linux_amd64": "muxdev_1.0.0_linux_amd64.tar.gz"}
}`))
	}))
	defer server.Close()

	t.Setenv("MUXDEV_UPDATE_URL", server.URL+"/stable/latest.json")

	result, err := update.Check(context.Background(), update.CheckOptions{})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.ManifestURL != server.URL+"/stable/latest.json" {
		t.Fatalf("ManifestURL = %q", result.ManifestURL)
	}
	if result.Latest != "v1.0.0" {
		t.Fatalf("Latest = %q, want v1.0.0", result.Latest)
	}
}

func TestCheckManifestPinnedVersion(t *testing.T) {
	const pinned = `{
  "version": "0.9.0",
  "tag": "v0.9.0",
  "base_url": "BASE/stable/v0.9.0",
  "assets": {"linux_amd64": "muxdev_0.9.0_linux_amd64.tar.gz"}
}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch path.Base(r.URL.Path) {
		case "manifest.json":
			body := strings.ReplaceAll(pinned, "BASE", "http://"+r.Host+"/repository/muxdev-releases")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(body))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	result, err := update.Check(context.Background(), update.CheckOptions{
		ManifestURL: server.URL + "/repository/muxdev-releases/stable/latest.json",
		Target:      "v0.9.0",
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.Latest != "v0.9.0" {
		t.Fatalf("Latest = %q, want v0.9.0", result.Latest)
	}
	want := server.URL + "/repository/muxdev-releases/stable/v0.9.0/manifest.json"
	if result.ManifestURL != want {
		t.Fatalf("ManifestURL = %q, want %q", result.ManifestURL, want)
	}
}

func TestManifestHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("<!DOCTYPE html><html><title>404</title></html>"))
	}))
	defer server.Close()

	_, err := update.Check(context.Background(), update.CheckOptions{
		ManifestURL: server.URL + "/stable/latest.json",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unexpected HTML response") {
		t.Fatalf("error = %q, want HTML summary", err)
	}
}
