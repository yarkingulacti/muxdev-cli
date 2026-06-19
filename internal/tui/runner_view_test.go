package tui

import (
	"strings"
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
)

func TestLogContentFiltersByService(t *testing.T) {
	m := runnerModel{
		cfg: &config.Config{
			Services: map[string]config.Service{
				"backend": {Label: "Backend", Command: "true"},
				"ui":      {Label: "Web UI", Command: "true"},
			},
		},
		serviceIDs: []string{"backend", "ui"},
		entries: []logEntry{
			{label: "Backend", text: "started"},
			{label: "Web UI", text: "ready"},
			{label: "Backend", text: "listening"},
		},
	}

	all := m.logContent()
	if !strings.Contains(all, "started") || !strings.Contains(all, "ready") {
		t.Fatalf("unfiltered content missing lines: %q", all)
	}

	m.filterLabel = "Backend"
	filtered := m.logContent()
	if strings.Contains(filtered, "ready") {
		t.Fatalf("filtered content should not include ui logs: %q", filtered)
	}
	if !strings.Contains(filtered, "started") || !strings.Contains(filtered, "listening") {
		t.Fatalf("filtered content missing backend logs: %q", filtered)
	}
}

func TestFilterMenuIndex(t *testing.T) {
	m := runnerModel{
		cfg: &config.Config{
			Services: map[string]config.Service{
				"backend": {Label: "Backend", Command: "true"},
				"ui":      {Label: "Web UI", Command: "true"},
			},
		},
		serviceIDs:  []string{"backend", "ui"},
		filterLabel: "Web UI",
	}

	if got := m.filterMenuIndex(); got != 2 {
		t.Fatalf("filterMenuIndex() = %d, want 2", got)
	}
}
