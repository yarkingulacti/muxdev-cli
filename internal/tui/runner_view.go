package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
	"github.com/yarkingulacti/muxdev-cli/internal/logs"
	"github.com/yarkingulacti/muxdev-cli/internal/portkill"
	"github.com/yarkingulacti/muxdev-cli/internal/runner"
)

const maxLogLines = 5000

type logEntry struct {
	label  string
	stderr bool
	text   string
}

type logMsg struct {
	label  string
	stderr bool
	text   string
}

type runDoneMsg struct {
	err error
}

type runnerStartedMsg struct {
	cancel  context.CancelFunc
	logCh   chan logMsg
	doneCh  chan runDoneMsg
	session *logs.Writer
}

type portKillMsg struct {
	port   int
	killed int
	err    error
}

type portAttachMsg struct {
	port    int
	process portkill.Process
	label   string
	err     error
}

type attachStartedMsg struct {
	cancel context.CancelFunc
	logCh  chan logMsg
	doneCh chan attachDoneMsg
}

type attachDoneMsg struct {
	err error
}

type runnerModel struct {
	cfg        *config.Config
	configPath string
	serviceIDs []string
	workDir    string
	viewport   viewport.Model
	entries    []logEntry
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

	runtime      config.Runtime
	filterMenu   bool
	filterCursor int
	filterLabel  string

	portConflict *portkill.Conflict
	conflictNote string
	awaitingKill bool
	killPending  bool
	attachPending bool
	attached     bool
	attachLabel  string
	attachCancel context.CancelFunc
	attachLogCh  chan logMsg
	attachDoneCh chan attachDoneMsg
	shuttingDown bool
	followTail   bool
	session      *logs.Writer
}

func runLogs(cfg *config.Config, configPath string, serviceIDs []string, workDir, updateHint string, runtime config.Runtime) error {
	model := newRunnerModel(cfg, configPath, serviceIDs, workDir, updateHint, runtime)
	p := tea.NewProgram(model, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return err
	}
	m := final.(runnerModel)
	return m.runErr
}

func newRunnerModel(cfg *config.Config, configPath string, serviceIDs []string, workDir, updateHint string, runtime config.Runtime) runnerModel {
	if runtime == "" {
		runtime = config.DefaultRuntime
	}
	return runnerModel{
		cfg:        cfg,
		configPath: configPath,
		serviceIDs: serviceIDs,
		workDir:    workDir,
		runtime:    runtime,
		entries:    make([]logEntry, 0, 256),
		updateHint: updateHint,
		followTail: true,
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

		session, err := logs.StartSession(m.workDir, m.configPath, m.serviceIDs, string(m.runtime))
		if err != nil {
			session = nil
		}

		go func() {
			r := runner.New(m.cfg, m.serviceIDs, m.runtime)
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
			cancel:  cancel,
			logCh:   logCh,
			doneCh:  doneCh,
			session: session,
		}
	}
}

func (m runnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case runnerStartedMsg:
		m.cancel = msg.cancel
		m.logCh = msg.logCh
		m.doneCh = msg.doneCh
		if m.session != nil {
			_ = m.session.Finish(nil)
		}
		m.session = msg.session
		m.done = false
		m.conflictNote = ""
		m.followTail = true
		m.refreshLogViewport()
		return m, tea.Batch(waitForLog(m.logCh), waitForDone(m.doneCh))
	case logMsg:
		m.appendLog(msg)
		if !m.attached {
			m.detectPortConflict(msg)
		}
		m.refreshLogViewport()
		if m.attached {
			return m, waitForAttachLog(m.attachLogCh)
		}
		return m, waitForLog(m.logCh)
	case portKillMsg:
		m.killPending = false
		if msg.err != nil {
			m.conflictNote = errStyle.Render(fmt.Sprintf(
				"Could not free port %d: %v",
				msg.port, msg.err,
			))
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
	case portAttachMsg:
		m.attachPending = false
		if msg.err != nil {
			m.conflictNote = errStyle.Render(fmt.Sprintf("Could not attach to port %d: %v", msg.port, msg.err))
			return m, nil
		}
		if m.cancel != nil {
			m.cancel()
		}
		m.clearLogs()
		m.portConflict = nil
		m.awaitingKill = false
		m.runErr = nil
		m.attached = true
		m.attachLabel = msg.label
		m.conflictNote = mutedStyle.Render(fmt.Sprintf(
			"Attached to PID %d on port %d — %s",
			msg.process.PID,
			msg.port,
			msg.process.Command,
		))
		return m, m.startAttach(msg.process.PID, msg.label)
	case attachStartedMsg:
		m.attachCancel = msg.cancel
		m.attachLogCh = msg.logCh
		m.attachDoneCh = msg.doneCh
		m.refreshLogViewport()
		return m, tea.Batch(waitForAttachLog(m.attachLogCh), waitForAttachDone(m.attachDoneCh))
	case attachDoneMsg:
		if !m.attached {
			return m, nil
		}
		m.attached = false
		if m.attachCancel != nil {
			m.attachCancel()
			m.attachCancel = nil
		}
		if msg.err != nil {
			m.conflictNote = errStyle.Render(fmt.Sprintf("Attach ended: %v", msg.err))
		} else {
			m.conflictNote = mutedStyle.Render("Attached process exited.")
		}
		return m, nil
	case runDoneMsg:
		m.drainPendingLogs()
		m.runErr = msg.err
		if m.session != nil {
			_ = m.session.Finish(msg.err)
			m.session = nil
		}
		if m.attachPending {
			return m, nil
		}
		if m.shuttingDown {
			return m, tea.Quit
		}
		if m.portConflict != nil && m.portConflict.Fatal {
			m.awaitingKill = true
			m.conflictNote = warnStyle.Render(m.portConflict.Message() + conflictActionHint(true))
			return m, nil
		}
		m.done = true
		if m.cancel != nil {
			m.cancel()
		}
		return m, tea.Quit
	case tea.KeyMsg:
		if m.filterMenu {
			switch msg.String() {
			case "esc":
				m.filterMenu = false
				return m, nil
			case "f":
				m.filterMenu = false
				return m, nil
			case "up", "k":
				if m.filterCursor > 0 {
					m.filterCursor--
				}
			case "down", "j":
				items := m.filterMenuItems()
				if m.filterCursor < len(items)-1 {
					m.filterCursor++
				}
			case "enter":
				items := m.filterMenuItems()
				if m.filterCursor >= 0 && m.filterCursor < len(items) {
					m.filterLabel = items[m.filterCursor].label
				}
				m.filterMenu = false
				m.refreshLogViewport()
				return m, nil
			}
			return m, nil
		}

		if m.ready && !m.awaitingKill && m.handleLogScroll(msg) {
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			if m.awaitingKill {
				m.done = true
			}
			if m.attachCancel != nil {
				m.attachCancel()
			}
			if m.cancel != nil {
				m.shuttingDown = true
				m.cancel()
			}
			if m.doneCh == nil {
				return m, tea.Quit
			}
			return m, nil
		case "a", "A":
			if m.attachPending || m.killPending || m.portConflict == nil || !m.portConflict.Fatal {
				return m, nil
			}
			m.attachPending = true
			port := m.portConflict.Port
			return m, attachPortCmd(m.cfg, m.serviceIDs, m.workDir, port)
		case "k", "K":
			if m.killPending || m.portConflict == nil {
				return m, nil
			}
			m.killPending = true
			port := m.portConflict.Port
			return m, killPortCmd(port, m.cancel)
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
		case "f", "F":
			if m.awaitingKill || m.attachPending || m.killPending {
				return m, nil
			}
			m.filterMenu = true
			m.filterCursor = m.filterMenuIndex()
			return m, nil
		}
		if m.ready && !m.awaitingKill && !logScrollKey(msg) {
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
			m.viewport.KeyMap = runnerLogViewportKeyMap()
			m.refreshLogViewport()
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewHeight
			m.refreshLogViewport()
		}
	}
	return m, nil
}

func (m *runnerModel) drainPendingLogs() {
	for m.logCh != nil {
		log, ok := <-m.logCh
		if !ok {
			m.logCh = nil
			return
		}
		m.appendLog(log)
		if !m.attached {
			m.detectPortConflict(log)
		}
	}
}

func (m *runnerModel) detectPortConflict(msg logMsg) {
	conflict, ok := portkill.ParseConflict(msg.text, m.hintPortForLog(msg))
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
		m.conflictNote = warnStyle.Render(conflict.Message() + conflictActionHint(true))
	} else if m.conflictNote == "" {
		m.conflictNote = mutedStyle.Render(conflict.Message() + " — press k to free port")
	}
}

func (m runnerModel) hintPortForLog(msg logMsg) int {
	label := strings.TrimSuffix(msg.label, "!")
	for _, id := range m.serviceIDs {
		svc, ok := m.cfg.Services[id]
		if !ok {
			continue
		}
		if label == serviceLogLabel(m.cfg, id) || label == id {
			if port := config.BindPortForService(m.workDir, svc); port > 0 {
				return port
			}
		}
	}
	return 0
}

func conflictActionHint(fatal bool) string {
	if fatal {
		return " — a attach  k kill & restart  n ignore"
	}
	return ""
}

func (m runnerModel) View() string {
	if m.width == 0 {
		return "Starting..."
	}

	status := fmt.Sprintf("Running: %s", strings.Join(m.serviceIDs, ", "))
	status += mutedStyle.Render("  ·  " + string(m.runtime))
	if m.filterLabel != "" {
		status += mutedStyle.Render("  ·  filter: " + m.filterLabel)
	}
	if m.attached {
		status = fmt.Sprintf("Attached: %s", m.attachLabel)
	} else if m.awaitingKill {
		status = "Port conflict — action required"
	} else if !m.followTail {
		status += mutedStyle.Render("  ·  history")
	}
	header := renderHeader(m.cfg, m.width, status)

	var body string
	if m.filterMenu {
		body = m.renderFilterMenu()
	} else if m.ready {
		body = m.viewport.View()
	} else {
		body = mutedStyle.Render("Waiting for logs...")
	}

	footer := m.renderFooter()
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m runnerModel) renderFooter() string {
	if m.filterMenu {
		return helpStyle.Render("↑/↓ select  enter apply  esc cancel")
	}
	if m.attached {
		base := logScrollHelpAttached
		if pag := m.logPaginationLabel(); pag != "" {
			base = pag + "  |  " + base
		}
		if m.conflictNote != "" {
			return helpStyle.Render(m.conflictNote + "  |  " + base)
		}
		return helpStyle.Render(base)
	}
	if m.conflictNote != "" {
		hint := "a attach  k kill & restart  n ignore  q quit"
		if m.portConflict != nil && !m.portConflict.Fatal {
			hint = "k free port  q quit"
		}
		if m.killPending {
			hint = "freeing port..."
		}
		if m.attachPending {
			hint = "attaching..."
		}
		line := m.conflictNote + "  |  " + hint
		if pag := m.logPaginationLabel(); pag != "" {
			line = m.conflictNote + "  |  " + pag + "  |  " + hint
		}
		return helpStyle.Render(line)
	}
	base := logScrollHelp
	if !m.followTail {
		base = logScrollHelpHistory
	}
	if m.filterLabel != "" {
		base = "pgup/pgdn line  ctrl+u/d page  f filter (" + m.filterLabel + ")  q quit"
	}
	if m.updateHint != "" {
		base = m.updateHint + "  |  " + base
	}
	if pag := m.logPaginationLabel(); pag != "" {
		base = pag + "  |  " + base
	}
	return helpStyle.Render(base)
}

func (m *runnerModel) clearLogs() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = m.entries[:0]
	m.followTail = true
}

func (m *runnerModel) appendLog(msg logMsg) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries = append(m.entries, logEntry{
		label:  msg.label,
		stderr: msg.stderr,
		text:   msg.text,
	})
	if len(m.entries) > maxLogLines {
		m.entries = m.entries[len(m.entries)-maxLogLines:]
	}
	if m.session != nil {
		_ = m.session.Append(msg.label, msg.text)
	}
}

func (m runnerModel) logContent() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	lines := make([]string, 0, len(m.entries))
	for _, entry := range m.entries {
		if m.filterLabel != "" && entry.label != m.filterLabel {
			continue
		}
		style := lipgloss.NewStyle()
		if entry.stderr {
			style = errStyle
		}
		line := fmt.Sprintf("[%s] %s", entry.label, style.Render(entry.text))
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return mutedStyle.Render("No logs yet.")
	}
	return strings.Join(lines, "\n")
}

func (m *runnerModel) refreshLogViewport() {
	if !m.ready || m.filterMenu {
		return
	}
	offset := m.viewport.YOffset
	m.viewport.SetContent(m.logContent())
	if m.followTail {
		m.viewport.GotoBottom()
	} else {
		m.viewport.SetYOffset(offset)
	}
}

type filterMenuItem struct {
	title string
	label string
}

func (m runnerModel) filterMenuItems() []filterMenuItem {
	items := []filterMenuItem{{title: "All services", label: ""}}
	for _, id := range m.serviceIDs {
		label := serviceLogLabel(m.cfg, id)
		items = append(items, filterMenuItem{
			title: fmt.Sprintf("%s (%s)", label, id),
			label: label,
		})
	}
	return items
}

func (m runnerModel) filterMenuIndex() int {
	if m.filterLabel == "" {
		return 0
	}
	items := m.filterMenuItems()
	for i, item := range items {
		if item.label == m.filterLabel {
			return i
		}
	}
	return 0
}

func (m runnerModel) renderFilterMenu() string {
	items := m.filterMenuItems()
	var b strings.Builder
	b.WriteString(titleStyle.Render("Filter logs by service"))
	b.WriteString("\n\n")
	for i, item := range items {
		marker := "  "
		if i == m.filterCursor {
			marker = cursorStyle.Render("> ")
		}
		line := item.title
		if i == m.filterCursor {
			line = selectedStyle.Render(line)
		}
		if item.label != "" && item.label == m.filterLabel {
			line += mutedStyle.Render("  (active)")
		}
		b.WriteString(marker + line + "\n")
	}
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("↑/↓ select  enter apply  esc cancel"))
	width := min(m.width-2, 56)
	if width < 20 {
		width = 20
	}
	return cardStyle.Width(width).Render(b.String())
}

func serviceLogLabel(cfg *config.Config, id string) string {
	svc := cfg.Services[id]
	if strings.TrimSpace(svc.Label) != "" {
		return svc.Label
	}
	return id
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

func killPortCmd(port int, cancel context.CancelFunc) tea.Cmd {
	return func() tea.Msg {
		if cancel != nil {
			cancel()
			deadline := time.Now().Add(1200 * time.Millisecond)
			for time.Now().Before(deadline) {
				time.Sleep(100 * time.Millisecond)
				pids, err := portkill.PIDsOnPort(port)
				if err != nil || len(pids) == 0 {
					break
				}
			}
		}
		killed, err := portkill.KillPort(port)
		return portKillMsg{port: port, killed: killed, err: err}
	}
}

func attachPortCmd(cfg *config.Config, serviceIDs []string, workDir string, port int) tea.Cmd {
	return func() tea.Msg {
		proc, err := portkill.ProcessOnPort(port)
		if err != nil {
			return portAttachMsg{port: port, err: err}
		}
		return portAttachMsg{
			port:    port,
			process: proc,
			label:   serviceLabelForPort(cfg, serviceIDs, workDir, port),
		}
	}
}

func (m runnerModel) startAttach(pid int, label string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		logCh := make(chan logMsg, 128)
		doneCh := make(chan attachDoneMsg, 1)

		go func() {
			err := portkill.AttachProcess(ctx, pid, func(stderr bool, text string) {
				select {
				case logCh <- logMsg{label: label, stderr: stderr, text: text}:
				case <-ctx.Done():
				}
			})
			close(logCh)
			doneCh <- attachDoneMsg{err: err}
		}()

		return attachStartedMsg{
			cancel: cancel,
			logCh:  logCh,
			doneCh: doneCh,
		}
	}
}

func serviceLabelForPort(cfg *config.Config, serviceIDs []string, workDir string, port int) string {
	portStr := strconv.Itoa(port)
	for _, id := range serviceIDs {
		svc := cfg.Services[id]
		if config.ExpandServicePort(cfg, workDir, svc) == portStr {
			if strings.TrimSpace(svc.Label) != "" {
				return svc.Label
			}
			return id
		}
	}
	return fmt.Sprintf("port:%d", port)
}

func waitForAttachLog(ch chan logMsg) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return attachDoneMsg{}
		}
		return msg
	}
}

func waitForAttachDone(ch chan attachDoneMsg) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		return <-ch
	}
}

var warnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
