package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
)

func TestResolveConfigPathExplicit(t *testing.T) {
	path, err := resolveConfigPath("./custom.yaml")
	if err != nil {
		t.Fatalf("resolveConfigPath() error = %v", err)
	}
	if path != "./custom.yaml" {
		t.Fatalf("path = %q, want ./custom.yaml", path)
	}
}

func TestResolveConfigPathFindsExisting(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, config.DefaultFilename)
	if err := os.WriteFile(cfgPath, []byte("name: Test\nservices:\n  a:\n    label: A\n    command: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(dir)
	path, err := resolveConfigPath("")
	if err != nil {
		t.Fatalf("resolveConfigPath() error = %v", err)
	}
	if path != cfgPath {
		t.Fatalf("path = %q, want %q", path, cfgPath)
	}
}

func TestResolveConfigPathDefaultsToCWD(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	path, err := resolveConfigPath("")
	if err != nil {
		t.Fatalf("resolveConfigPath() error = %v", err)
	}
	want := filepath.Join(dir, config.DefaultFilename)
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}
