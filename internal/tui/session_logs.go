package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yarkingulacti/muxdev-cli/internal/logs"
)

type SessionLogsOptions struct {
	WorkDir string
	All     bool
}

type sessionLogsPhase int

const (
	sessionLogsPhaseList sessionLogsPhase = iota
	sessionLogsPhaseView
)

type sessionLogsModel struct {
	workDir  string
	all      bool
	sessions []logs.Session
	phase    sessionLogsPhase
	cursor   int
	width    int
	height   int
	ready    bool

	viewport   viewport.Model
	content    string
	followTail bool
}

func RunSessionLogs(opts SessionLogsOptions) error {
	sessions, err := logs.ListSessions(opts.WorkDir)
	if err != nil {
		return err
	}
	model := sessionLogsModel{
		workDir:  opts.WorkDir,
		all:      opts.All,
		sessions: sessions,
	}
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func (m sessionLogsModel) Init() tea.Cmd {
	return nil
}

func (m sessionLogsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.phase == sessionLogsPhaseView {
			switch msg.String() {
			case "ctrl+c", "q", "esc", "b":
				m.phase = sessionLogsPhaseList
				m.followTail = true
				return m, nil
			}
			if m.ready && handleLogScrollViewport(&m.viewport, &m.followTail, msg) {
				return m, nil
			}
			if m.ready && !logScrollKey(msg) {
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.sessions)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.sessions) == 0 {
				return m, nil
			}
			return m.openSession(m.cursor)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 5
		footerHeight := 1
		viewHeight := msg.Height - headerHeight - footerHeight
		if viewHeight < 1 {
			viewHeight = 1
		}
		if !m.ready {
			m.viewport = viewport.New(msg.Width, viewHeight)
			m.viewport.KeyMap = runnerLogViewportKeyMap()
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewHeight
		}
		if m.phase == sessionLogsPhaseView {
			m.refreshViewport()
		}
	}
	return m, nil
}

func (m sessionLogsModel) openSession(index int) (sessionLogsModel, tea.Cmd) {
	if index < 0 || index >= len(m.sessions) {
		return m, nil
	}
	content, err := logs.ReadLog(m.sessions[index].Dir)
	if err != nil {
		return m, nil
	}
	m.phase = sessionLogsPhaseView
	m.content = content
	m.followTail = true
	m.refreshViewport()
	return m, nil
}

func (m *sessionLogsModel) refreshViewport() {
	if !m.ready {
		return
	}
	offset := m.viewport.YOffset
	if strings.TrimSpace(m.content) == "" {
		m.viewport.SetContent(mutedStyle.Render("Session log is empty."))
	} else {
		m.viewport.SetContent(m.content)
	}
	if m.followTail {
		m.viewport.GotoBottom()
	} else {
		m.viewport.SetYOffset(offset)
	}
}

func (m sessionLogsModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.phase == sessionLogsPhaseView && m.cursor >= 0 && m.cursor < len(m.sessions) {
		session := m.sessions[m.cursor]
		status := sessionStatus(session.Meta)
		header := renderSessionLogsHeader(m.width, fmt.Sprintf("Session %s · %s", session.Meta.ID, status))
		body := m.viewport.View()
		footer := helpStyle.Render("pgup/pgdn line  ctrl+u/d page  esc back  q quit")
		if pag := formatLogPagination(viewportPagination(m.viewport), m.followTail); pag != "" {
			footer = helpStyle.Render(pag + "  |  pgup/pgdn line  ctrl+u/d page  esc back  q quit")
		}
		return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	}

	header := renderSessionLogsHeader(m.width, m.listStatus())
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")

	if len(m.sessions) == 0 {
		b.WriteString(mutedStyle.Render("No saved runtime sessions yet."))
		b.WriteString("\n\n")
		b.WriteString(mutedStyle.Render("Run muxdev to create a session log."))
	} else {
		for i, session := range m.sessions {
			marker := "  "
			if i == m.cursor {
				marker = cursorStyle.Render("> ")
			}
			started := session.Meta.StartedAt.Local().Format("2006-01-02 15:04:05")
			line := fmt.Sprintf("%s%s  %s  %s  %s",
				marker,
				session.Meta.ID,
				started,
				session.Meta.Runtime,
				strings.Join(session.Meta.ServiceIDs, ", "),
			)
			if i == m.cursor {
				line = selectedStyle.Render(line)
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ select  enter view  q quit"))
	return b.String()
}

func (m sessionLogsModel) listStatus() string {
	if m.all {
		return "Saved runtime sessions (all projects)"
	}
	if m.workDir != "" {
		return "Saved runtime sessions · " + m.workDir
	}
	return "Saved runtime sessions"
}

func renderSessionLogsHeader(width int, status string) string {
	title := titleStyle.Render("muxdev logs")
	statusLine := mutedStyle.Render(status)
	content := fmt.Sprintf("%s\n%s", title, statusLine)
	return cardStyle.Width(min(width-2, 72)).Render(content)
}

func sessionStatus(meta logs.Meta) string {
	if meta.EndedAt == nil {
		return "running"
	}
	if meta.ExitError != "" {
		return "error"
	}
	ended := meta.EndedAt.Local().Format("15:04:05")
	return "ended " + ended
}
