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

type serviceEditChoice int

const (
	serviceEditLabel serviceEditChoice = iota
	serviceEditCommand
	serviceEditPort
	serviceEditDeps
	serviceEditAll
	serviceEditBack
)

type serviceEditOption struct {
	title  string
	choice serviceEditChoice
}

func serviceEditMenuOptions() []serviceEditOption {
	return []serviceEditOption{
		{title: "Label", choice: serviceEditLabel},
		{title: "Command", choice: serviceEditCommand},
		{title: "Port", choice: serviceEditPort},
		{title: "Dependencies", choice: serviceEditDeps},
		{title: "All service fields", choice: serviceEditAll},
		{title: "Back to services", choice: serviceEditBack},
	}
}

type rootEditChoice int

const (
	rootEditServices rootEditChoice = iota
	rootEditName
	rootEditSubtitle
	rootEditAll
	rootEditFinish
)

type rootEditOption struct {
	title  string
	choice rootEditChoice
}

func rootMenuOptions() []rootEditOption {
	return []rootEditOption{
		{title: "Services", choice: rootEditServices},
		{title: "Project name", choice: rootEditName},
		{title: "Subtitle", choice: rootEditSubtitle},
		{title: "All project fields", choice: rootEditAll},
		{title: "Save & finish", choice: rootEditFinish},
	}
}

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
	phaseCfgRootMenu
	phaseCfgName
	phaseCfgSubtitle
	phaseCfgServiceID
	phaseCfgServiceLabel
	phaseCfgServiceCommand
	phaseCfgServicePortDiscover
	phaseCfgServicePort
	phaseCfgServiceDeps
	phaseCfgServiceMenu
	phaseCfgServiceEditMenu
	phaseCfgDeleteConfirm
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
	edit       bool
	workDir    string
	errMsg     string
	done       bool
	aborted    bool

	portDiscoverNote string

	name     string
	subtitle string
	services map[string]config.Service

	currentID         string
	currentService    config.Service
	depCursor         int
	depSelected       map[string]bool
	serviceMenuCursor int
	serviceEditCursor int
	rootMenuCursor    int
	editingExisting   bool
	fromServiceMenu   bool
	fromServiceEditMenu bool
	partialEdit       bool
	partialProjectEdit bool
	projectEditAll     bool
	projectEditAllIdx  int
	partialEditField  serviceEditChoice
	pendingDeleteID   string
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
		model.fromServiceMenu = true
		model.enterMenuPhase(phaseCfgRootMenu)
	}

	p := tea.NewProgram(&model, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return err
	}
	m := final.(*configureModel)
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
	ti.CharLimit = 256
	ti.Width = 50

	return configureModel{
		phase:       phaseCfgWelcome,
		input:       ti,
		outputPath:  opts.OutputPath,
		force:       opts.Force,
		init:        opts.Init,
		edit:        opts.Edit,
		workDir:     opts.WorkDir,
		services:    make(map[string]config.Service),
		depSelected: make(map[string]bool),
	}
}

func (m *configureModel) Init() tea.Cmd {
	if configureInputPhase(m.phase) {
		return textinput.Blink
	}
	return nil
}

func (m *configureModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case portDiscoverMsg:
		return m.handlePortDiscover(msg)
	case tea.KeyMsg:
		switch m.phase {
		case phaseCfgWelcome:
			return m.handleWelcome(msg)
		case phaseCfgRootMenu:
			return m.handleRootMenu(msg)
		case phaseCfgServiceMenu:
			return m.handleServiceMenu(msg)
		case phaseCfgServiceEditMenu:
			return m.handleServiceEditMenu(msg)
		case phaseCfgDeleteConfirm:
			return m.handleDeleteConfirm(msg)
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

func (m *configureModel) handleWelcome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case configureKeyEnter(msg):
		m.input.SetValue(m.name)
		m.input.Placeholder = "My App"
		return m, m.enterInputPhase(phaseCfgName)
	case configureKeyQuit(msg), configureKeyBack(msg):
		m.aborted = true
		return m, tea.Quit
	}
	return m, nil
}

func (m *configureModel) handleTextInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.aborted = true
		return m, tea.Quit
	case "esc":
		if m.partialProjectEdit {
			return m.returnToRootMenu()
		}
		if m.projectEditAll {
			m.projectEditAll = false
			m.projectEditAllIdx = 0
			m.fromServiceEditMenu = false
			m.editingExisting = false
			return m.returnToRootMenu()
		}
		if m.partialEdit {
			return m.returnToServiceEditMenu()
		}
		if m.fromServiceEditMenu && m.editingExisting && !m.partialEdit {
			return m.returnToServiceEditMenu()
		}
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
		if m.partialProjectEdit {
			return m.returnToRootMenu()
		}
		if m.projectEditAll {
			m.phase = phaseCfgSubtitle
			m.input.SetValue(m.subtitle)
			m.input.Placeholder = "Local development stack (optional)"
			m.input.Focus()
			return m, textinput.Blink
		}
		m.phase = phaseCfgSubtitle
		m.input.SetValue(m.subtitle)
		m.input.Placeholder = "Local development stack (optional)"
	case phaseCfgSubtitle:
		m.subtitle = value
		if m.partialProjectEdit {
			return m.returnToRootMenu()
		}
		if m.projectEditAll {
			return m.beginProjectEditAllNextService()
		}
		if m.edit && len(m.services) > 0 {
			m.fromServiceMenu = true
			m.phase = phaseCfgServiceMenu
			m.serviceMenuCursor = 0
			return m, nil
		}
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
		if m.partialEdit {
			m.applyPartialField(serviceEditLabel, value)
			return m.returnToServiceEditMenu()
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
		if m.partialEdit {
			m.applyPartialField(serviceEditCommand, value)
			return m.returnToServiceEditMenu()
		}
		m.currentService.Command = value
		if m.editingExisting && m.fromServiceEditMenu && !m.partialEdit {
			m.phase = phaseCfgServicePort
			m.input.SetValue(m.currentService.Port)
			m.input.Placeholder = "${PORT} (optional)"
			m.input.Focus()
			return m, textinput.Blink
		}
		m.phase = phaseCfgServicePortDiscover
		m.portDiscoverNote = ""
		return m, m.discoverPortCmd()
	case phaseCfgServicePort:
		if m.partialEdit {
			m.applyPartialField(serviceEditPort, value)
			return m.returnToServiceEditMenu()
		}
		m.currentService.Port = value
		if len(m.depCandidates()) == 0 {
			m.finalizeCurrentService()
			return m.returnAfterServiceSave()
		}
		m.depCursor = 0
		m.depSelected = m.depSelectedForEdit()
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

func (m *configureModel) handlePortDiscover(msg portDiscoverMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.portDiscoverNote = "Could not detect port automatically — enter one manually or leave blank."
	} else if msg.port != "" {
		m.currentService.Port = msg.port
		m.portDiscoverNote = fmt.Sprintf("Detected port %s from command output.", msg.port)
	} else {
		m.portDiscoverNote = "No port found in command output — enter one manually or leave blank."
	}

	m.input.SetValue(m.currentService.Port)
	m.input.Placeholder = "${PORT} (optional)"
	return m, m.enterInputPhase(phaseCfgServicePort)
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
	if !m.fromServiceEditMenu {
		m.currentID = ""
	}
	m.currentService = config.Service{}
}

func (m *configureModel) handleDeps(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ids := m.depCandidates()
	switch {
	case configureKeyQuit(msg):
		m.aborted = true
		return m, tea.Quit
	case configureKeyBack(msg):
		if m.projectEditAll {
			m.projectEditAll = false
			m.projectEditAllIdx = 0
			m.fromServiceEditMenu = false
			m.editingExisting = false
			return m.returnToRootMenu()
		}
		if m.partialEdit || (m.fromServiceEditMenu && m.editingExisting && !m.partialEdit) {
			return m.returnToServiceEditMenu()
		}
		if m.edit {
			m.enterMenuPhase(phaseCfgRootMenu)
			return m, nil
		}
		m.aborted = true
		return m, tea.Quit
	case configureKeyUp(msg):
		if m.depCursor > 0 {
			m.depCursor--
		}
	case configureKeyDown(msg):
		if len(ids) > 0 && m.depCursor < len(ids)-1 {
			m.depCursor++
		}
	case configureKeySpace(msg):
		if len(ids) > 0 {
			id := ids[m.depCursor]
			m.depSelected[id] = !m.depSelected[id]
		}
	case configureKeyEnter(msg):
		if m.partialEdit {
			m.applyPartialDeps()
			return m.returnToServiceEditMenu()
		}
		m.finalizeCurrentService()
		return m.returnAfterServiceSave()
	}
	return m, nil
}

func (m *configureModel) handleRootMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	options := rootMenuOptions()
	switch {
	case configureKeyQuit(msg):
		m.aborted = true
		return m, tea.Quit
	case configureKeyUp(msg):
		if m.rootMenuCursor > 0 {
			m.rootMenuCursor--
		}
	case configureKeyDown(msg):
		if m.rootMenuCursor < len(options)-1 {
			m.rootMenuCursor++
		}
	case configureKeyEnter(msg):
		switch options[m.rootMenuCursor].choice {
		case rootEditServices:
			m.serviceMenuCursor = 0
			m.enterMenuPhase(phaseCfgServiceMenu)
		case rootEditName:
			m.partialProjectEdit = true
			m.input.SetValue(m.name)
			m.input.Placeholder = "My App"
			return m, m.enterInputPhase(phaseCfgName)
		case rootEditSubtitle:
			m.partialProjectEdit = true
			m.input.SetValue(m.subtitle)
			m.input.Placeholder = "Local development stack (optional)"
			return m, m.enterInputPhase(phaseCfgSubtitle)
		case rootEditAll:
			if len(m.services) == 0 {
				m.errMsg = "add at least one service"
				return m, nil
			}
			m.projectEditAll = true
			m.projectEditAllIdx = 0
			m.input.SetValue(m.name)
			m.input.Placeholder = "My App"
			return m, m.enterInputPhase(phaseCfgName)
		case rootEditFinish:
			if len(m.services) == 0 {
				m.errMsg = "add at least one service"
				return m, nil
			}
			m.enterMenuPhase(phaseCfgPreview)
		}
	}
	return m, nil
}

func (m *configureModel) returnToRootMenu() (tea.Model, tea.Cmd) {
	m.partialProjectEdit = false
	m.projectEditAll = false
	m.projectEditAllIdx = 0
	m.fromServiceEditMenu = false
	m.editingExisting = false
	m.partialEdit = false
	m.errMsg = ""
	m.enterMenuPhase(phaseCfgRootMenu)
	return m, nil
}

func (m *configureModel) beginProjectEditAllNextService() (tea.Model, tea.Cmd) {
	ids := m.sortedServiceIDs()
	if m.projectEditAllIdx >= len(ids) {
		m.projectEditAll = false
		m.projectEditAllIdx = 0
		m.fromServiceEditMenu = false
		m.editingExisting = false
		return m.returnToRootMenu()
	}

	id := ids[m.projectEditAllIdx]
	svc := m.services[id]
	m.currentID = id
	m.currentService = copyService(svc)
	m.editingExisting = true
	m.fromServiceEditMenu = true
	m.partialEdit = false
	m.phase = phaseCfgServiceLabel
	m.input.SetValue(svc.Label)
	m.input.Placeholder = "Backend"
	m.input.Focus()
	return m, textinput.Blink
}

func (m configureModel) rootMenuValueDisplay(choice rootEditChoice) string {
	switch choice {
	case rootEditServices:
		return fmt.Sprintf("%d service(s)", len(m.services))
	case rootEditName:
		return m.name
	case rootEditSubtitle:
		return emptyDisplay(m.subtitle)
	case rootEditAll:
		return "name, subtitle, and every service"
	default:
		return ""
	}
}

func (m *configureModel) handleServiceEditMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	options := serviceEditMenuOptions()
	switch {
	case configureKeyQuit(msg):
		m.aborted = true
		return m, tea.Quit
	case configureKeyBack(msg):
		return m.leaveServiceEditMenu()
	case configureKeyUp(msg):
		if m.serviceEditCursor > 0 {
			m.serviceEditCursor--
		}
	case configureKeyDown(msg):
		if m.serviceEditCursor < len(options)-1 {
			m.serviceEditCursor++
		}
	case configureKeyEnter(msg):
		choice := options[m.serviceEditCursor].choice
		if choice == serviceEditBack {
			return m.leaveServiceEditMenu()
		}
		return m.beginServiceEdit(choice)
	}
	return m, nil
}

func (m *configureModel) handleServiceMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ids := m.sortedServiceIDs()
	switch {
	case configureKeyQuit(msg):
		m.aborted = true
		return m, tea.Quit
	case configureKeyBack(msg):
		if m.edit {
			m.enterMenuPhase(phaseCfgRootMenu)
			return m, nil
		}
		m.aborted = true
		return m, tea.Quit
	case configureKeyUp(msg):
		if m.serviceMenuCursor > 0 {
			m.serviceMenuCursor--
		}
	case configureKeyDown(msg):
		if len(ids) > 0 && m.serviceMenuCursor < len(ids)-1 {
			m.serviceMenuCursor++
		}
	case msg.String() == "a", msg.String() == "A":
		return m.beginAddService()
	case msg.String() == "f", msg.String() == "F":
		if len(m.services) == 0 {
			m.errMsg = "add at least one service"
			return m, nil
		}
		if m.edit {
			m.enterMenuPhase(phaseCfgRootMenu)
			return m, nil
		}
		m.enterMenuPhase(phaseCfgPreview)
		return m, nil
	case msg.String() == "d", msg.String() == "D":
		if len(ids) == 0 {
			return m, nil
		}
		m.pendingDeleteID = ids[m.serviceMenuCursor]
		m.enterMenuPhase(phaseCfgDeleteConfirm)
		return m, nil
	case configureKeyEnter(msg):
		if len(ids) == 0 {
			return m.beginAddService()
		}
		return m.openServiceEditMenu(ids[m.serviceMenuCursor])
	}
	return m, nil
}

func (m *configureModel) handleDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc", "q":
		m.aborted = true
		return m, tea.Quit
	case "y", "Y":
		delete(m.services, m.pendingDeleteID)
		m.pendingDeleteID = ""
		ids := m.sortedServiceIDs()
		if m.serviceMenuCursor >= len(ids) && len(ids) > 0 {
			m.serviceMenuCursor = len(ids) - 1
		}
		m.enterMenuPhase(phaseCfgServiceMenu)
		return m, nil
	case "n", "N", "enter":
		m.pendingDeleteID = ""
		m.enterMenuPhase(phaseCfgServiceMenu)
		return m, nil
	}
	return m, nil
}

func (m *configureModel) openServiceEditMenu(id string) (tea.Model, tea.Cmd) {
	m.currentID = id
	m.editingExisting = true
	m.fromServiceEditMenu = true
	m.partialEdit = false
	m.serviceEditCursor = 0
	m.errMsg = ""
	m.enterMenuPhase(phaseCfgServiceEditMenu)
	return m, nil
}

func (m *configureModel) leaveServiceEditMenu() (tea.Model, tea.Cmd) {
	m.currentID = ""
	m.currentService = config.Service{}
	m.editingExisting = false
	m.fromServiceEditMenu = false
	m.partialEdit = false
	m.enterMenuPhase(phaseCfgServiceMenu)
	return m, nil
}

func (m *configureModel) returnToServiceEditMenu() (tea.Model, tea.Cmd) {
	m.partialEdit = false
	m.editingExisting = true
	m.currentService = config.Service{}
	m.errMsg = ""
	m.enterMenuPhase(phaseCfgServiceEditMenu)
	return m, nil
}

func (m *configureModel) beginServiceEdit(choice serviceEditChoice) (tea.Model, tea.Cmd) {
	svc := m.services[m.currentID]
	m.errMsg = ""

	if choice == serviceEditAll {
		m.partialEdit = false
		m.currentService = copyService(svc)
		m.phase = phaseCfgServiceLabel
		m.input.SetValue(svc.Label)
		m.input.Placeholder = "Backend"
		m.input.Focus()
		return m, textinput.Blink
	}

	m.partialEdit = true
	m.partialEditField = choice

	switch choice {
	case serviceEditLabel:
		m.phase = phaseCfgServiceLabel
		m.input.SetValue(svc.Label)
		m.input.Placeholder = "Backend"
	case serviceEditCommand:
		m.phase = phaseCfgServiceCommand
		m.input.SetValue(svc.Command)
		m.input.Placeholder = "npm run dev"
	case serviceEditPort:
		m.phase = phaseCfgServicePort
		m.input.SetValue(svc.Port)
		m.input.Placeholder = "${PORT} (optional)"
	case serviceEditDeps:
		m.depCursor = 0
		m.depSelected = m.depSelectedFromCurrentService()
		m.phase = phaseCfgServiceDeps
		return m, nil
	}

	m.input.Focus()
	return m, textinput.Blink
}

func (m *configureModel) applyPartialField(choice serviceEditChoice, value string) {
	svc := copyService(m.services[m.currentID])
	switch choice {
	case serviceEditLabel:
		svc.Label = value
	case serviceEditCommand:
		svc.Command = value
	case serviceEditPort:
		svc.Port = value
	}
	m.services[m.currentID] = svc
}

func (m *configureModel) applyPartialDeps() {
	svc := copyService(m.services[m.currentID])
	svc.DependsOn = sortedSelectedDeps(m.depSelected)
	m.services[m.currentID] = svc
}

func sortedSelectedDeps(selected map[string]bool) []string {
	deps := make([]string, 0)
	for id, on := range selected {
		if on {
			deps = append(deps, id)
		}
	}
	for i := 0; i < len(deps); i++ {
		for j := i + 1; j < len(deps); j++ {
			if deps[j] < deps[i] {
				deps[i], deps[j] = deps[j], deps[i]
			}
		}
	}
	return deps
}

func (m configureModel) depSelectedFromCurrentService() map[string]bool {
	selected := make(map[string]bool)
	for _, dep := range m.services[m.currentID].DependsOn {
		selected[dep] = true
	}
	return selected
}

func (m configureModel) serviceEditValueDisplay(choice serviceEditChoice) string {
	svc := m.services[m.currentID]
	switch choice {
	case serviceEditLabel:
		return svc.Label
	case serviceEditCommand:
		return truncateDisplay(svc.Command, 48)
	case serviceEditPort:
		return emptyDisplay(svc.Port)
	case serviceEditDeps:
		return emptyDisplay(strings.Join(svc.DependsOn, ", "))
	case serviceEditAll:
		return "walk through every service field"
	case serviceEditBack:
		return ""
	default:
		return ""
	}
}

func emptyDisplay(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}

func truncateDisplay(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max-1] + "…"
}

func (m *configureModel) beginAddService() (tea.Model, tea.Cmd) {
	m.editingExisting = false
	m.currentID = ""
	m.currentService = config.Service{DependsOn: []string{}}
	m.errMsg = ""
	m.phase = phaseCfgServiceID
	m.input.SetValue("")
	m.input.Placeholder = ""
	m.input.Focus()
	return m, textinput.Blink
}

func (m *configureModel) returnAfterServiceSave() (tea.Model, tea.Cmd) {
	m.editingExisting = false
	m.partialEdit = false
	if m.projectEditAll {
		m.projectEditAllIdx++
		return m.beginProjectEditAllNextService()
	}
	if m.fromServiceEditMenu {
		m.currentService = config.Service{}
		m.enterMenuPhase(phaseCfgServiceEditMenu)
		return m, nil
	}
	m.currentID = ""
	m.currentService = config.Service{}
	if m.fromServiceMenu {
		m.enterMenuPhase(phaseCfgServiceMenu)
		return m, nil
	}
	m.enterMenuPhase(phaseCfgAddAnother)
	return m, nil
}

func (m configureModel) depCandidates() []string {
	ids := m.sortedServiceIDs()
	if !m.editingExisting {
		return ids
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != m.currentID {
			out = append(out, id)
		}
	}
	return out
}

func (m configureModel) depSelectedForEdit() map[string]bool {
	selected := make(map[string]bool)
	if !m.editingExisting {
		return selected
	}
	for _, dep := range m.currentService.DependsOn {
		selected[dep] = true
	}
	return selected
}

func (m configureModel) dependentsOf(id string) []string {
	var out []string
	for sid, svc := range m.services {
		for _, dep := range svc.DependsOn {
			if dep == id {
				out = append(out, sid)
				break
			}
		}
	}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func copyService(svc config.Service) config.Service {
	copy := svc
	if svc.DependsOn != nil {
		copy.DependsOn = append([]string(nil), svc.DependsOn...)
	}
	if svc.Env != nil {
		copy.Env = make(map[string]string, len(svc.Env))
		for k, v := range svc.Env {
			copy.Env[k] = v
		}
	}
	return copy
}

func (m *configureModel) handleYesNo(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc", "q":
		m.aborted = true
		return m, tea.Quit
	case "y", "Y":
		if m.phase == phaseCfgAddAnother {
			m.input.SetValue("")
			m.input.Placeholder = ""
			return m, m.enterInputPhase(phaseCfgServiceID)
		}
		return m.saveAndQuit()
	case "n", "N":
		if m.phase == phaseCfgAddAnother {
			if len(m.services) == 0 {
				m.errMsg = "add at least one service"
				return m, nil
			}
			m.enterMenuPhase(phaseCfgPreview)
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
			m.enterMenuPhase(phaseCfgPreview)
			return m, nil
		}
	}
	return m, nil
}

func (m *configureModel) handlePreview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case configureKeyQuit(msg):
		m.aborted = true
		return m, tea.Quit
	case configureKeyBack(msg):
		if m.edit {
			m.enterMenuPhase(phaseCfgRootMenu)
			return m, nil
		}
		m.aborted = true
		return m, tea.Quit
	case configureKeyEnter(msg):
		m.enterMenuPhase(phaseCfgConfirm)
		return m, nil
	}
	return m, nil
}

func (m *configureModel) handleDone(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case phaseCfgRootMenu:
		b.WriteString("What do you want to edit?\n\n")
		options := rootMenuOptions()
		for i, opt := range options {
			marker := "  "
			if i == m.rootMenuCursor {
				marker = cursorStyle.Render("> ")
			}
			line := fmt.Sprintf("%s%s", marker, titleStyle.Render(opt.title))
			if value := m.rootMenuValueDisplay(opt.choice); value != "" {
				line += mutedStyle.Render("  " + value)
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("↑/↓ move  enter select  q quit"))
	case phaseCfgServiceDeps:
		b.WriteString(fmt.Sprintf("Which services must start before %q?\n\n", m.currentID))
		b.WriteString(mutedStyle.Render("Select dependencies if this service needs others running first (e.g. API before UI).\n\n"))
		ids := m.depCandidates()
		if len(ids) == 0 {
			b.WriteString(mutedStyle.Render("No other services available.\n\n"))
		}
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
		if m.partialEdit {
			b.WriteString(helpStyle.Render("↑/↓ move  space toggle  enter save  esc back"))
		} else if m.projectEditAll {
			b.WriteString(helpStyle.Render("↑/↓ move  space toggle  enter continue  esc cancel"))
		} else if m.fromServiceEditMenu && m.editingExisting {
			b.WriteString(helpStyle.Render("↑/↓ move  space toggle  enter continue  esc back"))
		} else {
			b.WriteString(helpStyle.Render("↑/↓ move  space toggle  enter continue"))
		}
	case phaseCfgServiceMenu:
		b.WriteString("Manage services:\n\n")
		ids := m.sortedServiceIDs()
		if len(ids) == 0 {
			b.WriteString(mutedStyle.Render("No services yet.\n\n"))
		}
		for i, id := range ids {
			marker := "  "
			if i == m.serviceMenuCursor {
				marker = cursorStyle.Render("> ")
			}
			svc := m.services[id]
			line := fmt.Sprintf("%s%s (%s)", marker, selectedStyle.Render(svc.Label), id)
			if len(svc.DependsOn) > 0 {
				line += mutedStyle.Render("  depends: " + strings.Join(svc.DependsOn, ", "))
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("↑/↓ move  enter open  a add  d delete  esc back"))
	case phaseCfgServiceEditMenu:
		svc := m.services[m.currentID]
		b.WriteString(fmt.Sprintf("Edit %s (%s) — pick a field:\n\n", selectedStyle.Render(m.currentID), svc.Label))
		options := serviceEditMenuOptions()
		for i, opt := range options {
			marker := "  "
			if i == m.serviceEditCursor {
				marker = cursorStyle.Render("> ")
			}
			line := fmt.Sprintf("%s%s", marker, titleStyle.Render(opt.title))
			if value := m.serviceEditValueDisplay(opt.choice); value != "" {
				line += mutedStyle.Render("  " + value)
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("↑/↓ move  enter select  esc back"))
	case phaseCfgDeleteConfirm:
		svc := m.services[m.pendingDeleteID]
		b.WriteString(fmt.Sprintf("Delete service %q (%s)?\n\n", m.pendingDeleteID, svc.Label))
		if dependents := m.dependentsOf(m.pendingDeleteID); len(dependents) > 0 {
			b.WriteString(errStyle.Render("Warning: also used by: " + strings.Join(dependents, ", ") + "\n\n"))
		}
		b.WriteString(helpStyle.Render("y delete  n cancel"))
	case phaseCfgAddAnother:
		b.WriteString(fmt.Sprintf("Service saved. You now have %d service(s).\n\n", len(m.services)))
		b.WriteString("Add another service?\n")
		b.WriteString(helpStyle.Render("y yes  n no / enter to finish"))
	case phaseCfgPreview:
		b.WriteString("Review your configuration before saving:\n\n")
		b.WriteString(m.previewYAML())
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("enter to continue  esc back  q quit"))
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
		if m.partialEdit || m.partialProjectEdit {
			b.WriteString(helpStyle.Render("enter save  esc back"))
		} else if m.projectEditAll {
			b.WriteString(helpStyle.Render("enter continue  esc cancel"))
		} else if m.fromServiceEditMenu && m.editingExisting {
			b.WriteString(helpStyle.Render("enter continue  esc back"))
		} else {
			b.WriteString(helpStyle.Render("enter continue  esc quit"))
		}
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
	if len(m.services) > 0 {
		b.WriteString(fmt.Sprintf("Loaded %d existing service(s). Pick a service, then choose which field to update.\n\n", len(m.services)))
	}
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
	b.WriteString(selectedStyle.Render("muxdev list"))
	b.WriteString("   List configured services\n")
	b.WriteString("  ")
	b.WriteString(selectedStyle.Render("muxdev configure"))
	b.WriteString("  Edit this setup later\n\n")
	b.WriteString(helpStyle.Render("press any key to exit"))
	return b.String()
}

func (m configureModel) phaseTitle() string {
	if m.projectEditAll && m.currentID != "" {
		ids := m.sortedServiceIDs()
		return fmt.Sprintf("All project fields · %d/%d · %s", m.projectEditAllIdx+1, len(ids), m.currentID)
	}
	if m.projectEditAll {
		return "All project fields"
	}

	switch m.phase {
	case phaseCfgWelcome:
		return "Welcome"
	case phaseCfgRootMenu:
		return "Edit config"
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
	case phaseCfgServiceMenu:
		return "Services"
	case phaseCfgServiceEditMenu:
		return "Edit service"
	case phaseCfgDeleteConfirm:
		return "Delete service"
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
		if m.partialEdit || (m.fromServiceEditMenu && m.editingExisting && !m.partialEdit) {
			return fmt.Sprintf("Edit display name for %q:", m.currentID)
		}
		return fmt.Sprintf("Display name for %q:", m.currentID)
	case phaseCfgServiceCommand:
		if m.partialEdit || (m.fromServiceEditMenu && m.editingExisting && !m.partialEdit) {
			return fmt.Sprintf("Edit shell command for %q:", m.currentID)
		}
		return fmt.Sprintf("Shell command to start %q:", m.currentID)
	case phaseCfgServicePort:
		if m.partialEdit || (m.fromServiceEditMenu && m.editingExisting && !m.partialEdit) {
			return fmt.Sprintf("Edit port for %q (optional):", m.currentID)
		}
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
