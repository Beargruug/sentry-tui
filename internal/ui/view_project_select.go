package ui

import (
	"fmt"
	"strings"

	"github.com/Beargruug/sentry-tui/internal/models"
	"github.com/Beargruug/sentry-tui/internal/ui/styles"
)

// filteredProjects returns projects matching the current search input.
func (m Model) filteredProjects() []models.Project {
	query := strings.ToLower(m.projectSelectInput.Value())
	if query == "" {
		return m.projects
	}
	var filtered []models.Project
	for _, p := range m.projects {
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Slug), query) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func (m Model) viewProjectSelect() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render(" Select Project ") + "\n\n")

	b.WriteString("  " + m.projectSelectInput.View() + "\n\n")

	filtered := m.filteredProjects()

	if len(filtered) == 0 {
		b.WriteString("  " + styles.Subtitle.Render("No projects match your search.") + "\n")
	} else {
		// Show "All projects" option first
		allCursor := "  "
		allStyle := styles.NormalItem
		if m.projectSelectCursor == 0 {
			allCursor = styles.SelectedItem.Render("▸ ")
			allStyle = styles.SelectedItem
		}
		b.WriteString(fmt.Sprintf("  %s%s\n", allCursor, allStyle.Render("All projects")))

		maxVisible := m.height - 12
		if maxVisible < 5 {
			maxVisible = 5
		}

		visibleStart := 0
		// cursor 0 = "all", 1..N = projects
		if m.projectSelectCursor-1 >= maxVisible {
			visibleStart = m.projectSelectCursor - maxVisible
		}
		visibleEnd := visibleStart + maxVisible
		if visibleEnd > len(filtered) {
			visibleEnd = len(filtered)
		}

		for i := visibleStart; i < visibleEnd; i++ {
			p := filtered[i]
			cursor := "  "
			nameStyle := styles.NormalItem
			// cursor index for projects is i+1 (0 is "all")
			if i+1 == m.projectSelectCursor {
				cursor = styles.SelectedItem.Render("▸ ")
				nameStyle = styles.SelectedItem
			}
			selected := ""
			if p.Slug == m.filter.Project {
				selected = styles.SuccessMsg.Render(" [active]")
			}
			b.WriteString(fmt.Sprintf("  %s%s %s%s\n",
				cursor,
				nameStyle.Render(p.Name),
				styles.Subtitle.Render("("+p.Slug+")"),
				selected,
			))
		}

		if visibleEnd < len(filtered) {
			b.WriteString(fmt.Sprintf("\n  %s\n", styles.Subtitle.Render(fmt.Sprintf("  ... and %d more", len(filtered)-visibleEnd))))
		}
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
