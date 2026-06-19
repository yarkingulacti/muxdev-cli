package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
)

func TestRenderServiceList(t *testing.T) {
	cfg := &config.Config{
		Name:     "My App",
		Subtitle: "Local development stack",
		Services: map[string]config.Service{
			"backend": {
				Label:   "Backend",
				Command: "go run ./cmd/api",
				Port:    "4000",
			},
			"ui": {
				Label:     "Web UI",
				Command:   "npm run dev",
				Port:      "3000",
				DependsOn: []string{"backend"},
			},
		},
	}

	out := RenderServiceList(cfg, "", 80)
	for _, want := range []string{"My App", "Local development stack", "2 services", "ID", "DEPENDS ON", "backend", "Backend", "ui", "Web UI", "backend"} {
		if !strings.Contains(out, want) {
			t.Fatalf("RenderServiceList() missing %q\noutput:\n%s", want, out)
		}
	}
}

func TestRenderServiceListExpandsPortFromEnv(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("UI_PORT=4000\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Name: "App",
		Services: map[string]config.Service{
			"ui": {Label: "Web UI", Command: "npm run dev", Port: "${UI_PORT}"},
		},
	}

	out := RenderServiceList(cfg, dir, 80)
	if strings.Contains(out, "${UI_PORT}") {
		t.Fatalf("RenderServiceList() should expand port, got:\n%s", out)
	}
	if !strings.Contains(out, "4000") {
		t.Fatalf("RenderServiceList() missing expanded port 4000:\n%s", out)
	}
	if !strings.Contains(out, ".env") {
		t.Fatalf("RenderServiceList() missing port source .env:\n%s", out)
	}
}

func TestRenderServiceListSingleService(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.Service{
			"api": {Label: "API", Command: "true"},
		},
	}

	out := RenderServiceList(cfg, "", 80)
	if !strings.Contains(out, "1 service") {
		t.Fatalf("RenderServiceList() = %q, want singular service count", out)
	}
}
