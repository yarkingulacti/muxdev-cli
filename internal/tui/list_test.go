package tui

import (
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

	out := RenderServiceList(cfg, 80)
	for _, want := range []string{"My App", "Local development stack", "2 services", "ID", "DEPENDS ON", "backend", "Backend", "ui", "Web UI", "backend"} {
		if !strings.Contains(out, want) {
			t.Fatalf("RenderServiceList() missing %q\noutput:\n%s", want, out)
		}
	}
}

func TestRenderServiceListSingleService(t *testing.T) {
	cfg := &config.Config{
		Name: "Solo",
		Services: map[string]config.Service{
			"api": {Label: "API", Command: "true"},
		},
	}

	out := RenderServiceList(cfg, 80)
	if !strings.Contains(out, "1 service") {
		t.Fatalf("RenderServiceList() = %q, want singular service count", out)
	}
}
