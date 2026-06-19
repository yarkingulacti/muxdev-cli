package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
)

func TestConfigureRootMenuArrowKeys(t *testing.T) {
	m := newConfigureModel(ConfigureOptions{})
	m.enterMenuPhase(phaseCfgRootMenu)
	m.width = 80
	m.rootMenuCursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 := updated.(*configureModel)
	if m2.rootMenuCursor != 1 {
		t.Fatalf("KeyDown: cursor = %d, want 1", m2.rootMenuCursor)
	}

	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeyUp})
	m3 := updated.(*configureModel)
	if m3.rootMenuCursor != 0 {
		t.Fatalf("KeyUp: cursor = %d, want 0", m3.rootMenuCursor)
	}
}

func TestConfigureServiceMenuArrowKeys(t *testing.T) {
	m := newConfigureModel(ConfigureOptions{Edit: true})
	m.enterMenuPhase(phaseCfgServiceMenu)
	m.width = 80
	m.services = map[string]config.Service{
		"a": {Label: "A", Command: "true"},
		"b": {Label: "B", Command: "true"},
		"c": {Label: "C", Command: "true"},
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 := updated.(*configureModel)
	if m2.serviceMenuCursor != 1 {
		t.Fatalf("KeyDown: cursor = %d, want 1", m2.serviceMenuCursor)
	}
}

func TestConfigureServiceMenuAfterOpenEdit(t *testing.T) {
	m := newConfigureModel(ConfigureOptions{Edit: true})
	m.enterMenuPhase(phaseCfgServiceMenu)
	m.width = 80
	m.serviceMenuCursor = 0
	m.services = map[string]config.Service{
		"a": {Label: "A", Command: "true"},
		"b": {Label: "B", Command: "true"},
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := updated.(*configureModel)
	if m2.phase != phaseCfgServiceEditMenu {
		t.Fatalf("phase = %v, want service edit menu", m2.phase)
	}

	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeyDown})
	m3 := updated.(*configureModel)
	if m3.serviceEditCursor != 1 {
		t.Fatalf("edit menu cursor = %d, want 1", m3.serviceEditCursor)
	}
}

func TestConfigureMenuPhaseBlursInput(t *testing.T) {
	m := newConfigureModel(ConfigureOptions{Edit: true})
	m.input.Focus()
	m.enterMenuPhase(phaseCfgRootMenu)
	if m.input.Focused() {
		t.Fatal("input should be blurred in menu phase")
	}
}
