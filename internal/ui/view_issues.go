package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/Beargruug/sentry-tui/internal/ui/styles"
)

func (m Model) viewIssueList() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString(m.renderSearchBar())

	if len(m.issues) == 0 && !m.loading {
		b.WriteString("\n  No issues found. Try changing your filters or refreshing.\n")
		b.WriteString(m.renderFooter())
		return b.String()
	}

	// Calculate visible area
	headerLines := 2
	footerLines := 1
	searchLines := 0
	if m.searching {
		searchLines = 3
	}
	issueLineHeight := 3 // lines per issue row
	availableHeight := m.height - headerLines - footerLines - searchLines
	visibleCount := availableHeight / issueLineHeight
	if visibleCount < 1 {
		visibleCount = 1
	}

	// Scroll window
	start := 0
	if m.cursor >= visibleCount {
		start = m.cursor - visibleCount + 1
	}
	end := start + visibleCount
	if end > len(m.issues) {
		end = len(m.issues)
	}

	for idx := start; idx < end; idx++ {
		issue := m.issues[idx]
		selected := idx == m.cursor

		// Level badge
		level := styles.LevelStyle(issue.Level).Render(strings.ToUpper(issue.Level))

		// Status icon
		statusIcon := "●"
		switch issue.Status {
		case "resolved":
			statusIcon = styles.SuccessMsg.Render("✓")
		case "ignored":
			statusIcon = styles.Subtitle.Render("⊘")
		default:
			statusIcon = styles.ErrorMsg.Render("●")
		}

		// Title line
		titleStr := truncate(issue.Title, m.width-30)
		if selected {
			titleStr = styles.SelectedItem.Render(titleStr)
		} else {
			titleStr = styles.NormalItem.Render(titleStr)
		}

		// Meta line
		project := styles.Subtitle.Render(issue.Project.Slug)
		count := styles.Subtitle.Render(fmt.Sprintf("×%s", issue.Count))
		users := styles.Subtitle.Render(fmt.Sprintf("👤%d", issue.UserCount))
		lastSeen := styles.Subtitle.Render(relativeTime(issue.LastSeen))

		assignee := ""
		if issue.AssignedTo != nil {
			name := issue.AssignedTo.Name
			if name == "" {
				name = issue.AssignedTo.Email
			}
			assignee = styles.Subtitle.Render(fmt.Sprintf(" → %s", name))
		}

		// Cursor indicator
		cursor := "  "
		if selected {
			cursor = styles.SelectedItem.Render("▸ ")
		}

		line1 := fmt.Sprintf("%s%s %s %s", cursor, statusIcon, level, titleStr)
		line2 := fmt.Sprintf("    %s  %s  %s  %s%s", project, count, users, lastSeen, assignee)

		b.WriteString(line1 + "\n")
		b.WriteString(line2 + "\n")

		// Separator between items
		if idx < end-1 {
			sep := styles.Subtitle.Render(strings.Repeat("─", min(m.width-4, 80)))
			b.WriteString("  " + sep + "\n")
		}
	}

	// Pagination indicator
	if m.pageCursor.HasPrev || m.pageCursor.HasNext {
		b.WriteString("\n")
		nav := "  "
		if m.pageCursor.HasPrev {
			nav += styles.StatusKey.Render("← p") + " prev  "
		}
		if m.pageCursor.HasNext {
			nav += styles.StatusKey.Render("n →") + " next"
		}
		b.WriteString(nav + "\n")
	}

	// Fill remaining space
	content := b.String()
	contentLines := strings.Count(content, "\n")
	for contentLines < m.height-1 {
		content += "\n"
		contentLines++
	}

	return content + m.renderFooter()
}

// ---------- Helpers ----------

func truncate(s string, maxLen int) string {
	if maxLen < 4 {
		maxLen = 4
	}
	// Count runes for proper unicode handling
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf("%dh ago", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("Jan 02")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}
