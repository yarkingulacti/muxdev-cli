package tui

import (
	"strings"
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
	"github.com/yarkingulacti/muxdev-cli/internal/version"
)

func TestRenderRuntimeHeaderIncludesVersion(t *testing.T) {
	cfg := &config.Config{Name: "My App", Subtitle: "Local stack"}
	out := renderRuntimeHeader(cfg, 80, "Running: api")

	for _, want := range []string{"My App", "Local stack", "Running: api", "muxdev", version.Short()} {
		if !strings.Contains(out, want) {
			t.Fatalf("renderRuntimeHeader() missing %q:\n%s", want, out)
		}
	}
}

func TestRenderHeaderOmitsVersion(t *testing.T) {
	cfg := &config.Config{Name: "My App"}
	out := renderHeader(cfg, 80, "2 services")

	if !strings.Contains(out, "2 services") {
		t.Fatalf("renderHeader() missing status:\n%s", out)
	}
	if strings.Contains(out, "muxdev "+version.Short()) {
		t.Fatalf("renderHeader() should not include muxdev version:\n%s", out)
	}
}

func TestAppendRuntimeStatusMeta(t *testing.T) {
	got := appendRuntimeStatusMeta("")
	if got != "muxdev "+version.Short() {
		t.Fatalf("appendRuntimeStatusMeta(\"\") = %q", got)
	}
	got = appendRuntimeStatusMeta("Select services")
	if !strings.Contains(got, "Select services") || !strings.Contains(got, "muxdev") {
		t.Fatalf("appendRuntimeStatusMeta() = %q", got)
	}
}
