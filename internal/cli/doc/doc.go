package doc

import (
	"strings"

	"github.com/spf13/cobra"
)

// Attach wires help metadata onto the command tree for the help command wiki.
func Attach(root *cobra.Command) {
	ensureLong(root)
	for _, cmd := range root.Commands() {
		ensureLong(cmd)
	}
}

func ensureLong(cmd *cobra.Command) {
	if cmd == nil {
		return
	}
	if cmd.Long == "" && cmd.Short != "" {
		cmd.Long = cmd.Short
	}
	for _, sub := range cmd.Commands() {
		ensureLong(sub)
	}
}
// MatchPages filters pages by a search query (title, id, body).
func MatchPages(pages []Page, query string) []Page {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return pages
	}
	out := make([]Page, 0, len(pages))
	for _, p := range pages {
		if strings.Contains(strings.ToLower(p.ID), query) ||
			strings.Contains(strings.ToLower(p.Title), query) ||
			strings.Contains(strings.ToLower(p.Body), query) ||
			strings.Contains(strings.ToLower(p.Category), query) {
			out = append(out, p)
		}
	}
	return out
}

// Categories returns unique category names in first-seen order.
func Categories(pages []Page) []string {
	seen := make(map[string]bool)
	var out []string
	for _, p := range pages {
		if seen[p.Category] {
			continue
		}
		seen[p.Category] = true
		out = append(out, p.Category)
	}
	return out
}
