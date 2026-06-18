package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
)

type ConfigureOptions struct {
	OutputPath string
	Force      bool
	Edit       bool
	WorkDir    string
}

type configurePhase int

const (
	phaseCfgWelcome configurePhase = iota
	phaseCfgName
	phaseCfgSubtitle
	phaseCfgServiceID
	phaseCfgServiceLabel
	phaseCfgServiceCommand
	phaseCfgServicePort
	phaseCfgServiceDeps
	phaseCfgAddAnother
	phaseCfgPreview
	phaseCfgConfirm
)

type configureModel struct {
	phase      configurePhase
	input      textinput.Model
	width      int
	height     int
	outputPath string
	force      bool
	errMsg     string
	done       bool
	aborted    bool

	name     string
	subtitle string
	services map[string]config.Service

	currentID      string
	currentService config.Service
	depCursor      int
	depSelected    map[string]bool
}

func RunConfigure(opts ConfigureOptions) error {
	if opts.OutputPath == "" {
		opts.OutputPath = config.DefaultFilename
	}
	if !opts.Force && config.Exists(opts.OutputPath) && !opts.Edit {
		return fmt.Errorf("%s already exists (use --force to overwrite)", opts.OutputPath)
	}

	model := newConfigureModel(opts)
	if opts.Edit && config.Exists(opts.OutputPath) {
		cfg, err := config.Load(opts.OutputPath)
		if err != nil {
			return err
		}
		model.name = cfg.Name
		model.subtitle = cfg.Subtitle
		model.services = cfg.Services
		model.phase = phaseCfgName
		model.input.SetValue(cfg.Name)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return err
	}
	m := final.(configureModel)
	if m.aborted {
		return ErrAborted
	}
	if m.errMsg != "" {
		return fmt.Errorf("%s", m.errMsg)
	}
	return nil
}

func newConfigureModel(opts ConfigureOptions) configureModel {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return configureModel{
		phase:      phaseCfgWelcome,
		input:      ti,
		outputPath: opts.OutputPath,
		force:      opts.Force,
		services:   make(map[string]config.Service),
		depSelected: make(map[string]bool),
	}
}

func (m configureModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m configureModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.phase {
		case phaseCfgWelcome:
			return m.handleWelcome(msg)
		case phaseCfgServiceDeps:
			return m.handleDeps(msg)
		case phaseCfgAddAnother, phaseCfgConfirm:
			return m.handleYesNo(msg)
		case phaseCfgPreview:
			return m.handlePreview(msg)
		default:
			return m.handleTextInput(msg)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m configureModel) handleWelcome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", " ":
		m.phase = phaseCfgName
		m.input.SetValue(m.name)
		m.input.Placeholder = "My App"
		m.input.Focus()
		return m, textinput.Blink
	case "ctrl+c", "q", "esc":
		m.aborted = true
		return m, tea.Quit
	}
	return m, nil
}

func (m configureModel) handleTextInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.aborted = true
		return m, tea.Quit
	case "enter":
		return m.advanceFromInput()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *configureModel) advanceFromInput() (tea.Model, tea.Cmd) {
	m.errMsg = ""
	value := strings.TrimSpace(m.input.Value())

	switch m.phase {
	case phaseCfgName:
		if value == "" {
			m.errMsg = "project name is required"
			return m, nil
		}
		m.name = value
		m.phase = phaseCfgSubtitle
		m.input.SetValue(m.subtitle)
		m.input.Placeholder = "Local development stack (optional)"
	case phaseCfgSubtitle:
		m.subtitle = value
		m.phase = phaseCfgServiceID
		m.input.SetValue("")
		m.input.Placeholder = "backend"
	case phaseCfgServiceID:
		id, err := config.NormalizeServiceID(value)
		if err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		if _, exists := m.services[id]; exists {
			m.errMsg = fmt.Sprintf("service %q already exists", id)
			return m, nil
		}
		m.currentID = id
		m.currentService = config.Service{DependsOn: []string{}}
		m.phase = phaseCfgServiceLabel
		m.input.SetValue("")
		m.input.Placeholder = "Backend"
	case phaseCfgServiceLabel:
		if value == "" {
			m.errMsg = "service label is required"
			return m, nil
		}
		m.currentService.Label = value
		m.phase = phaseCfgServiceCommand
		m.input.SetValue("")
		m.input.Placeholder = "npm run dev"
	case phaseCfgServiceCommand:
		if value == "" {
			m.errMsg = "command is required"
			return m, nil
		}
		m.currentService.Command = value
		m.phase = phaseCfgServicePort
		m.input.SetValue(m.currentService.Port)
		m.input.Placeholder = "${PORT} (optional)"
	case phaseCfgServicePort:
		m.currentService.Port = value
		if len(m.services) == 0 {
			m.finalizeCurrentService()
			m.phase = phaseCfgAddAnother
			return m, nil
		}
		m.depCursor = 0
		m.depSelected = make(map[string]bool)
		m.phase = phaseCfgServiceDeps
	default:
		return m, nil
	}
	m.input.Focus()
	return m, textinput.Blink
}

func (m *configureModel) finalizeCurrentService() {
	deps := make([]string, 0)
	for id := range m.depSelected {
		if m.depSelected[id] {
			deps = append(deps, id)
		}
	}
	m.currentService.DependsOn = deps
	m.services[m.currentID] = m.currentService
	m.currentID = ""
	m.currentService = config.Service{}
}

func (m configureModel) handleDeps(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ids := m.sortedServiceIDs()
	switch msg.String() {
	case "ctrl+c", "esc", "q":
		m.aborted = true
		return m, tea.Quit
	case "up", "k":
		if m.depCursor > 0 {
			m.depCursor--
		}
	case "down", "j":
		if m.depCursor < len(ids)-1 {
			m.depCursor++
		}
	case " ":
		if len(ids) > 0 {
			id := ids[m.depCursor]
			m.depSelected[id] = !m.depSelected[id]
		}
	case "enter":
		m.finalizeCurrentService()
		m.phase = phaseCfgAddAnother
	}
	return m, nil
}

func (m configureModel) handleYesNo(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc", "q":
		m.aborted = true
		return m, tea.Quit
	case "y", "Y":
		if m.phase == phaseCfgAddAnother {
			m.phase = phaseCfgServiceID
			m.input.SetValue("")
			m.input.Placeholder = "ui"
			m.input.Focus()
			return m, textinput.Blink
		}
		return m.saveAndQuit()
	case "n", "N":
		if m.phase == phaseCfgAddAnother {
			if len(m.services) == 0 {
				m.errMsg = "add at least one service"
				return m, nil
			}
			m.phase = phaseCfgPreview
			return m, nil
		}
		m.aborted = true
		return m, tea.Quit
	case "enter":
		if m.phase == phaseCfgAddAnother {
			if len(m.services) == 0 {
				m.errMsg = "add at least one service"
				return m, nil
			}
			m.phase = phaseCfgPreview
			return m, nil
		}
	}
	return m, nil
}

func (m configureModel) handlePreview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc", "q":
		m.aborted = true
		return m, tea.Quit
	case "enter":
		m.phase = phaseCfgConfirm
		return m, nil
	}
	return m, nil
}

func (m *configureModel) saveAndQuit() (tea.Model, tea.Cmd) {
	cfg := &config.Config{
		Name:     m.name,
		Subtitle: m.subtitle,
		Services: m.services,
	}
	if err := config.Save(m.outputPath, cfg); err != nil {
		m.errMsg = err.Error()
		m.phase = phaseCfgPreview
		return m, nil
	}
	m.done = true
	return m, tea.Quit
}

func (m configureModel) sortedServiceIDs() []string {
	ids := make([]string, 0, len(m.services))
	for id := range m.services {
		ids = append(ids, id)
	}
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			if ids[j] < ids[i] {
				ids[i], ids[j] = ids[j], ids[i]
			}
		}
	}
	return ids
}

func (m configureModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder
	b.WriteString(renderHeader(&config.Config{Name: "muxdev configure", Subtitle: "Interactive configuration"}, m.width, m.phaseTitle()))
	b.WriteString("\n")

	if m.errMsg != "" {
		b.WriteString(errStyle.Render("Error: "+m.errMsg) + "\n\n")
	}

	switch m.phase {
	case phaseCfgWelcome:
		b.WriteString("Create a new muxdev.yaml step by step.\n\n")
		b.WriteString(mutedStyle.Render("Press enter to start, q to quit"))
	case phaseCfgServiceDeps:
		b.WriteString(fmt.Sprintf("Dependencies for %q (optional)\n\n", m.currentID))
		ids := m.sortedServiceIDs()
		for i, id := range ids {
			marker := "  "
			if i == m.depCursor {
				marker = cursorStyle.Render("> ")
			}
			check := "[ ]"
			if m.depSelected[id] {
				check = selectedStyle.Render("[x]")
			}
			b.WriteString(fmt.Sprintf("%s%s %s\n", marker, check, id))
		}
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("↑/↓ move  space toggle  enter continue"))
	case phaseCfgAddAnother:
		b.WriteString(fmt.Sprintf("Saved service. Total: %d\n\n", len(m.services)))
		b.WriteString("Add another service?\n")
		b.WriteString(helpStyle.Render("y yes  n no / enter"))
	case phaseCfgPreview:
		b.WriteString(m.previewYAML())
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("enter to save  q quit"))
	case phaseCfgConfirm:
		b.WriteString(fmt.Sprintf("Write %s?\n\n", m.outputPath))
		b.WriteString(helpStyle.Render("y save  n cancel"))
	default:
		b.WriteString(m.phasePrompt())
		b.WriteString("\n\n")
		b.WriteString(m.input.View())
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("enter continue  esc quit"))
	}

	if m.done {
		b.WriteString("\n\n")
		b.WriteString(selectedStyle.Render("Saved "+m.outputPath))
	}

	return b.String()
}

func (m configureModel) phaseTitle() string {
	switch m.phase {
	case phaseCfgWelcome:
		return "Welcome"
	case phaseCfgName:
		return "Project name"
	case phaseCfgSubtitle:
		return "Subtitle"
	case phaseCfgServiceID:
		return "Service ID"
	case phaseCfgServiceLabel:
		return "Service label"
	case phaseCfgServiceCommand:
		return "Command"
	case phaseCfgServicePort:
		return "Port"
	case phaseCfgServiceDeps:
		return "Dependencies"
	case phaseCfgAddAnother:
		return "Add service"
	case phaseCfgPreview:
		return "Preview"
	case phaseCfgConfirm:
		return "Confirm"
	default:
		return "Configure"
	}
}

func (m configureModel) phasePrompt() string {
	switch m.phase {
	case phaseCfgName:
		return "Project name:"
	case phaseCfgSubtitle:
		return "Subtitle (optional):"
	case phaseCfgServiceID:
		return "Service ID (e.g. backend, ui):"
	case phaseCfgServiceLabel:
		return fmt.Sprintf("Label for %q:", m.currentID)
	case phaseCfgServiceCommand:
		return fmt.Sprintf("Command for %q:", m.currentID)
	case phaseCfgServicePort:
		return fmt.Sprintf("Port for %q (optional):", m.currentID)
	default:
		return ""
	}
}

func (m configureModel) previewYAML() string {
	cfg := &config.Config{
		Name:     m.name,
		Subtitle: m.subtitle,
		Services: m.services,
	}
	data, err := config.Format(cfg)
	if err != nil {
		return errStyle.Render(err.Error())
	}
	return string(data)
}
