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

func TestResolveConfigPathErrorsWhenNotFound(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	_, err := resolveConfigPath("")
	if err == nil {
		t.Fatal("resolveConfigPath() expected error, got nil")
	}
}

func TestResolveConfigPathFindsParent(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "apps", "ui")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(root, config.DefaultFilename)
	if err := os.WriteFile(cfgPath, []byte("name: Test\nservices:\n  a:\n    label: A\n    command: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(sub)
	path, err := resolveConfigPath("")
	if err != nil {
		t.Fatalf("resolveConfigPath() error = %v", err)
	}
	if path != cfgPath {
		t.Fatalf("path = %q, want %q", path, cfgPath)
	}
}

func TestResolveWorkDirFromConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, config.DefaultFilename)
	t.Chdir(dir)

	workDir, err := resolveWorkDir(cfgPath)
	if err != nil {
		t.Fatalf("resolveWorkDir() error = %v", err)
	}
	if workDir != dir {
		t.Fatalf("workDir = %q, want %q", workDir, dir)
	}
}
