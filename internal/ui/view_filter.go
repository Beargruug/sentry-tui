package ui

import (
	"fmt"
	"strings"

	"github.com/Beargruug/sentry-tui/internal/ui/styles"
)

func (m Model) viewFilter() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render(" 🔭 Filter & Sort ") + "\n\n")

	// Current filters summary
	b.WriteString(styles.DetailHeader.Render("━━━ Current Filters ━━━") + "\n\n")

	statusDisplay := m.filter.Status
	if statusDisplay == "" {
		statusDisplay = "all"
	}
	projectDisplay := m.filter.Project
	if projectDisplay == "" {
		projectDisplay = "all"
	}
	envDisplay := m.filter.Environment
	if envDisplay == "" {
		envDisplay = "all"
	}
	sortDisplay := m.filter.Sort

	b.WriteString(fmt.Sprintf("  %s %s\n", styles.DetailLabel.Render("Status:"), statusDisplay))
	b.WriteString(fmt.Sprintf("  %s %s\n", styles.DetailLabel.Render("Project:"), projectDisplay))
	b.WriteString(fmt.Sprintf("  %s %s\n", styles.DetailLabel.Render("Environment:"), envDisplay))
	b.WriteString(fmt.Sprintf("  %s %s\n", styles.DetailLabel.Render("Sort:"), sortDisplay))
	if m.filter.Query != "" {
		b.WriteString(fmt.Sprintf("  %s \"%s\"\n", styles.DetailLabel.Render("Search:"), m.filter.Query))
	}

	b.WriteString("\n" + styles.DetailHeader.Render("━━━ Quick Actions ━━━") + "\n\n")

	options := []struct {
		key  string
		desc string
	}{
		{"1", "Unresolved issues only"},
		{"2", "Resolved issues only"},
		{"3", "Ignored issues only"},
		{"4", "All statuses"},
		{"5", fmt.Sprintf("Cycle sort (current: %s)", sortDisplay)},
		{"6", fmt.Sprintf("Select project (current: %s)", projectDisplay)},
		{"7", fmt.Sprintf("Select environment (current: %s)", envDisplay)},
		{"0", "Reset all filters"},
	}

	for _, opt := range options {
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			styles.HelpKey.Render(padRight(opt.key, 4)),
			styles.HelpDesc.Render(opt.desc),
		))
	}

	b.WriteString("\n" + styles.Subtitle.Render("  Press esc to go back") + "\n")

	content := b.String()
	lines := strings.Count(content, "\n")
	for lines < m.height-1 {
		content += "\n"
		lines++
	}

	return content + m.renderFooter()
}
