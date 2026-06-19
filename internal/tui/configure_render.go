package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	initAccentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
	initCalloutStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.Color("81")).
				Foreground(lipgloss.Color("252")).
				PaddingLeft(1)
	initCodeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("236")).Padding(0, 1)
	initSuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	initPathStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("236")).Padding(0, 1)
)

func initPanel(width int, body string) string {
	return cardStyle.Width(min(width-2, 72)).Render(body)
}

func initStepForPhase(phase configurePhase) (step int, label string) {
	switch phase {
	case phaseCfgWelcome, phaseCfgSetupPrompt:
		return 1, "Start"
	case phaseCfgName, phaseCfgSubtitle:
		return 2, "Project"
	case phaseCfgServiceID, phaseCfgServiceLabel, phaseCfgServiceCommand,
		phaseCfgServicePortDiscover, phaseCfgServicePortConfirm, phaseCfgServicePort,
		phaseCfgServiceDeps, phaseCfgAddAnother:
		return 3, "Services"
	case phaseCfgPreview, phaseCfgConfirm:
		return 4, "Review"
	case phaseCfgDone:
		return 4, "Done"
	default:
		return 0, ""
	}
}

func renderInitProgress(phase configurePhase) string {
	current, _ := initStepForPhase(phase)
	if current == 0 {
		return ""
	}
	steps := []string{"Start", "Project", "Services", "Review"}
	var parts []string
	for i, name := range steps {
		step := i + 1
		switch {
		case step < current:
			parts = append(parts, initAccentStyle.Render("✓ "+name))
		case step == current:
			parts = append(parts, selectedStyle.Render("› "+name))
		default:
			parts = append(parts, mutedStyle.Render(name))
		}
	}
	return strings.Join(parts, mutedStyle.Render("  ·  "))
}

func renderInitChecklist(items []string) string {
	var b strings.Builder
	for _, item := range items {
		b.WriteString(initAccentStyle.Render("  ✦ "))
		b.WriteString(item)
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderInitDetectionCallout(name, command, outputPath string) string {
	if name == "" && command == "" {
		return ""
	}
	var lines []string
	if name != "" {
		lines = append(lines, "Suggested name: "+selectedStyle.Render(name))
	}
	if command != "" {
		lines = append(lines, "Detected command: "+initCodeStyle.Render(command))
	}
	if outputPath != "" {
		lines = append(lines, "Output file: "+initPathStyle.Render(outputPath))
	}
	return initCalloutStyle.Render(strings.Join(lines, "\n"))
}

func renderInitNextSteps() string {
	rows := []struct {
		cmd  string
		desc string
	}{
		{"muxdev", "Start the interactive dev stack"},
		{"muxdev list", "List configured services"},
		{"muxdev configure", "Edit this setup later"},
	}
	var b strings.Builder
	for _, row := range rows {
		b.WriteString("  ")
		b.WriteString(initCodeStyle.Render(row.cmd))
		b.WriteString(mutedStyle.Render("  " + row.desc))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderInitFieldPanel(width int, prompt, hint, note, input string) string {
	var body strings.Builder
	body.WriteString(titleStyle.Render(prompt))
	if hint != "" {
		body.WriteString("\n")
		body.WriteString(mutedStyle.Render(hint))
	}
	if note != "" {
		body.WriteString("\n")
		body.WriteString(selectedStyle.Render(note))
	}
	body.WriteString("\n\n")
	body.WriteString(input)
	return initPanel(width, body.String())
}

func renderInitYAMLPreview(width int, yaml string) string {
	return initPanel(width, initCodeStyle.Render(yaml))
}

func renderInitYesNoPanel(width int, title, subtitle, help string) string {
	var body strings.Builder
	body.WriteString(titleStyle.Render(title))
	if subtitle != "" {
		body.WriteString("\n\n")
		body.WriteString(mutedStyle.Render(subtitle))
	}
	return initPanel(width, body.String()) + "\n\n" + helpStyle.Render(help)
}

func (m configureModel) showConfigureHeader() bool {
	if m.done {
		return false
	}
	if m.init && (m.phase == phaseCfgWelcome || m.phase == phaseCfgSetupPrompt) {
		return false
	}
	return true
}

func (m configureModel) appendInitProgress(b *strings.Builder) {
	if !m.init || m.edit {
		return
	}
	if progress := renderInitProgress(m.phase); progress != "" {
		b.WriteString(progress)
		b.WriteString("\n\n")
	}
}
