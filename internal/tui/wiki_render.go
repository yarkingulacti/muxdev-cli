package tui

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"

	"github.com/yarkingulacti/muxdev-cli/internal/cli/doc"
)

var (
	wikiAccentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
	wikiCodeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("236")).Padding(0, 1)
	wikiCalloutStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.Color("214")).
				Foreground(lipgloss.Color("252")).
				PaddingLeft(1)
	wikiSectionStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	wikiBulletStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("81"))
	wikiIDStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Background(lipgloss.Color("235")).Padding(0, 1)
)

func pageSummary(page doc.Page) string {
	if s := strings.TrimSpace(page.Summary); s != "" {
		return s
	}
	return firstSentence(page.Body)
}

func firstSentence(body string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasSuffix(line, ":") {
			continue
		}
		if len(line) > 88 {
			return line[:85] + "…"
		}
		return line
	}
	return ""
}

func renderRichDoc(body string) string {
	lines := strings.Split(body, "\n")
	var b strings.Builder
	i := 0
	for i < len(lines) {
		line := lines[i]
		trim := strings.TrimSpace(line)

		if trim == "" {
			b.WriteString("\n")
			i++
			continue
		}

		if isCodeBlockLine(line) {
			var block []string
			for i < len(lines) && (isCodeBlockLine(lines[i]) || strings.TrimSpace(lines[i]) == "") {
				if strings.TrimSpace(lines[i]) != "" {
					block = append(block, strings.TrimLeft(lines[i], " \t"))
				}
				i++
			}
			b.WriteString(wikiCodeStyle.Render(strings.Join(block, "\n")))
			b.WriteString("\n\n")
			continue
		}

		if isCalloutLine(trim) {
			text := trim
			text = strings.TrimPrefix(text, "> ")
			for _, prefix := range []string{"Tip:", "Note:"} {
				if strings.HasPrefix(text, prefix) {
					text = wikiAccentStyle.Render(prefix) + text[len(prefix):]
					break
				}
			}
			b.WriteString(wikiCalloutStyle.Render(text))
			b.WriteString("\n\n")
			i++
			continue
		}

		if isSectionHeading(trim, line) {
			b.WriteString(wikiSectionStyle.Render(trim))
			b.WriteString("\n")
			i++
			continue
		}

		if isBulletLine(trim) {
			text := strings.TrimLeft(trim, "•-*0123456789. ")
			b.WriteString(wikiBulletStyle.Render("  • "))
			b.WriteString(highlightInline(text))
			b.WriteString("\n")
			i++
			continue
		}

		if strings.HasPrefix(trim, "  --") || strings.HasPrefix(trim, "  -") {
			b.WriteString(mutedStyle.Render(line))
			b.WriteString("\n")
			i++
			continue
		}

		b.WriteString(highlightInline(line))
		b.WriteString("\n")
		i++
	}
	return strings.TrimRight(b.String(), "\n")
}

func isCodeBlockLine(line string) bool {
	trim := strings.TrimSpace(line)
	if trim == "" || isBulletLine(trim) || isSectionHeading(trim, line) {
		return false
	}
	if strings.HasPrefix(line, "  ") && !strings.HasPrefix(trim, "•") && !strings.HasPrefix(trim, "-") {
		if strings.HasPrefix(trim, "muxdev") || strings.Contains(trim, "${") || strings.Contains(trim, "npm ") {
			return true
		}
		if strings.Contains(trim, ":") && strings.Contains(trim, "/") {
			return true
		}
	}
	return strings.HasPrefix(trim, "$ ") || strings.HasPrefix(trim, "# ")
}

func isCalloutLine(trim string) bool {
	return strings.HasPrefix(trim, "> ") ||
		strings.HasPrefix(trim, "Tip:") ||
		strings.HasPrefix(trim, "Note:")
}

func isSectionHeading(trim, raw string) bool {
	if strings.HasPrefix(raw, " ") || strings.HasPrefix(raw, "\t") {
		return false
	}
	if !strings.HasSuffix(trim, ":") {
		return false
	}
	if strings.HasPrefix(trim, "http") {
		return false
	}
	name := strings.TrimSuffix(trim, ":")
	if name == "" {
		return false
	}
	for _, r := range name {
		if unicode.IsLetter(r) || r == ' ' || r == '&' || r == '/' {
			continue
		}
		return false
	}
	return true
}

func isBulletLine(trim string) bool {
	if strings.HasPrefix(trim, "• ") {
		return true
	}
	if strings.HasPrefix(trim, "- ") && !strings.HasPrefix(trim, "--") {
		return true
	}
	if len(trim) > 2 && trim[0] >= '0' && trim[0] <= '9' && strings.Contains(trim[:min(4, len(trim))], ".") {
		return true
	}
	return false
}

func highlightInline(s string) string {
	if !strings.Contains(s, "muxdev") {
		return s
	}
	parts := strings.Split(s, "muxdev")
	if len(parts) == 1 {
		return s
	}
	var b strings.Builder
	for i, part := range parts {
		b.WriteString(part)
		if i < len(parts)-1 {
			b.WriteString(wikiAccentStyle.Render("muxdev"))
		}
	}
	return b.String()
}

func renderWikiTopicRow(selected bool, page doc.Page) string {
	title := page.Title
	marker := "  "
	if selected {
		marker = cursorStyle.Render("› ")
		title = selectedStyle.Render(title)
	}
	id := wikiIDStyle.Render(page.ID)
	summary := mutedStyle.Render("      " + pageSummary(page))
	return marker + title + "  " + id + "\n" + summary
}

func renderWikiCategoryHeader(name string, count int) string {
	label := wikiSectionStyle.Render(name)
	countBadge := mutedStyle.Render(" (" + itoa(count) + ")")
	return label + countBadge
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
