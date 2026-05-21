package ui

import (
	"fmt"
	"time"

	"github.com/Beargruug/sentry-tui/internal/api"
	"github.com/Beargruug/sentry-tui/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// ---------- Tea commands that perform API calls ----------

func fetchIssues(client *api.Client, filter models.FilterState) tea.Cmd {
	return func() tea.Msg {
		result, err := client.ListIssues(filter)
		return IssuesLoadedMsg{Result: result, Err: err}
	}
}

func fetchIssueDetail(client *api.Client, issueID string) tea.Cmd {
	return func() tea.Msg {
		issue, err := client.GetIssue(issueID)
		if err != nil {
			return IssueDetailLoadedMsg{Err: fmt.Errorf("GetIssue(id=%s): %w", issueID, err)}
		}
		event, err := client.GetLatestEvent(issueID)
		if err != nil {
			return IssueDetailLoadedMsg{Issue: issue, Err: nil}
		}
		return IssueDetailLoadedMsg{Issue: issue, Event: event}
	}
}

func fetchIssueEvent(client *api.Client, issueID string) tea.Cmd {
	return func() tea.Msg {
		event, err := client.GetLatestEvent(issueID)
		if err != nil {
			return IssueEventLoadedMsg{Err: err}
		}
		return IssueEventLoadedMsg{Event: event}
	}
}

func fetchProjects(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		projects, err := client.ListProjects()
		return ProjectsLoadedMsg{Projects: projects, Err: err}
	}
}

func fetchMembers(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		members, err := client.ListMembers()
		return MembersLoadedMsg{Members: members, Err: err}
	}
}

func fetchEnvironments(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		envs, err := client.ListEnvironments()
		return EnvironmentsLoadedMsg{Environments: envs, Err: err}
	}
}

func resolveIssue(client *api.Client, issueID string) tea.Cmd {
	return func() tea.Msg {
		err := client.ResolveIssue(issueID)
		return ActionResultMsg{Action: "resolve", IssueID: issueID, Success: err == nil, Err: err}
	}
}

func unresolveIssue(client *api.Client, issueID string) tea.Cmd {
	return func() tea.Msg {
		err := client.UnresolveIssue(issueID)
		return ActionResultMsg{Action: "unresolve", IssueID: issueID, Success: err == nil, Err: err}
	}
}

func ignoreIssue(client *api.Client, issueID string) tea.Cmd {
	return func() tea.Msg {
		err := client.IgnoreIssue(issueID)
		return ActionResultMsg{Action: "ignore", IssueID: issueID, Success: err == nil, Err: err}
	}
}

func assignIssue(client *api.Client, issueID, assignee string) tea.Cmd {
	return func() tea.Msg {
		err := client.AssignIssue(issueID, assignee)
		return ActionResultMsg{Action: "assign", IssueID: issueID, Success: err == nil, Err: err}
	}
}

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}
