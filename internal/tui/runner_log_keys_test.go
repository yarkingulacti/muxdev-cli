package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

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
