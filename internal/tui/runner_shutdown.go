package tui

import tea "github.com/charmbracelet/bubbletea"

func (m runnerModel) canWaitForRunnerShutdown() bool {
	return m.cancel != nil || m.attached || m.attachPending || m.killPending
}

func (m *runnerModel) quitNow() (runnerModel, tea.Cmd) {
	if m.session != nil {
		_ = m.session.Finish(nil)
		m.session = nil
	}
	return *m, tea.Quit
}

func (m *runnerModel) requestShutdown(forceful bool) (runnerModel, tea.Cmd) {
	if m.shuttingDown && !forceful {
		forceful = true
	}

	m.filterMenu = false

	if m.shutdown != nil {
		m.shutdown.Forceful = forceful
	}

	if m.attachCancel != nil {
		m.attachCancel()
	}

	if m.cancel != nil {
		if !forceful {
			m.shuttingDown = true
		}
		m.cancel()
	} else if !forceful {
		m.shuttingDown = true
	}

	if m.awaitingKill && !m.canWaitForRunnerShutdown() {
		m.done = true
	}

	if forceful {
		return m.quitNow()
	}
	if !m.canWaitForRunnerShutdown() {
		return m.quitNow()
	}

	m.shuttingDown = true
	return *m, nil
}
