package ui

import (
	"strings"

	"github.com/Beargruug/sentry-tui/internal/config"
	"github.com/Beargruug/sentry-tui/internal/ui/styles"
)

func (m Model) viewConfig() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render(" ⚙️  Configuration ") + "\n\n")

	labels := []string{
		"Auth Token",
		"Organization",
		"Default Project",
		"Refresh (secs)",
	}

	for i, label := range labels {
		cursor := "  "
		if i == m.configCursor {
			cursor = styles.SelectedItem.Render("▸ ")
		}
		b.WriteString(cursor + styles.DetailLabel.Render(label) + "\n")
		if i < len(m.configInputs) {
			b.WriteString("    " + m.configInputs[i].View() + "\n\n")
		}
	}

	b.WriteString("\n  " + styles.Subtitle.Render("tab/↓ next field · shift+tab/↑ prev · enter save · esc cancel") + "\n")

	// Config file path
	cfgPath, err := config.ConfigPath()
	if err == nil {
		b.WriteString("\n  " + styles.Subtitle.Render("Config: "+cfgPath) + "\n")
	}

	b.WriteString("\n  " + styles.Subtitle.Render("Environment variables: SENTRY_AUTH_TOKEN, SENTRY_ORG, SENTRY_PROJECT, SENTRY_BASE_URL") + "\n")

	content := b.String()
	lines := strings.Count(content, "\n")
	for lines < m.height-1 {
		content += "\n"
		lines++
	}

	return content + m.renderFooter()
}
