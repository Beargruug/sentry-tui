package ui

import (
	"strings"

	"github.com/Beargruug/sentry-tui/internal/ui/styles"
)

func (m Model) viewHelp() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render(" 🔭 Sentry TUI — Help ") + "\n\n")

	sections := []struct {
		title string
		keys  []struct{ key, desc string }
	}{
		{
			title: "Navigation",
			keys: []struct{ key, desc string }{
				{"↑/k", "Move cursor up"},
				{"↓/j", "Move cursor down"},
				{"g", "Go to top"},
				{"G", "Go to bottom"},
				{"enter", "Open issue detail"},
				{"esc/backspace", "Go back"},
				{"n / ctrl+f", "Next page"},
				{"p / ctrl+b", "Previous page"},
				{"tab", "Next section"},
			},
		},
		{
			title: "Issue Actions",
			keys: []struct{ key, desc string }{
				{"R", "Resolve / unresolve issue"},
				{"i", "Ignore issue"},
				{"a", "Assign issue to team member"},
				{"o", "Show issue permalink"},
			},
		},
		{
			title: "Search & Filter",
			keys: []struct{ key, desc string }{
				{"/", "Open search bar"},
				{"f", "Open filter panel"},
				{"r", "Refresh issues"},
			},
		},
		{
			title: "General",
			keys: []struct{ key, desc string }{
				{"?", "Toggle this help"},
				{"C", "Open configuration"},
				{"q / ctrl+c", "Quit application"},
			},
		},
		{
			title: "Filter Panel",
			keys: []struct{ key, desc string }{
				{"1", "Filter: Unresolved"},
				{"2", "Filter: Resolved"},
				{"3", "Filter: Ignored"},
				{"4", "Filter: All statuses"},
				{"5", "Cycle sort mode"},
				{"6", "Cycle project filter"},
				{"0", "Reset all filters"},
			},
		},
	}

	for _, section := range sections {
		b.WriteString(styles.DetailHeader.Render("━━━ "+section.title+" ━━━") + "\n\n")
		for _, k := range section.keys {
			keyStr := styles.HelpKey.Render(padRight(k.key, 16))
			descStr := styles.HelpDesc.Render(k.desc)
			b.WriteString("  " + keyStr + " " + descStr + "\n")
		}
		b.WriteString("\n")
	}

	// Pad
	content := b.String()
	lines := strings.Count(content, "\n")
	for lines < m.height-1 {
		content += "\n"
		lines++
	}

	return content + m.renderFooter()
}
