package doc_test

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yarkingulacti/muxdev-cli/internal/cli/doc"
)

func TestBuildPagesIncludesCommandsAndStatic(t *testing.T) {
	root := &cobra.Command{Use: "muxdev", Short: "root"}
	root.AddCommand(&cobra.Command{Use: "list", Short: "List services"})
	root.AddCommand(&cobra.Command{Use: "version", Short: "Show version"})

	pages := doc.BuildPages(root)
	if len(pages) < 4 {
		t.Fatalf("pages = %d, want static + commands", len(pages))
	}

	foundList := false
	for _, p := range pages {
		if strings.Contains(p.Title, "list") {
			foundList = true
			if p.TryCommand == "" {
				t.Fatal("list page should have try command")
			}
		}
	}
	if !foundList {
		t.Fatal("missing list command page")
	}
}

func TestFindPage(t *testing.T) {
	pages := []doc.Page{{ID: "runtime-tui", Title: "Runtime TUI"}}
	if _, ok := doc.FindPage(pages, "runtime"); !ok {
		t.Fatal("expected fuzzy find")
	}
}

func TestFindPageIndex(t *testing.T) {
	pages := []doc.Page{
		{ID: "welcome", Title: "Welcome"},
		{ID: "list", Title: "muxdev list"},
	}
	if got := doc.FindPageIndex(pages, "list"); got != 1 {
		t.Fatalf("FindPageIndex() = %d, want 1", got)
	}
	if got := doc.FindPageIndex(pages, "missing"); got != -1 {
		t.Fatalf("FindPageIndex() = %d, want -1", got)
	}
}
