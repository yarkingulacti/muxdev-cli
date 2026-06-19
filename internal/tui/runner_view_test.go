package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

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

func TestOpenRerunMenuSelectsAllRunning(t *testing.T) {
	m := runnerModel{serviceIDs: []string{"backend", "ui"}}
	m.openRerunMenu()

	if !m.rerunMenu {
		t.Fatal("expected rerun menu open")
	}
	if !m.rerunSelected["backend"] || !m.rerunSelected["ui"] {
		t.Fatalf("expected all services selected: %v", m.rerunSelected)
	}
}

func TestRerunChosenIDs(t *testing.T) {
	m := runnerModel{
		serviceIDs: []string{"backend", "ui"},
		rerunSelected: map[string]bool{
			"backend": true,
			"ui":      false,
		},
	}

	chosen := m.rerunChosenIDs()
	if len(chosen) != 1 || chosen[0] != "backend" {
		t.Fatalf("chosen = %v, want [backend]", chosen)
	}
}

func TestApplyRerunResolvesDependencies(t *testing.T) {
	m := runnerModel{
		cfg: &config.Config{
			Services: map[string]config.Service{
				"db":      {Label: "DB", Command: "true"},
				"backend": {Label: "Backend", Command: "true", DependsOn: []string{"db"}},
			},
		},
		serviceIDs: []string{"db", "backend"},
		rerunSelected: map[string]bool{
			"backend": true,
			"db":      false,
		},
		rerunMenu: true,
		done:      true,
	}

	next, cmd := m.applyRerun([]string{"backend"})
	if cmd == nil {
		t.Fatal("expected restart command")
	}
	if len(next.serviceIDs) != 2 {
		t.Fatalf("serviceIDs = %v, want db and backend", next.serviceIDs)
	}
	if next.rerunMenu || next.done {
		t.Fatalf("expected rerun menu closed and running state: %+v", next)
	}
}

func TestRefreshLogViewportPreservesHistoryScroll(t *testing.T) {
	m := runnerModel{
		ready:      true,
		followTail: false,
		entries: []logEntry{
			{label: "svc", text: "line-0"},
			{label: "svc", text: "line-1"},
			{label: "svc", text: "line-2"},
			{label: "svc", text: "line-3"},
			{label: "svc", text: "line-4"},
		},
	}
	m.viewport = viewport.New(40, 2)
	m.viewport.KeyMap = runnerLogViewportKeyMap()
	m.refreshLogViewport()

	m.viewport.SetYOffset(1)
	m.appendLog(logMsg{label: "svc", text: "line-5"})
	m.refreshLogViewport()

	if m.viewport.YOffset != 1 {
		t.Fatalf("YOffset = %d, want 1 while browsing history", m.viewport.YOffset)
	}

	m.followTail = true
	m.appendLog(logMsg{label: "svc", text: "line-6"})
	m.refreshLogViewport()
	if !m.viewport.AtBottom() {
		t.Fatalf("expected viewport at bottom when followTail is true")
	}
}

func TestHandleLogScrollLineAndPage(t *testing.T) {
	m := runnerModel{
		ready:      true,
		followTail: true,
		entries: []logEntry{
			{label: "svc", text: "a"},
			{label: "svc", text: "b"},
			{label: "svc", text: "c"},
			{label: "svc", text: "d"},
			{label: "svc", text: "e"},
		},
	}
	m.viewport = viewport.New(20, 2)
	m.viewport.KeyMap = runnerLogViewportKeyMap()
	m.refreshLogViewport()

	bottom := m.viewport.YOffset
	if !m.handleLogScroll(tea.KeyMsg{Type: tea.KeyPgUp}) {
		t.Fatal("pgup should be handled")
	}
	if m.viewport.YOffset != bottom-1 || m.followTail {
		t.Fatalf("after pgup: offset=%d followTail=%v", m.viewport.YOffset, m.followTail)
	}

	start := m.viewport.YOffset
	if !m.handleLogScroll(tea.KeyMsg{Type: tea.KeyCtrlPgUp}) {
		t.Fatal("ctrl+pgup should be handled")
	}
	if m.viewport.YOffset >= start {
		t.Fatalf("ctrl+pgup should scroll further up: start=%d now=%d", start, m.viewport.YOffset)
	}

	m.refreshLogViewport()
	if !m.handleLogScroll(tea.KeyMsg{Type: tea.KeyCtrlU}) {
		t.Fatal("ctrl+u should page up")
	}
	if m.viewport.YOffset != 0 {
		t.Fatalf("ctrl+u should reach top, offset=%d", m.viewport.YOffset)
	}
}
