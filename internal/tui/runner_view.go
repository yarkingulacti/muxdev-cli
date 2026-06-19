package tui

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
	"github.com/yarkingulacti/muxdev-cli/internal/portkill"
	"github.com/yarkingulacti/muxdev-cli/internal/runner"
)

const maxLogLines = 5000

type logMsg struct {
	label  string
	stderr bool
	text   string
}

type runDoneMsg struct {
	err error
}

type runnerStartedMsg struct {
	cancel context.CancelFunc
	logCh  chan logMsg
	doneCh chan runDoneMsg
}

type portKillMsg struct {
	port   int
	killed int
	err    error
}

type runnerModel struct {
	cfg        *config.Config
	serviceIDs []string
	workDir    string
	viewport   viewport.Model
	lines      []string
	width      int
	height     int
	ready      bool
	cancel     context.CancelFunc
	logCh      chan logMsg
	doneCh     chan runDoneMsg
	runErr     error
	done       bool
	mu         sync.Mutex
	updateHint string

	portConflict *portkill.Conflict
	conflictNote string
	awaitingKill bool
	killPending  bool
}

func runLogs(cfg *config.Config, serviceIDs []string, workDir, updateHint string) error {
	model := newRunnerModel(cfg, serviceIDs, workDir, updateHint)
	p := tea.NewProgram(model, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return err
	}
	m := final.(runnerModel)
	return m.runErr
}

func newRunnerModel(cfg *config.Config, serviceIDs []string, workDir, updateHint string) runnerModel {
	return runnerModel{
		cfg:        cfg,
		serviceIDs: serviceIDs,
		workDir:    workDir,
		lines:      make([]string, 0, 256),
		updateHint: updateHint,
	}
}

func (m runnerModel) Init() tea.Cmd {
	return m.startRunner()
}

func (m runnerModel) startRunner() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		logCh := make(chan logMsg, 128)
		doneCh := make(chan runDoneMsg, 1)

		go func() {
			r := runner.New(m.cfg, m.serviceIDs)
			err := r.Run(runner.Context{
				WorkDir: m.workDir,
				OnLine: func(label string, stderr bool, text string) {
					select {
					case logCh <- logMsg{label: label, stderr: stderr, text: text}:
					case <-ctx.Done():
					}
				},
				CancelFunc: cancel,
			})
			close(logCh)
			doneCh <- runDoneMsg{err: err}
		}()

		return runnerStartedMsg{
			cancel: cancel,
			logCh:  logCh,
			doneCh: doneCh,
		}
	}
}

func (m runnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case runnerStartedMsg:
		m.cancel = msg.cancel
		m.logCh = msg.logCh
		m.doneCh = msg.doneCh
		m.done = false
		m.conflictNote = ""
		if m.ready {
			m.viewport.SetContent(m.logContent())
			m.viewport.GotoTop()
		}
		return m, tea.Batch(waitForLog(m.logCh), waitForDone(m.doneCh))
	case logMsg:
		m.appendLog(msg)
		m.detectPortConflict(msg)
		if m.ready {
			m.viewport.SetContent(m.logContent())
			m.viewport.GotoBottom()
		}
		return m, waitForLog(m.logCh)
	case portKillMsg:
		m.killPending = false
		if msg.err != nil {
			m.conflictNote = errStyle.Render(fmt.Sprintf("Could not free port %d: %v", msg.port, msg.err))
		} else {
			m.clearLogs()
			m.portConflict = nil
			m.awaitingKill = false
			m.runErr = nil
		}
		if msg.err == nil {
			if m.cancel != nil {
				m.cancel()
			}
			return m, m.startRunner()
		}
		return m, nil
	case runDoneMsg:
		m.runErr = msg.err
		if m.portConflict != nil && m.portConflict.Fatal {
			m.awaitingKill = true
			m.conflictNote = warnStyle.Render(m.portConflict.Message() + " — press k to kill & restart")
			return m, nil
		}
		m.done = true
		if m.cancel != nil {
			m.cancel()
		}
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.awaitingKill {
				m.done = true
			}
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		case "k", "K":
			if m.killPending || m.portConflict == nil {
				return m, nil
			}
			m.killPending = true
			port := m.portConflict.Port
			return m, killPortCmd(port)
		case "n", "N", "enter":
			if m.awaitingKill {
				m.awaitingKill = false
				m.portConflict = nil
				m.conflictNote = mutedStyle.Render("Port conflict ignored.")
				if m.runErr != nil {
					m.done = true
					return m, tea.Quit
				}
			}
			return m, nil
		}
		if m.ready && !m.awaitingKill {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
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
			m.viewport.SetContent(m.logContent())
			m.viewport.GotoBottom()
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewHeight
			m.viewport.SetContent(m.logContent())
		}
	}
	return m, nil
}

func (m *runnerModel) detectPortConflict(msg logMsg) {
	conflict, ok := portkill.ParseConflict(msg.text)
	if !ok {
		return
	}
	if m.portConflict != nil && m.portConflict.Port == conflict.Port && m.portConflict.Fatal {
		return
	}
	if m.portConflict == nil || conflict.Fatal || conflict.Port != m.portConflict.Port {
		m.portConflict = &conflict
	}
	if conflict.Fatal {
		m.conflictNote = warnStyle.Render(conflict.Message() + " — press k to kill & restart")
	} else if m.conflictNote == "" {
		m.conflictNote = mutedStyle.Render(conflict.Message() + " — press k to free port")
	}
}

func (m runnerModel) View() string {
	if m.width == 0 {
		return "Starting..."
	}

	status := fmt.Sprintf("Running: %s", strings.Join(m.serviceIDs, ", "))
	if m.awaitingKill {
		status = "Port conflict — action required"
	}
	header := renderHeader(m.cfg, m.width, status)

	var body string
	if m.ready {
		body = m.viewport.View()
	} else {
		body = mutedStyle.Render("Waiting for logs...")
	}

	footer := m.renderFooter()
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m runnerModel) renderFooter() string {
	if m.conflictNote != "" {
		hint := "k kill & restart  n ignore  q quit"
		if m.killPending {
			hint = "freeing port..."
		}
		return helpStyle.Render(m.conflictNote + "  |  " + hint)
	}
	base := "↑/↓ scroll  pgup/pgdn  q quit"
	if m.updateHint != "" {
		base = m.updateHint + "  |  " + base
	}
	return helpStyle.Render(base)
}

func (m *runnerModel) clearLogs() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lines = m.lines[:0]
}

func (m *runnerModel) appendLog(msg logMsg) {
	m.mu.Lock()
	defer m.mu.Unlock()

	style := lipgloss.NewStyle()
	if msg.stderr {
		style = errStyle
	}
	line := fmt.Sprintf("[%s] %s", msg.label, style.Render(msg.text))
	m.lines = append(m.lines, line)
	if len(m.lines) > maxLogLines {
		m.lines = m.lines[len(m.lines)-maxLogLines:]
	}
}

func (m runnerModel) logContent() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strings.Join(m.lines, "\n")
}

func waitForLog(ch chan logMsg) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func waitForDone(ch chan runDoneMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func killPortCmd(port int) tea.Cmd {
	return func() tea.Msg {
		killed, err := portkill.KillPort(port)
		return portKillMsg{port: port, killed: killed, err: err}
	}
}

var warnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
