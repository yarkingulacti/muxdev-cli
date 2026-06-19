package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
)

func TestRunList(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, config.DefaultFilename)
	if err := os.WriteFile(cfgPath, []byte(`name: Test
subtitle: Local stack
services:
  backend:
    label: Backend
    command: true
    port: "4000"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	if err := runList(""); err != nil {
		t.Fatalf("runList() error = %v", err)
	}
}

func TestRunListMissingConfig(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if err := runList(""); err == nil {
		t.Fatal("runList() expected error, got nil")
	}
}
