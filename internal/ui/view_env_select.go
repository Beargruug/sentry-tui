package ui

import (
	"fmt"
	"strings"

	"github.com/Beargruug/sentry-tui/internal/models"
	"github.com/Beargruug/sentry-tui/internal/ui/styles"
)

// filteredEnvironments returns environments matching the current search input.
func (m Model) filteredEnvironments() []models.Environment {
	query := strings.ToLower(m.envSelectInput.Value())
	if query == "" {
		return m.environments
	}
	var filtered []models.Environment
	for _, e := range m.environments {
		if strings.Contains(strings.ToLower(e.Name), query) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func (m Model) viewEnvSelect() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render(" Select Environment ") + "\n\n")

	b.WriteString("  " + m.envSelectInput.View() + "\n\n")

	filtered := m.filteredEnvironments()

	if len(filtered) == 0 && len(m.environments) == 0 {
		b.WriteString("  " + styles.Subtitle.Render("No environments loaded.") + "\n")
	} else if len(filtered) == 0 {
		b.WriteString("  " + styles.Subtitle.Render("No environments match your search.") + "\n")
	} else {
		// Show "All environments" option first
		allCursor := "  "
		allStyle := styles.NormalItem
		if m.envSelectCursor == 0 {
			allCursor = styles.SelectedItem.Render("▸ ")
			allStyle = styles.SelectedItem
		}
		b.WriteString(fmt.Sprintf("  %s%s\n", allCursor, allStyle.Render("All environments")))

		maxVisible := m.height - 12
		if maxVisible < 5 {
			maxVisible = 5
		}

		visibleStart := 0
		if m.envSelectCursor-1 >= maxVisible {
			visibleStart = m.envSelectCursor - maxVisible
		}
		visibleEnd := visibleStart + maxVisible
		if visibleEnd > len(filtered) {
			visibleEnd = len(filtered)
		}

		for i := visibleStart; i < visibleEnd; i++ {
			e := filtered[i]
			cursor := "  "
			nameStyle := styles.NormalItem
			if i+1 == m.envSelectCursor {
				cursor = styles.SelectedItem.Render("▸ ")
				nameStyle = styles.SelectedItem
			}
			selected := ""
			if e.Name == m.filter.Environment {
				selected = styles.SuccessMsg.Render(" [active]")
			}
			b.WriteString(fmt.Sprintf("  %s%s%s\n", cursor, nameStyle.Render(e.Name), selected))
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
