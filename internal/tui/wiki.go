package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yarkingulacti/muxdev-cli/internal/cli/doc"
)

type WikiOptions struct {
	Pages     []doc.Page
	OpenTopic string // open directly on this topic (e.g. from --help on a subcommand)
}

type wikiPhase int

const (
	wikiPhaseBrowse wikiPhase = iota
	wikiPhaseRead
	wikiPhaseSearch
)

type wikiModel struct {
	pages      []doc.Page
	filtered   []doc.Page
	phase      wikiPhase
	cursor     int
	readIndex  int
	width      int
	height     int
	ready      bool
	viewport   viewport.Model
	tryView    viewport.Model
	search     textinput.Model
	tryOutput  string
	tryPending bool
	tryErr     error
	pendingTopic string
}

type wikiTryResultMsg struct {
	output string
	err    error
}

func RunWiki(opts WikiOptions) error {
	pages := opts.Pages
	if len(pages) == 0 {
		pages = []doc.Page{{ID: "empty", Category: "Start here", Title: "No topics", Body: "No help pages registered."}}
	}
	m := wikiModel{
		pages:        pages,
		filtered:     append([]doc.Page(nil), pages...),
		search:       textinput.New(),
		pendingTopic: strings.TrimSpace(opts.OpenTopic),
	}
	m.search.Prompt = "search: "
	m.search.Placeholder = "topic, command, keyword…"
	m.search.CharLimit = 64
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m wikiModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m wikiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case wikiTryResultMsg:
		m.tryPending = false
		m.tryOutput = strings.TrimSpace(msg.output)
		m.tryErr = msg.err
		if m.tryOutput == "" && msg.err != nil {
			m.tryOutput = msg.err.Error()
		}
		m.layoutViewports()
		return m, nil
	case tea.KeyMsg:
		if m.phase == wikiPhaseSearch {
			switch msg.String() {
			case "esc":
				m.phase = wikiPhaseBrowse
				m.search.Blur()
				m.filtered = append([]doc.Page(nil), m.pages...)
				m.cursor = 0
				return m, nil
			case "enter":
				m.phase = wikiPhaseBrowse
				m.search.Blur()
				if len(m.filtered) > 0 {
					m.cursor = 0
				}
				return m, nil
			}
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			m.filtered = doc.MatchPages(m.pages, m.search.Value())
			if m.cursor >= len(m.filtered) {
				m.cursor = max(0, len(m.filtered)-1)
			}
			return m, cmd
		}

		if m.phase == wikiPhaseRead {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "esc", "b":
				m.phase = wikiPhaseBrowse
				m.tryOutput = ""
				m.tryErr = nil
				return m, nil
			case "t", "T":
				if page, ok := m.currentReadPage(); ok && page.TryCommand != "" {
					m.tryPending = true
					m.tryOutput = "Running…"
					m.layoutViewports()
					return m, wikiTryCmd(page.TryCommand)
				}
				return m, nil
			case "n":
				if m.readIndex < len(m.filtered)-1 {
					m.readIndex++
					m.openRead(m.readIndex)
				}
				return m, nil
			case "p":
				if m.readIndex > 0 {
					m.readIndex--
					m.openRead(m.readIndex)
				}
				return m, nil
			}
			if m.ready && handleLogScrollViewport(&m.viewport, nil, msg) {
				return m, nil
			}
			if m.tryOutput != "" && m.ready && handleLogScrollViewport(&m.tryView, nil, msg) {
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "/":
			m.phase = wikiPhaseSearch
			m.search.SetValue("")
			m.search.Focus()
			return m, textinput.Blink
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.filtered) == 0 {
				return m, nil
			}
			m.openRead(m.cursor)
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layoutViewports()
		if m.pendingTopic != "" {
			topic := m.pendingTopic
			m.pendingTopic = ""
			if idx := doc.FindPageIndex(m.filtered, topic); idx >= 0 {
				m.openRead(idx)
			}
		}
	}
	return m, nil
}

func (m *wikiModel) layoutViewports() {
	if m.width == 0 || m.height == 0 {
		return
	}
	header := 5
	footer := 2
	tryH := 0
	if m.phase == wikiPhaseRead && m.tryOutput != "" {
		tryH = 8
	}
	contentH := m.height - header - footer - tryH
	if contentH < 4 {
		contentH = 4
	}
	if !m.ready {
		m.viewport = viewport.New(m.width, contentH)
		m.tryView = viewport.New(m.width, tryH)
		m.viewport.KeyMap = runnerLogViewportKeyMap()
		m.tryView.KeyMap = runnerLogViewportKeyMap()
		m.ready = true
	} else {
		m.viewport.Width = m.width
		m.viewport.Height = contentH
		m.tryView.Width = m.width
		m.tryView.Height = tryH
	}
	if m.phase == wikiPhaseRead {
		if page, ok := m.currentReadPage(); ok {
			m.viewport.SetContent(renderWikiBody(page.Title, page.Body))
			m.viewport.GotoTop()
		}
		if m.tryOutput != "" {
			label := "Try output"
			if m.tryErr != nil {
				label = "Try output (error)"
			}
			m.tryView.SetContent(renderWikiBody(label, m.tryOutput))
			m.tryView.GotoBottom()
		}
	}
}

func (m *wikiModel) openRead(index int) {
	if index < 0 || index >= len(m.filtered) {
		return
	}
	m.readIndex = index
	m.phase = wikiPhaseRead
	m.tryOutput = ""
	m.tryErr = nil
	m.layoutViewports()
}

func (m wikiModel) currentReadPage() (doc.Page, bool) {
	if m.readIndex < 0 || m.readIndex >= len(m.filtered) {
		return doc.Page{}, false
	}
	return m.filtered[m.readIndex], true
}

func (m wikiModel) View() string {
	if m.width == 0 {
		return "Loading muxdev guide…"
	}

	switch m.phase {
	case wikiPhaseSearch:
		return m.renderSearch()
	case wikiPhaseRead:
		return m.renderRead()
	default:
		return m.renderBrowse()
	}
}

func (m wikiModel) renderBrowse() string {
	header := renderWikiHeader(m.width, "Local guide — pick a topic")
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")

	lastCat := ""
	for i, page := range m.filtered {
		if page.Category != lastCat {
			lastCat = page.Category
			b.WriteString("\n")
			b.WriteString(titleStyle.Render(page.Category))
			b.WriteString("\n")
		}
		marker := "  "
		line := page.Title
		if i == m.cursor {
			marker = cursorStyle.Render("› ")
			line = selectedStyle.Render(line)
		}
		hint := mutedStyle.Render("  " + page.ID)
		b.WriteString(marker + line + hint + "\n")
	}

	if len(m.filtered) == 0 {
		b.WriteString(mutedStyle.Render("\nNo topics match.\n"))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ move  enter open  / search  q quit"))
	return b.String()
}

func (m wikiModel) renderSearch() string {
	header := renderWikiHeader(m.width, "Search topics")
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(m.search.View())
	b.WriteString("\n\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("%d topics", len(m.filtered))))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("enter apply  esc cancel"))
	return b.String()
}

func (m wikiModel) renderRead() string {
	page, ok := m.currentReadPage()
	if !ok {
		return ""
	}
	header := renderWikiHeader(m.width, page.Category+" · "+page.Title)
	var parts []string
	parts = append(parts, header, m.viewport.View())
	if m.tryOutput != "" {
		parts = append(parts, m.tryView.View())
	}
	footer := helpStyle.Render("t try  n/p next/prev  pgup/pgdn scroll  esc back  q quit")
	if page.TryCommand != "" {
		footer = helpStyle.Render("t → "+page.TryCommand+"  n/p  scroll  esc  q")
	}
	parts = append(parts, footer)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func renderWikiHeader(width int, status string) string {
	title := titleStyle.Render("muxdev wiki")
	sub := mutedStyle.Render(status)
	content := fmt.Sprintf("%s\n%s", title, sub)
	return cardStyle.Width(min(width-2, 72)).Render(content)
}

func renderWikiBody(title, body string) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")
	for _, line := range strings.Split(body, "\n") {
		trim := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trim, "  --"):
			b.WriteString(mutedStyle.Render(line))
		case strings.HasSuffix(trim, ":") && !strings.HasPrefix(line, " "):
			b.WriteString(subtitleStyle.Render(line))
		default:
			b.WriteString(line)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func wikiTryCmd(line string) tea.Cmd {
	return func() tea.Msg {
		exe, err := os.Executable()
		if err != nil {
			return wikiTryResultMsg{err: err}
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			return wikiTryResultMsg{err: fmt.Errorf("empty command")}
		}
		if fields[0] == "muxdev" {
			fields[0] = exe
		}
		cmd := exec.Command(fields[0], fields[1:]...)
		cmd.Dir = mustGetwd()
		out, err := cmd.CombinedOutput()
		return wikiTryResultMsg{output: string(out), err: err}
	}
}

func mustGetwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
