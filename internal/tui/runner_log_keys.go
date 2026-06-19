package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type runnerLogKeyMap struct {
	LineUp     key.Binding
	LineDown   key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
}

var runnerLogKeys = runnerLogKeyMap{
	LineUp: key.NewBinding(
		key.WithKeys("pgup"),
	),
	LineDown: key.NewBinding(
		key.WithKeys("pgdown"),
	),
	// ctrl+u/d work in terminals (e.g. Warp) that swallow ctrl+pgup/pgdown.
	PageUp: key.NewBinding(
		key.WithKeys("ctrl+pgup", "ctrl+u", "u"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("ctrl+pgdown", "ctrl+d", "d", "space"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("up", "k"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("down", "j"),
	),
}

func runnerLogViewportKeyMap() viewport.KeyMap {
	disabled := key.NewBinding(key.WithDisabled())
	return viewport.KeyMap{
		PageUp:       disabled,
		PageDown:     disabled,
		HalfPageUp:   disabled,
		HalfPageDown: disabled,
		Up:           disabled,
		Down:         disabled,
		Left:         disabled,
		Right:        disabled,
	}
}

func logScrollPageUp(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyCtrlPgUp || msg.Type == tea.KeyCtrlU || key.Matches(msg, runnerLogKeys.PageUp)
}

func logScrollPageDown(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyCtrlPgDown || msg.Type == tea.KeyCtrlD || key.Matches(msg, runnerLogKeys.PageDown)
}

func logScrollLineUp(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyPgUp || key.Matches(msg, runnerLogKeys.LineUp)
}

func logScrollLineDown(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyPgDown || key.Matches(msg, runnerLogKeys.LineDown)
}

func logScrollKey(msg tea.KeyMsg) bool {
	return logScrollPageUp(msg) || logScrollPageDown(msg) ||
		logScrollLineUp(msg) || logScrollLineDown(msg) ||
		runnerLogArrowScroll(msg)
}

func handleLogScrollViewport(vp *viewport.Model, followTail *bool, msg tea.KeyMsg) bool {
	if vp == nil {
		return false
	}
	switch {
	case logScrollPageUp(msg):
		scrollViewport(vp, followTail, vp.Height, -1)
		return true
	case logScrollPageDown(msg):
		scrollViewport(vp, followTail, vp.Height, 1)
		return true
	case logScrollLineUp(msg):
		scrollViewport(vp, followTail, 1, -1)
		return true
	case logScrollLineDown(msg):
		scrollViewport(vp, followTail, 1, 1)
		return true
	case runnerLogArrowScroll(msg):
		if msg.Type == tea.KeyUp || key.Matches(msg, runnerLogKeys.ScrollUp) {
			scrollViewport(vp, followTail, 1, -1)
		} else {
			scrollViewport(vp, followTail, 1, 1)
		}
		return true
	}
	return false
}

func scrollViewport(vp *viewport.Model, followTail *bool, lines int, direction int) {
	if lines <= 0 {
		return
	}
	if direction < 0 {
		vp.ScrollUp(lines)
	} else {
		vp.ScrollDown(lines)
	}
	if followTail != nil {
		*followTail = vp.AtBottom()
	}
}

func syncViewportFollowTail(vp viewport.Model, followTail *bool) {
	if followTail != nil && vp.AtBottom() {
		*followTail = true
	}
}

func (m *runnerModel) handleLogScroll(msg tea.KeyMsg) bool {
	if !m.ready || m.awaitingKill {
		return false
	}
	return handleLogScrollViewport(&m.viewport, &m.followTail, msg)
}

func runnerLogArrowScroll(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyUp || msg.Type == tea.KeyDown ||
		key.Matches(msg, runnerLogKeys.ScrollUp) || key.Matches(msg, runnerLogKeys.ScrollDown)
}

const logScrollHelp = "pgup/pgdn line  ctrl+u/d page  f filter  r re-run  q quit"
const logScrollHelpHistory = "history  pgdn to live  pgup/pgdn line  ctrl+u/d page  f filter  r re-run  q quit"
const logScrollHelpAttached = "pgup/pgdn line  ctrl+u/d page  q quit"
