// Package ui defines the Bubble Tea model, views, and update logic.
package ui

import (
	"github.com/Beargruug/sentry-tui/internal/api"
	"github.com/Beargruug/sentry-tui/internal/models"
)

// ---------- Custom messages (tea.Msg) ----------

// IssuesLoadedMsg carries a fetched page of issues.
type IssuesLoadedMsg struct {
	Result api.IssueResult
	Err    error
}

// IssueDetailLoadedMsg carries a single issue + latest event.
type IssueDetailLoadedMsg struct {
	Issue models.Issue
	Event models.Event
	Err   error
}

// ProjectsLoadedMsg carries the project list.
type ProjectsLoadedMsg struct {
	Projects []models.Project
	Err      error
}

// MembersLoadedMsg carries the member list.
type MembersLoadedMsg struct {
	Members []models.Member
	Err     error
}

// EnvironmentsLoadedMsg carries the environment list.
type EnvironmentsLoadedMsg struct {
	Environments []models.Environment
	Err          error
}

// ActionResultMsg carries the result of an action (resolve, assign, etc.).
type ActionResultMsg struct {
	Action  string // "resolve", "unresolve", "ignore", "assign"
	IssueID string
	Success bool
	Err     error
}

// TickMsg triggers auto-refresh.
type TickMsg struct{}

// StatusMsg sets a temporary status message.
type StatusMsg struct {
	Text    string
	IsError bool
}
