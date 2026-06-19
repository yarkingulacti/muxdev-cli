package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/yarkingulacti/muxdev-cli/internal/runner"
)

func TestRunnerGracefulQuitKeyDetectsCtrlQ(t *testing.T) {
	if !runnerGracefulQuitKey(tea.KeyMsg{Type: tea.KeyCtrlQ}) {
		t.Fatal("KeyCtrlQ should count as graceful quit")
	}
	if !runnerGracefulQuitKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{17}}) {
		t.Fatal("ctrl+q rune should count as graceful quit")
	}
	if runnerGracefulQuitKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}) {
		t.Fatal("plain q should not count as graceful quit")
	}
}

func TestRunnerUpdateCtrlQRequestsShutdown(t *testing.T) {
	shutdown := &runner.ShutdownRequest{}
	doneCh := make(chan runDoneMsg, 1)
	m := runnerModel{
		ready:    true,
		shutdown: shutdown,
		doneCh:   doneCh,
		cancel:   func() {},
	}
	m.viewport = viewport.New(80, 24)
	m.viewport.KeyMap = runnerLogViewportKeyMap()

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})
	nm := next.(runnerModel)
	if !nm.shuttingDown {
		t.Fatal("expected shuttingDown after ctrl+q")
	}
	if cmd != nil {
		t.Fatal("graceful shutdown should wait for runner completion")
	}
}

func TestLogScrollPageUpDetectsCtrlU(t *testing.T) {
	if !logScrollPageUp(tea.KeyMsg{Type: tea.KeyCtrlU}) {
		t.Fatal("ctrl+u should count as page up")
	}
	if logScrollPageUp(tea.KeyMsg{Type: tea.KeyPgUp}) {
		t.Fatal("pgup alone should be line scroll, not page")
	}
}

func TestLogScrollKeyBlocksViewport(t *testing.T) {
	if !logScrollKey(tea.KeyMsg{Type: tea.KeyCtrlD}) {
		t.Fatal("ctrl+d should be a muxdev scroll key")
	}
	if logScrollKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")}) {
		t.Fatal("f should not be treated as scroll")
	}
}
