package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type configureKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Space key.Binding
	Back  key.Binding
	Quit  key.Binding
}

var configureKeys = configureKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k", "ctrl+p"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j", "ctrl+n"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
	),
}

func configureKeyUp(msg tea.KeyMsg) bool {
	return key.Matches(msg, configureKeys.Up)
}

func configureKeyDown(msg tea.KeyMsg) bool {
	return key.Matches(msg, configureKeys.Down)
}

func configureKeyEnter(msg tea.KeyMsg) bool {
	return key.Matches(msg, configureKeys.Enter)
}

func configureKeySpace(msg tea.KeyMsg) bool {
	return key.Matches(msg, configureKeys.Space)
}

func configureKeyBack(msg tea.KeyMsg) bool {
	return key.Matches(msg, configureKeys.Back)
}

func configureKeyQuit(msg tea.KeyMsg) bool {
	return key.Matches(msg, configureKeys.Quit)
}

func configureInputPhase(phase configurePhase) bool {
	switch phase {
	case phaseCfgName, phaseCfgSubtitle, phaseCfgServiceID,
		phaseCfgServiceLabel, phaseCfgServiceCommand, phaseCfgServicePort:
		return true
	default:
		return false
	}
}

func (m *configureModel) enterMenuPhase(phase configurePhase) {
	m.input.Blur()
	m.phase = phase
}

func (m *configureModel) enterInputPhase(phase configurePhase) tea.Cmd {
	m.phase = phase
	m.input.Focus()
	return textinput.Blink
}
