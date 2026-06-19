package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
)

func TestViewportPagination(t *testing.T) {
	vp := viewport.New(40, 5)
	lines := make([]string, 23)
	for i := range lines {
		lines[i] = "log line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	p := viewportPagination(vp)
	if p.TotalLines != 23 || p.TopLine != 1 || p.BottomLine != 5 || p.Page != 1 || p.TotalPages != 5 {
		t.Fatalf("top page = %+v", p)
	}

	vp.SetYOffset(10)
	p = viewportPagination(vp)
	if p.TopLine != 11 || p.BottomLine != 15 || p.Page != 3 {
		t.Fatalf("middle page = %+v", p)
	}

	vp.GotoBottom()
	p = viewportPagination(vp)
	if !p.AtBottom || p.BottomLine != 23 || p.Page != 5 {
		t.Fatalf("bottom page = %+v", p)
	}
}

func TestFormatLogPaginationLive(t *testing.T) {
	got := formatLogPagination(LogPagination{
		TotalLines: 100,
		TopLine:    81,
		BottomLine: 100,
		Page:       5,
		TotalPages: 5,
		AtBottom:   true,
	}, true)
	if !strings.Contains(got, "page 5/5") || !strings.Contains(got, "live") {
		t.Fatalf("format = %q", got)
	}
}
