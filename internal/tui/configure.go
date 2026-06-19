package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
	"github.com/yarkingulacti/muxdev-cli/internal/portdiscover"
)

type ConfigureOptions struct {
	OutputPath string
	Force      bool
	Edit       bool
	Init       bool
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
	phaseCfgServicePortDiscover
	phaseCfgServicePort
	phaseCfgServiceDeps
	phaseCfgAddAnother
	phaseCfgPreview
	phaseCfgConfirm
	phaseCfgDone
)

type configureModel struct {
	phase      configurePhase
	input      textinput.Model
	width      int
	height     int
	outputPath string
	force      bool
	init       bool
	workDir    string
	errMsg     string
	done       bool
	aborted    bool

	portDiscoverNote string

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
		phase:       phaseCfgWelcome,
		input:       ti,
		outputPath:  opts.OutputPath,
		force:       opts.Force,
		init:        opts.Init,
		workDir:     opts.WorkDir,
		services:    make(map[string]config.Service),
		depSelected: make(map[string]bool),
	}
}

func (m configureModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m configureModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case portDiscoverMsg:
		return m.handlePortDiscover(msg)
	case tea.KeyMsg:
		switch m.phase {
		case phaseCfgWelcome:
			return m.handleWelcome(msg)
		case phaseCfgServiceDeps:
			return m.handleDeps(msg)
		case phaseCfgAddAnother, phaseCfgConfirm:
			return m.handleYesNo(msg)
		case phaseCfgServicePortDiscover:
			if msg.String() == "ctrl+c" || msg.String() == "esc" || msg.String() == "q" {
				m.aborted = true
				return m, tea.Quit
			}
			return m, nil
		case phaseCfgPreview:
			return m.handlePreview(msg)
		case phaseCfgDone:
			return m.handleDone(msg)
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
		m.input.Placeholder = ""
	case phaseCfgServiceID:
		if value == "" {
			m.errMsg = "type a service id, then press enter (e.g. backend, ui)"
			return m, nil
		}
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
		m.phase = phaseCfgServicePortDiscover
		m.portDiscoverNote = ""
		return m, m.discoverPortCmd()
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

type portDiscoverMsg struct {
	port string
	err  error
}

func (m configureModel) discoverPortCmd() tea.Cmd {
	workDir := m.workDir
	command := m.currentService.Command
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		port, err := portdiscover.Discover(ctx, workDir, command)
		return portDiscoverMsg{port: port, err: err}
	}
}

func (m configureModel) handlePortDiscover(msg portDiscoverMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.portDiscoverNote = "Could not detect port automatically — enter one manually or leave blank."
	} else if msg.port != "" {
		m.currentService.Port = msg.port
		m.portDiscoverNote = fmt.Sprintf("Detected port %s from command output.", msg.port)
	} else {
		m.portDiscoverNote = "No port found in command output — enter one manually or leave blank."
	}

	m.phase = phaseCfgServicePort
	m.input.SetValue(m.currentService.Port)
	m.input.Placeholder = "${PORT} (optional)"
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
			m.input.Placeholder = ""
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

func (m configureModel) handleDone(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc", "enter", " ":
		return m, tea.Quit
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
	m.phase = phaseCfgDone
	return m, nil
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

	if m.done {
		return m.renderDone()
	}

	var b strings.Builder
	b.WriteString(m.renderConfigureHeader())
	b.WriteString("\n")

	if m.errMsg != "" {
		b.WriteString(errStyle.Render("Error: "+m.errMsg) + "\n\n")
	}

	switch m.phase {
	case phaseCfgWelcome:
		b.WriteString(m.renderWelcome())
	case phaseCfgServiceDeps:
		b.WriteString(fmt.Sprintf("Which services must start before %q?\n\n", m.currentID))
		b.WriteString(mutedStyle.Render("Select dependencies if this service needs others running first (e.g. API before UI).\n\n"))
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
			svc := m.services[id]
			b.WriteString(fmt.Sprintf("%s%s %s (%s)\n", marker, check, svc.Label, id))
		}
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("↑/↓ move  space toggle  enter continue"))
	case phaseCfgAddAnother:
		b.WriteString(fmt.Sprintf("Service saved. You now have %d service(s).\n\n", len(m.services)))
		b.WriteString("Add another service?\n")
		b.WriteString(helpStyle.Render("y yes  n no / enter to finish"))
	case phaseCfgPreview:
		b.WriteString("Review your configuration before saving:\n\n")
		b.WriteString(m.previewYAML())
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("enter to continue  q quit"))
	case phaseCfgConfirm:
		b.WriteString(fmt.Sprintf("Save configuration to %s?\n\n", m.outputPath))
		b.WriteString(mutedStyle.Render("You can edit this file later with `muxdev configure`.\n\n"))
		b.WriteString(helpStyle.Render("y save  n cancel"))
	case phaseCfgServicePortDiscover:
		b.WriteString(fmt.Sprintf("Running %q to detect its port...\n\n", m.currentService.Command))
		b.WriteString(mutedStyle.Render("This starts the dev command briefly and scans logs for localhost URLs.\n"))
		b.WriteString(mutedStyle.Render("It may take a few seconds. Press ctrl+c to cancel.\n"))
	default:
		b.WriteString(m.phasePrompt())
		if hint := m.phaseHint(); hint != "" {
			b.WriteString("\n")
			b.WriteString(mutedStyle.Render(hint))
		}
		if m.portDiscoverNote != "" && m.phase == phaseCfgServicePort {
			b.WriteString("\n")
			b.WriteString(selectedStyle.Render(m.portDiscoverNote))
		}
		b.WriteString("\n\n")
		b.WriteString(m.input.View())
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("enter continue  esc quit"))
	}

	return b.String()
}

func (m configureModel) renderConfigureHeader() string {
	name := "muxdev configure"
	subtitle := "Interactive configuration"
	if m.init {
		name = "muxdev init"
		subtitle = "Set up your project"
	}
	return renderHeader(&config.Config{Name: name, Subtitle: subtitle}, m.width, m.phaseTitle())
}

func (m configureModel) renderWelcome() string {
	var b strings.Builder

	if m.init {
		b.WriteString(renderLogo(m.width))
		b.WriteString("\n\n")
		b.WriteString("Welcome! This wizard creates a ")
		b.WriteString(selectedStyle.Render("muxdev.yaml"))
		b.WriteString(" manifest for your project.\n\n")
		b.WriteString("muxdev runs your local dev services together in one interactive terminal — pick services, stream logs, and shut down cleanly.\n\n")
		b.WriteString("You'll configure:\n")
		b.WriteString("  • Project name and subtitle (shown in the TUI header)\n")
		b.WriteString("  • Dev services — command, port auto-detection, and dependencies\n")
		b.WriteString("  • A preview before anything is written to disk\n\n")
		b.WriteString(fmt.Sprintf("Output file: %s\n\n", m.outputPath))
		b.WriteString(helpStyle.Render("enter to begin  q to quit"))
		return b.String()
	}

	b.WriteString("Create or update your muxdev.yaml step by step.\n\n")
	b.WriteString(fmt.Sprintf("Output file: %s\n\n", m.outputPath))
	b.WriteString(helpStyle.Render("enter to start  q to quit"))
	return b.String()
}

func (m configureModel) renderDone() string {
	var b strings.Builder
	if m.init {
		b.WriteString(renderLogo(m.width))
		b.WriteString("\n\n")
	}
	b.WriteString(m.renderConfigureHeader())
	b.WriteString("\n\n")
	b.WriteString(selectedStyle.Render("✓ Saved "+m.outputPath))
	b.WriteString("\n\n")
	b.WriteString("You're all set! Next steps:\n\n")
	b.WriteString("  ")
	b.WriteString(selectedStyle.Render("muxdev"))
	b.WriteString("          Start the interactive dev stack\n")
	b.WriteString("  ")
	b.WriteString(selectedStyle.Render("muxdev --list"))
	b.WriteString("   List configured services\n")
	b.WriteString("  ")
	b.WriteString(selectedStyle.Render("muxdev configure"))
	b.WriteString("  Edit this setup later\n\n")
	b.WriteString(helpStyle.Render("press any key to exit"))
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
	case phaseCfgServicePortDiscover:
		return "Port discovery"
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
	case phaseCfgDone:
		return "Done"
	default:
		return "Configure"
	}
}

func (m configureModel) phasePrompt() string {
	switch m.phase {
	case phaseCfgName:
		return "What is your project called?"
	case phaseCfgSubtitle:
		return "Add a short subtitle (optional):"
	case phaseCfgServiceID:
		return "Service ID — a short identifier used in CLI flags:"
	case phaseCfgServiceLabel:
		return fmt.Sprintf("Display name for %q:", m.currentID)
	case phaseCfgServiceCommand:
		return fmt.Sprintf("Shell command to start %q:", m.currentID)
	case phaseCfgServicePort:
		return fmt.Sprintf("Port for %q (optional):", m.currentID)
	default:
		return ""
	}
}

func (m configureModel) phaseHint() string {
	switch m.phase {
	case phaseCfgName:
		return "Shown in the TUI header when you run muxdev."
	case phaseCfgSubtitle:
		return "e.g. \"Local development stack\" — leave blank to skip."
	case phaseCfgServiceID:
		return "Lowercase letters, numbers, and hyphens — e.g. backend, ui, worker."
	case phaseCfgServiceLabel:
		return "Human-readable name shown in the service picker."
	case phaseCfgServiceCommand:
		return "e.g. npm run dev, go run ./cmd/api, docker compose up."
	case phaseCfgServicePort:
		return "Confirm or edit the detected port — leave blank to skip."
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
