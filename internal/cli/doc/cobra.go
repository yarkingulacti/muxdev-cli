package doc

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Page is a single help/wiki entry.
type Page struct {
	ID         string
	Category   string
	Title      string
	Body       string
	TryCommand string
}

// BuildPages returns static guides plus one page per registered cobra command.
func BuildPages(root *cobra.Command) []Page {
	pages := append([]Page{}, staticPages()...)
	pages = append(pages, pagesFromCommand(root, "")...)
	return pages
}

func pagesFromCommand(cmd *cobra.Command, prefix string) []Page {
	if cmd == nil || cmd.Name() == "help" || cmd.Hidden {
		return nil
	}

	var pages []Page
	namePath := cmd.Name()
	if prefix != "" {
		namePath = prefix + " " + cmd.Name()
	}

	if cmd.Short != "" && cmd != cmd.Root() {
		pages = append(pages, Page{
			ID:         slug(namePath),
			Category:   categoryFor(cmd),
			Title:      "muxdev " + namePath,
			Body:       renderCommandBody(cmd),
			TryCommand: tryCommandFor(cmd, namePath),
		})
	}

	for _, sub := range cmd.Commands() {
		childPrefix := namePath
		if cmd == cmd.Root() {
			childPrefix = ""
		}
		pages = append(pages, pagesFromCommand(sub, childPrefix)...)
	}
	return pages
}

func categoryFor(cmd *cobra.Command) string {
	switch cmd.Name() {
	case "init", "configure", "config":
		return "Setup"
	case "logs":
		return "Logs & sessions"
	case "update", "version":
		return "Install & updates"
	case "list", "ls":
		return "Reference"
	default:
		if cmd.Parent() != nil && cmd.Parent().Name() == "logs" {
			return "Logs & sessions"
		}
		return "Commands"
	}
}

func renderCommandBody(cmd *cobra.Command) string {
	var b strings.Builder

	if cmd.Long != "" {
		b.WriteString(strings.TrimSpace(cmd.Long))
		b.WriteString("\n\n")
	} else if cmd.Short != "" {
		b.WriteString(cmd.Short)
		b.WriteString("\n\n")
	}

	if aliases := cmd.Aliases; len(aliases) > 0 {
		b.WriteString("Aliases: ")
		b.WriteString(strings.Join(aliases, ", "))
		b.WriteString("\n\n")
	}

	if cmd.HasSubCommands() {
		b.WriteString("Subcommands\n")
		for _, sub := range cmd.Commands() {
			if sub.Hidden {
				continue
			}
			b.WriteString("  ")
			b.WriteString(sub.Name())
			if sub.Short != "" {
				b.WriteString("  —  ")
				b.WriteString(sub.Short)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	writeFlags := func(title string, set *pflag.FlagSet) {
		if set == nil {
			return
		}
		var lines []string
		set.VisitAll(func(f *pflag.Flag) {
			if f.Hidden || f.Name == "help" {
				return
			}
			line := "  --" + f.Name
			if f.Shorthand != "" {
				line += ", -" + f.Shorthand
			}
			if f.DefValue != "" {
				line += " (default " + f.DefValue + ")"
			}
			line += "\n      " + f.Usage
			lines = append(lines, line)
		})
		if len(lines) == 0 {
			return
		}
		b.WriteString(title)
		b.WriteString("\n")
		b.WriteString(strings.Join(lines, "\n"))
		b.WriteString("\n\n")
	}

	writeFlags("Flags", cmd.NonInheritedFlags())
	writeFlags("Inherited flags", cmd.InheritedFlags())

	b.WriteString("Try it: press t in the interactive guide, or run:\n  ")
	b.WriteString(tryCommandFor(cmd, strings.TrimPrefix(cmd.CommandPath(), "muxdev ")))
	b.WriteString("\n")

	return strings.TrimRight(b.String(), "\n")
}

func tryCommandFor(cmd *cobra.Command, namePath string) string {
	path := "muxdev " + strings.TrimSpace(namePath)
	switch cmd.Name() {
	case "version":
		return path + " --short"
	case "list", "ls":
		return path
	case "path":
		return path
	case "logs":
		return path + " path"
	default:
		if cmd.HasSubCommands() {
			return "muxdev help " + slug(namePath)
		}
		return "muxdev help " + slug(namePath)
	}
}

func slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.NewReplacer(" ", "-", "/", "-", "_", "-").Replace(s)
	return s
}

// FindPage resolves a topic name or id from user input.
func FindPage(pages []Page, query string) (Page, bool) {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return Page{}, false
	}
	for _, p := range pages {
		if strings.ToLower(p.ID) == query {
			return p, true
		}
	}
	for _, p := range pages {
		if strings.ToLower(p.Title) == query {
			return p, true
		}
	}
	for _, p := range pages {
		if strings.Contains(strings.ToLower(p.ID), query) ||
			strings.Contains(strings.ToLower(p.Title), query) {
			return p, true
		}
	}
	return Page{}, false
}

// FindPageIndex returns the index of the page matching query, or -1.
func FindPageIndex(pages []Page, query string) int {
	page, ok := FindPage(pages, query)
	if !ok {
		return -1
	}
	for i, p := range pages {
		if p.ID == page.ID {
			return i
		}
	}
	return -1
}

// IndexPlain renders a text table of contents for non-interactive help.
func IndexPlain(pages []Page) string {
	var b strings.Builder
	b.WriteString("muxdev guide — topics\n\n")
	lastCat := ""
	for _, p := range pages {
		if p.Category != lastCat {
			lastCat = p.Category
			b.WriteString(lastCat)
			b.WriteString("\n")
		}
		b.WriteString("  ")
		b.WriteString(p.Title)
		b.WriteString("  (")
		b.WriteString(p.ID)
		b.WriteString(")\n")
	}
	b.WriteString("\nInteractive wiki: muxdev help\n")
	b.WriteString("Plain topic:      muxdev help <topic>\n")
	return b.String()
}
