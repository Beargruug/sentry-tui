package ui

import (
	"fmt"
	"strings"

	"github.com/Beargruug/sentry-tui/internal/ui/styles"
)

func (m Model) viewAssign() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render(" 👤 Assign Issue ") + "\n\n")

	// Current issue info
	issueTitle := ""
	if len(m.issues) > 0 && m.cursor < len(m.issues) {
		issueTitle = m.issues[m.cursor].ShortID + " — " + truncate(m.issues[m.cursor].Title, 60)
	}
	if m.detailIssue.ID != "" {
		issueTitle = m.detailIssue.ShortID + " — " + truncate(m.detailIssue.Title, 60)
	}
	if issueTitle != "" {
		b.WriteString("  " + styles.Subtitle.Render("Issue: "+issueTitle) + "\n\n")
	}

	b.WriteString("  Type email/username or select from list:\n\n")
	b.WriteString("  " + m.assignInput.View() + "\n\n")

	// Member list
	if len(m.members) > 0 {
		b.WriteString(styles.DetailHeader.Render("━━━ Team Members ━━━") + "\n\n")
		visibleStart := 0
		maxVisible := m.height - 15
		if maxVisible < 5 {
			maxVisible = 5
		}
		if m.assignCursor >= maxVisible {
			visibleStart = m.assignCursor - maxVisible + 1
		}
		visibleEnd := visibleStart + maxVisible
		if visibleEnd > len(m.members) {
			visibleEnd = len(m.members)
		}

		for i := visibleStart; i < visibleEnd; i++ {
			member := m.members[i]
			name := member.Name
			if name == "" {
				name = member.User.Name
			}
			if name == "" {
				name = member.Email
			}
			email := member.Email
			role := member.Role

			cursor := "  "
			nameStyle := styles.NormalItem
			if i == m.assignCursor {
				cursor = styles.SelectedItem.Render("▸ ")
				nameStyle = styles.SelectedItem
			}

			b.WriteString(fmt.Sprintf("  %s%s <%s> [%s]\n",
				cursor,
				nameStyle.Render(name),
				styles.Subtitle.Render(email),
				styles.Subtitle.Render(role),
			))
		}
	} else {
		b.WriteString("  " + styles.Subtitle.Render("No team members loaded. Type an email and press enter.") + "\n")
	}

	b.WriteString("\n" + styles.Subtitle.Render("  ↑↓ select · enter confirm · esc cancel") + "\n")

	content := b.String()
	lines := strings.Count(content, "\n")
	for lines < m.height-1 {
		content += "\n"
		lines++
	}

	return content + m.renderFooter()
}
