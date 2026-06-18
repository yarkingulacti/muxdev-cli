package tui

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
)

var ErrAborted = errors.New("aborted")

type Options struct {
	Cfg        *config.Config
	Focus      []string
	WorkDir    string
	UpdateHint string
}

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	subtitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	cardStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	mutedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func Run(opts Options) error {
	serviceIDs := opts.Focus
	if len(serviceIDs) == 0 {
		picked, err := runPicker(opts.Cfg, opts.UpdateHint)
		if err != nil {
			return err
		}
		serviceIDs = picked
	} else {
		resolved, err := opts.Cfg.ResolveServices(serviceIDs)
		if err != nil {
			return err
		}
		serviceIDs = resolved
	}

	return runLogs(opts.Cfg, serviceIDs, opts.WorkDir, opts.UpdateHint)
}

func runPicker(cfg *config.Config, updateHint string) ([]string, error) {
	ids, err := cfg.SortedServiceIDs()
	if err != nil {
		return nil, err
	}

	model := newPickerModel(cfg, ids, updateHint)
	p := tea.NewProgram(model, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	m := final.(pickerModel)
	if m.aborted {
		return nil, ErrAborted
	}
	if len(m.chosen) == 0 {
		return nil, ErrAborted
	}
	return cfg.ResolveServices(m.chosen)
}

type pickerModel struct {
	cfg        *config.Config
	ids        []string
	cursor     int
	selected   map[string]bool
	width      int
	height     int
	chosen     []string
	aborted    bool
	updateHint string
}

func newPickerModel(cfg *config.Config, ids []string, updateHint string) pickerModel {
	selected := make(map[string]bool, len(ids))
	for _, id := range ids {
		selected[id] = true
	}
	return pickerModel{
		cfg:        cfg,
		ids:        ids,
		selected:   selected,
		updateHint: updateHint,
	}
}

func (m pickerModel) Init() tea.Cmd {
	return nil
}

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.aborted = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.ids)-1 {
				m.cursor++
			}
		case " ":
			id := m.ids[m.cursor]
			m.selected[id] = !m.selected[id]
		case "a":
			all := !m.allSelected()
			for _, id := range m.ids {
				m.selected[id] = all
			}
		case "enter":
			chosen := make([]string, 0, len(m.ids))
			for _, id := range m.ids {
				if m.selected[id] {
					chosen = append(chosen, id)
				}
			}
			if len(chosen) == 0 {
				return m, nil
			}
			m.chosen = chosen
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m pickerModel) allSelected() bool {
	for _, id := range m.ids {
		if !m.selected[id] {
			return false
		}
	}
	return true
}

func (m pickerModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := renderHeader(m.cfg, m.width, "Select services to run")
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")

	for i, id := range m.ids {
		svc := m.cfg.Services[id]
		marker := "  "
		if i == m.cursor {
			marker = cursorStyle.Render("> ")
		}

		check := "[ ]"
		if m.selected[id] {
			check = selectedStyle.Render("[x]")
		}

		line := fmt.Sprintf("%s%s %s (%s)", marker, check, svc.Label, id)
		if len(svc.DependsOn) > 0 {
			line += mutedStyle.Render("  depends: "+strings.Join(svc.DependsOn, ", "))
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	help := "↑/↓ move  space toggle  a all  enter start  q quit"
	if m.updateHint != "" {
		help = m.updateHint + "  |  " + help
	}
	b.WriteString(helpStyle.Render(help))
	return b.String()
}

func renderHeader(cfg *config.Config, width int, status string) string {
	title := titleStyle.Render(cfg.Name)
	subtitle := ""
	if cfg.Subtitle != "" {
		subtitle = subtitleStyle.Render(" — " + cfg.Subtitle)
	}
	statusLine := mutedStyle.Render(status)
	content := fmt.Sprintf("%s%s\n%s", title, subtitle, statusLine)
	return cardStyle.Width(min(width-2, 72)).Render(content)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
