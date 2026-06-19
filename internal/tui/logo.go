package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var logoLines = []string{
	"███╗   ███╗██╗   ██╗██╗  ██╗",
	"████╗ ████║██║   ██║╚██╗██╔╝",
	"██╔████╔██║██║   ██║ ╚███╔╝ ",
	"██║╚██╔╝██║╚██╗ ██╔╝ ██╔██╗ ",
	"██║ ╚═╝ ██║ ╚████╔╝ ██╔╝ ██╗",
	"╚═╝     ╚═╝  ╚═══╝  ╚═╝  ╚═╝",
	" ██████╗ ███████╗██╗   ██╗",
	" ██╔══██╗██╔════╝██║   ██║",
	" ██║  ██║█████╗  ██║   ██║",
	" ██║  ██║██╔══╝  ╚██╗ ██╔╝",
	" ██████╔╝███████╗ ╚████╔╝ ",
	" ╚═════╝ ╚══════╝  ╚═══╝  ",
}

var (
	logoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	taglineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func renderLogo(width int) string {
	var art strings.Builder
	for _, line := range logoLines {
		art.WriteString(logoStyle.Render(line))
		art.WriteString("\n")
	}
	art.WriteString(taglineStyle.Render("Multiplexed dev stack runner"))

	maxW := min(width-2, 72)
	return lipgloss.NewStyle().Width(maxW).Align(lipgloss.Center).Render(strings.TrimRight(art.String(), "\n"))
}
