package ui

import (
	"strings"

	"github.com/Beargruug/sentry-tui/internal/ui/styles"
)

func (m Model) viewSetup() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render(" 🔭 Sentry TUI — First Run Setup ") + "\n\n")
	b.WriteString("  Welcome! Let's configure your Sentry connection.\n\n")

	steps := []struct {
		label  string
		detail string
	}{
		{"Auth Token", "Go to Sentry → Settings → Auth Tokens to create one"},
		{"Organization", "Your Sentry organization slug (from the URL)"},
		{"Default Project", "Optional: set a default project filter"},
	}

	for i, step := range steps {
		icon := "○"
		if i < int(m.setupStep) {
			icon = styles.SuccessMsg.Render("✓")
		} else if i == int(m.setupStep) {
			icon = styles.ErrorMsg.Render("▸")
		}

		b.WriteString("  " + icon + " " + styles.DetailLabel.Render(step.label) + "\n")
		b.WriteString("    " + styles.Subtitle.Render(step.detail) + "\n")

		if i == int(m.setupStep) && i < len(m.setupInputs) {
			b.WriteString("\n    " + m.setupInputs[i].View() + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("  " + styles.Subtitle.Render("Press enter to continue · esc to quit") + "\n")

	content := b.String()
	lines := strings.Count(content, "\n")
	for lines < m.height-1 {
		content += "\n"
		lines++
	}

	return content
}
