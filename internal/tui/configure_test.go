package tui

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
)

func TestDepCandidatesExcludesCurrentServiceWhenEditing(t *testing.T) {
	m := configureModel{
		editingExisting: true,
		currentID:       "ui",
		services: map[string]config.Service{
			"backend": {Label: "Backend", Command: "true"},
			"ui":      {Label: "Web UI", Command: "true", DependsOn: []string{"backend"}},
		},
	}

	got := m.depCandidates()
	if len(got) != 1 || got[0] != "backend" {
		t.Fatalf("depCandidates() = %v, want [backend]", got)
	}
}

func TestDepSelectedForEditPrefillsDependencies(t *testing.T) {
	m := configureModel{
		editingExisting: true,
		currentService: config.Service{
			DependsOn: []string{"backend"},
		},
	}

	selected := m.depSelectedForEdit()
	if !selected["backend"] || len(selected) != 1 {
		t.Fatalf("depSelectedForEdit() = %v, want backend selected", selected)
	}
}

func TestDependentsOf(t *testing.T) {
	m := configureModel{
		services: map[string]config.Service{
			"backend": {Label: "Backend", Command: "true"},
			"ui":      {Label: "Web UI", Command: "true", DependsOn: []string{"backend"}},
			"worker":  {Label: "Worker", Command: "true", DependsOn: []string{"backend"}},
		},
	}

	got := m.dependentsOf("backend")
	if len(got) != 2 || got[0] != "ui" || got[1] != "worker" {
		t.Fatalf("dependentsOf() = %v, want [ui worker]", got)
	}
}

func TestCopyService(t *testing.T) {
	original := config.Service{
		Label:     "UI",
		Command:   "npm run dev",
		Port:      "3000",
		DependsOn: []string{"backend"},
		Env:       map[string]string{"FOO": "bar"},
	}

	copy := copyService(original)
	copy.DependsOn[0] = "other"
	copy.Env["FOO"] = "baz"

	if original.DependsOn[0] != "backend" {
		t.Fatalf("original.DependsOn = %v, want [backend]", original.DependsOn)
	}
	if original.Env["FOO"] != "bar" {
		t.Fatalf("original.Env[F00] = %q, want bar", original.Env["FOO"])
	}
}

func TestHandlePortDiscoverOffersConfirmWhenFound(t *testing.T) {
	m := &configureModel{currentID: "ui"}
	_, _ = m.handlePortDiscover(portDiscoverMsg{port: "4321"})

	if m.phase != phaseCfgServicePortConfirm {
		t.Fatalf("phase = %v, want port confirm", m.phase)
	}
	if m.discoveredPort != "4321" {
		t.Fatalf("discoveredPort = %q, want 4321", m.discoveredPort)
	}
}

func TestHandlePortDiscoverFallsBackToManualWhenMissing(t *testing.T) {
	m := newConfigureModel(ConfigureOptions{})
	_, cmd := m.handlePortDiscover(portDiscoverMsg{})

	if m.phase != phaseCfgServicePort {
		t.Fatalf("phase = %v, want manual port entry", m.phase)
	}
	if cmd == nil {
		t.Fatal("expected input focus command")
	}
}

func TestHandlePortConfirmRejectOpensManualEntry(t *testing.T) {
	m := newConfigureModel(ConfigureOptions{})
	m.phase = phaseCfgServicePortConfirm
	m.discoveredPort = "4321"
	m.currentID = "ui"
	_, cmd := m.handlePortConfirm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})

	if m.phase != phaseCfgServicePort {
		t.Fatalf("phase = %v, want manual port entry", m.phase)
	}
	if m.discoveredPort != "" {
		t.Fatalf("discoveredPort = %q, want cleared", m.discoveredPort)
	}
	if cmd == nil {
		t.Fatal("expected input focus command")
	}
}

func TestHandlePortConfirmAcceptContinues(t *testing.T) {
	m := newConfigureModel(ConfigureOptions{})
	m.phase = phaseCfgServicePortConfirm
	m.discoveredPort = "4321"
	m.currentID = "ui"
	_, _ = m.handlePortConfirm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})

	svc, ok := m.services["ui"]
	if !ok || svc.Port != "4321" {
		t.Fatalf("saved port = %q, want 4321", svc.Port)
	}
	if m.phase != phaseCfgAddAnother {
		t.Fatalf("phase = %v, want add another after finalize", m.phase)
	}
}

func TestSortedSelectedDeps(t *testing.T) {
	got := sortedSelectedDeps(map[string]bool{
		"ui":      true,
		"backend": true,
		"worker":  false,
	})
	if len(got) != 2 || got[0] != "backend" || got[1] != "ui" {
		t.Fatalf("sortedSelectedDeps() = %v, want [backend ui]", got)
	}
}

func TestServiceEditValueDisplay(t *testing.T) {
	m := configureModel{
		currentID: "ui",
		services: map[string]config.Service{
			"ui": {
				Label:     "Web UI",
				Command:   "npm run dev",
				Port:      "3000",
				DependsOn: []string{"backend"},
			},
		},
	}

	if got := m.serviceEditValueDisplay(serviceEditDeps); got != "backend" {
		t.Fatalf("serviceEditValueDisplay(deps) = %q, want backend", got)
	}
	if got := m.serviceEditValueDisplay(serviceEditPort); got != "3000" {
		t.Fatalf("serviceEditValueDisplay(port) = %q, want 3000", got)
	}
}

func TestRootMenuValueDisplay(t *testing.T) {
	m := configureModel{
		name:     "My App",
		subtitle: "Local stack",
		services: map[string]config.Service{
			"api": {Label: "API", Command: "true"},
		},
	}

	if got := m.rootMenuValueDisplay(rootEditServices); got != "1 service(s)" {
		t.Fatalf("rootMenuValueDisplay(services) = %q", got)
	}
	if got := m.rootMenuValueDisplay(rootEditName); got != "My App" {
		t.Fatalf("rootMenuValueDisplay(name) = %q", got)
	}
	if got := m.rootMenuValueDisplay(rootEditAll); got != "name, subtitle, and every service" {
		t.Fatalf("rootMenuValueDisplay(all) = %q", got)
	}
}

func TestApplyPartialField(t *testing.T) {
	m := configureModel{
		currentID: "ui",
		services: map[string]config.Service{
			"ui": {Label: "Web UI", Command: "npm run dev", Port: "3000"},
		},
	}

	m.applyPartialField(serviceEditLabel, "Frontend")
	if m.services["ui"].Label != "Frontend" {
		t.Fatalf("Label = %q, want Frontend", m.services["ui"].Label)
	}
	if m.services["ui"].Command != "npm run dev" {
		t.Fatalf("Command changed unexpectedly: %q", m.services["ui"].Command)
	}
}

func TestNewConfigureModelPrefillsProjectName(t *testing.T) {
	m := newConfigureModel(ConfigureOptions{Init: true, WorkDir: "/home/user/my-cool-app"})
	if m.name != "My Cool App" {
		t.Fatalf("name = %q, want My Cool App", m.name)
	}
}

func TestNewConfigureModelFromRunStartsAtSetupPrompt(t *testing.T) {
	m := newConfigureModel(ConfigureOptions{Init: true, FromRun: true})
	if m.phase != phaseCfgSetupPrompt {
		t.Fatalf("phase = %v, want setup prompt", m.phase)
	}
}

func TestSaveAndQuitFromRunExitsImmediately(t *testing.T) {
	m := newConfigureModel(ConfigureOptions{Init: true, FromRun: true})
	m.name = "Demo"
	m.services = map[string]config.Service{
		"app": {Label: "App", Command: "true"},
	}
	dir := t.TempDir()
	m.outputPath = filepath.Join(dir, "muxdev.yaml")
	m.workDir = dir

	_, cmd := m.saveAndQuit()
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if !m.done {
		t.Fatal("expected done=true")
	}
	if m.phase == phaseCfgDone {
		t.Fatal("fromRun should skip done screen")
	}
	if !config.Exists(m.outputPath) {
		t.Fatal("expected config file to be saved")
	}
}

func TestBeginProjectSetupPrefillsName(t *testing.T) {
	m := newConfigureModel(ConfigureOptions{Init: true, WorkDir: "/tmp/demo-app"})
	cmd := m.beginProjectSetup()
	if cmd == nil {
		t.Fatal("expected input focus command")
	}
	if m.input.Value() != "Demo App" {
		t.Fatalf("input value = %q, want Demo App", m.input.Value())
	}
}
