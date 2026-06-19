package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
)

func TestBindPortForServiceUsesTemplateEnvKeys(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".env", "XTTS_PORT=5005\n")
	cfg := &config.Config{
		Services: map[string]config.Service{
			"backend": {
				Command: "bash run-dev.sh",
				Port:    "${BACKEND_PORT}",
			},
		},
	}

	got := config.BindPortForService(dir, cfg.Services["backend"])
	if got != 5005 {
		t.Fatalf("BindPortForService() = %d, want 5005 via XTTS_PORT fallback", got)
	}
}

func TestBindPortForServiceExplicitPort(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.Service{
			"api": {Port: "8080"},
		},
	}
	if got := config.BindPortForService("", cfg.Services["api"]); got != 8080 {
		t.Fatalf("BindPortForService() = %d, want 8080", got)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
