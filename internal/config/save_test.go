package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
)

func TestNormalizeServiceID(t *testing.T) {
	got, err := config.NormalizeServiceID("Web UI")
	if err != nil {
		t.Fatal(err)
	}
	if got != "web_ui" {
		t.Fatalf("got %q, want web_ui", got)
	}

	if _, err := config.NormalizeServiceID("9bad"); err == nil {
		t.Fatal("expected error for id starting with digit")
	}
}

func TestSaveRoundTrip(t *testing.T) {
	cfg := &config.Config{
		Name:     "Test",
		Subtitle: "Stack",
		Services: map[string]config.Service{
			"api": {
				Label:     "API",
				Command:   "echo hi",
				DependsOn: []string{},
			},
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "muxdev.yaml")
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Name != cfg.Name {
		t.Fatalf("name = %q", loaded.Name)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
