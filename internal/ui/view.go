package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/Beargruug/sentry-tui/internal/ui/styles"
)

// View renders the current screen.
func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	switch m.currentView {
	case ViewSetup:
		return m.viewSetup()
	case ViewIssueList:
		return m.viewIssueList()
	case ViewIssueDetail:
		return m.viewIssueDetail()
	case ViewHelp:
		return m.viewHelp()
	case ViewFilter:
		return m.viewFilter()
	case ViewAssign:
		return m.viewAssign()
	case ViewConfig:
		return m.viewConfig()
	case ViewProjectSelect:
		return m.viewProjectSelect()
	case ViewEnvSelect:
		return m.viewEnvSelect()
	default:
		return "Unknown view"
	}
}

// ---------- Header / Footer ----------

func (m Model) renderHeader() string {
	title := styles.Title.Render(" 🔭 Sentry TUI ")
	org := styles.Subtitle.Render(fmt.Sprintf(" %s ", m.cfg.Organization))

	// Active filters indicator
	filters := []string{}
	if m.filter.Project != "" {
		filters = append(filters, fmt.Sprintf("project:%s", m.filter.Project))
	}
	if m.filter.Status != "" {
		filters = append(filters, fmt.Sprintf("status:%s", m.filter.Status))
	}
	if m.filter.Query != "" {
		filters = append(filters, fmt.Sprintf("search:\"%s\"", m.filter.Query))
	}
	filterStr := ""
	if len(filters) > 0 {
		filterStr = styles.Subtitle.Render(" [" + strings.Join(filters, " · ") + "]")
	}

	left := lipgloss.JoinHorizontal(lipgloss.Center, title, org, filterStr)

	pageInfo := ""
	if m.currentView == ViewIssueList {
		pageInfo = styles.Subtitle.Render(fmt.Sprintf(" page %d · %d issues ", m.filter.Page, len(m.issues)))
	}

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(pageInfo)
	if gap < 0 {
		gap = 0
	}

	return left + strings.Repeat(" ", gap) + pageInfo + "\n"
}

func (m Model) renderFooter() string {
	status := m.currentStatus()
	if status != "" {
		if m.statusIsErr {
			return styles.ErrorMsg.Render(status)
		}
		return styles.SuccessMsg.Render(status)
	}

	if m.loading {
		return styles.Spinner.Render(m.spinner.View()) + " Loading..."
	}

	switch m.currentView {
	case ViewIssueList:
		return styles.StatusBar.Render(
			styles.StatusKey.Render("↑↓") + " navigate  " +
				styles.StatusKey.Render("enter") + " detail  " +
				styles.StatusKey.Render("/") + " search  " +
				styles.StatusKey.Render("f") + " filter  " +
				styles.StatusKey.Render("R") + " resolve  " +
				styles.StatusKey.Render("?") + " help  " +
				styles.StatusKey.Render("q") + " quit")
	case ViewIssueDetail:
		return styles.StatusBar.Render(
			styles.StatusKey.Render("↑↓") + " scroll  " +
				styles.StatusKey.Render("esc") + " back  " +
				styles.StatusKey.Render("R") + " resolve  " +
				styles.StatusKey.Render("a") + " assign  " +
				styles.StatusKey.Render("?") + " help")
	default:
		return styles.StatusBar.Render(styles.StatusKey.Render("esc") + " back  " + styles.StatusKey.Render("?") + " help")
	}
}

// ---------- Search bar ----------

func (m Model) renderSearchBar() string {
	if m.searching {
		return styles.BorderedBox.Width(m.width - 4).Render(
			"🔍 " + m.searchInput.View(),
		) + "\n"
	}
	return ""
}
