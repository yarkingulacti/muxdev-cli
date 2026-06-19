package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
)

func RenderServiceList(cfg *config.Config, workDir string, width int) string {
	ids, err := cfg.SortedServiceIDs()
	if err != nil {
		return errStyle.Render(fmt.Sprintf("Error: %v", err))
	}

	if width <= 0 {
		width = 80
	}

	var b strings.Builder
	status := fmt.Sprintf("%d service%s", len(ids), pluralSuffix(len(ids)))
	b.WriteString(renderHeader(cfg, width, status))
	b.WriteString("\n\n")
	b.WriteString(renderServiceTable(cfg, ids, workDir, width))
	return b.String()
}

func renderServiceTable(cfg *config.Config, ids []string, workDir string, width int) string {
	rows := make([][]string, 0, len(ids))
	for _, id := range ids {
		svc := cfg.Services[id]
		resolved := config.ResolveServicePort(cfg, workDir, svc)
		rows = append(rows, []string{
			id,
			svc.Label,
			emptyDash(resolved.Port),
			emptyDash(resolved.Source),
			emptyDash(strings.Join(svc.DependsOn, ", ")),
			svc.Command,
		})
	}

	tbl := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(mutedStyle).
		BorderHeader(true).
		BorderRow(true).
		Headers("ID", "LABEL", "PORT", "FROM", "DEPENDS ON", "COMMAND").
		Rows(rows...).
		Width(min(width, 120)).
		Wrap(true).
		StyleFunc(serviceTableStyle)

	return tbl.Render()
}

func serviceTableStyle(row, col int) lipgloss.Style {
	if row == table.HeaderRow {
		return mutedStyle.Bold(true)
	}
	switch col {
	case 0:
		return selectedStyle
	case 1:
		return titleStyle
	default:
		return lipgloss.NewStyle()
	}
}

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
