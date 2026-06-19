package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
)

// LogPagination describes scroll position within a log viewport.
type LogPagination struct {
	TotalLines int
	TopLine    int
	BottomLine int
	Page       int
	TotalPages int
	AtBottom   bool
}

func viewportPagination(vp viewport.Model) LogPagination {
	total := vp.TotalLineCount()
	if total == 0 {
		return LogPagination{}
	}

	visible := vp.VisibleLineCount()
	if visible < 1 {
		visible = 1
	}

	top := vp.YOffset + 1
	bottom := vp.YOffset + visible
	if bottom > total {
		bottom = total
	}
	if top > total {
		top = total
	}

	height := vp.Height
	if height < 1 {
		height = 1
	}

	totalPages := (total + height - 1) / height
	if totalPages < 1 {
		totalPages = 1
	}

	page := (vp.YOffset / height) + 1
	if vp.AtBottom() {
		page = totalPages
	}
	if page > totalPages {
		page = totalPages
	}
	if page < 1 {
		page = 1
	}

	return LogPagination{
		TotalLines: total,
		TopLine:    top,
		BottomLine: bottom,
		Page:       page,
		TotalPages: totalPages,
		AtBottom:   vp.AtBottom(),
	}
}

func formatLogPagination(p LogPagination, live bool) string {
	if p.TotalLines == 0 {
		return "0 lines"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "lines %d-%d of %d", p.TopLine, p.BottomLine, p.TotalLines)
	if p.TotalPages > 1 {
		fmt.Fprintf(&b, "  ·  page %d/%d", p.Page, p.TotalPages)
	}
	if live && p.AtBottom {
		b.WriteString("  ·  live")
	} else if !p.AtBottom {
		b.WriteString("  ·  history")
	}
	return b.String()
}

func (m runnerModel) logPaginationLabel() string {
	if !m.ready || m.filterMenu || m.rerunMenu {
		return ""
	}
	return formatLogPagination(viewportPagination(m.viewport), m.followTail)
}
