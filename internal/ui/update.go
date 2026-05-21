package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Beargruug/sentry-tui/internal/api"
	"github.com/Beargruug/sentry-tui/internal/config"
	"github.com/Beargruug/sentry-tui/internal/models"
)

// Update handles all messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case TickMsg:
		// Auto-refresh only on the issue list view
		if m.currentView == ViewIssueList && m.client != nil && !m.loading {
			m.loading = true
			cmds = append(cmds, fetchIssues(m.client, m.filter))
		}
		cmds = append(cmds, tickCmd(m.refreshInterval))
		return m, tea.Batch(cmds...)

	case IssuesLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.setStatus(fmt.Sprintf("Error: %s", msg.Err), true)
		} else {
			m.issues = msg.Result.Issues
			m.pageCursor = msg.Result.Cursor
			if m.cursor >= len(m.issues) {
				m.cursor = max(0, len(m.issues)-1)
			}
		}
		return m, tea.Batch(cmds...)

	case IssueDetailLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.setStatus(fmt.Sprintf("Error loading detail: %s", msg.Err), true)
			return m, tea.Batch(cmds...)
		}
		m.detailIssue = msg.Issue
		m.detailEvent = msg.Event
		m.detailScroll = 0
		m.frameFolds = make(map[string]bool)
		m.frameCursor = 0
		m.frameNavMode = false
		m.currentView = ViewIssueDetail
		return m, tea.Batch(cmds...)

	case IssueEventLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.setStatus("Event data unavailable", true)
			return m, tea.Batch(cmds...)
		}
		m.detailEvent = msg.Event
		return m, tea.Batch(cmds...)

	case ProjectsLoadedMsg:
		if msg.Err == nil {
			m.projects = msg.Projects
			// Resolve ProjectID if Project slug is set (e.g. from DefaultProject config)
			if m.filter.Project != "" && m.filter.ProjectID == "" {
				for _, p := range m.projects {
					if p.Slug == m.filter.Project {
						m.filter.ProjectID = p.ID
						break
					}
				}
			}
		}
		return m, tea.Batch(cmds...)

	case MembersLoadedMsg:
		if msg.Err == nil {
			m.members = msg.Members
		}
		return m, tea.Batch(cmds...)

	case EnvironmentsLoadedMsg:
		if msg.Err == nil {
			m.environments = msg.Environments
		}
		return m, tea.Batch(cmds...)

	case ActionResultMsg:
		if msg.Err != nil {
			m.setStatus(fmt.Sprintf("%s failed: %s", msg.Action, msg.Err), true)
		} else {
			m.setStatus(fmt.Sprintf("Issue %s: %sd ✓", msg.IssueID, msg.Action), false)
			// Refresh issues
			cmds = append(cmds, fetchIssues(m.client, m.filter))
			if m.currentView == ViewIssueDetail {
				cmds = append(cmds, fetchIssueDetail(m.client, msg.IssueID))
			}
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) setStatus(text string, isErr bool) {
	m.statusMsg = text
	m.statusIsErr = isErr
	m.statusExpiry = time.Now().Add(5 * time.Second)
}

func (m *Model) currentStatus() string {
	if time.Now().After(m.statusExpiry) {
		return ""
	}
	return m.statusMsg
}

// ---------- Key handler dispatch ----------

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If searching, delegate to search input first
	if m.searching {
		return m.handleSearchKey(msg)
	}

	// If assigning
	if m.currentView == ViewAssign {
		return m.handleAssignKey(msg)
	}

	// If selecting project
	if m.currentView == ViewProjectSelect {
		return m.handleProjectSelectKey(msg)
	}

	// If selecting environment
	if m.currentView == ViewEnvSelect {
		return m.handleEnvSelectKey(msg)
	}

	// If in setup wizard
	if m.currentView == ViewSetup {
		return m.handleSetupKey(msg)
	}

	// If in config view
	if m.currentView == ViewConfig {
		return m.handleConfigKey(msg)
	}

	// Global keys
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Help):
		if m.currentView == ViewHelp {
			m.currentView = m.prevView
		} else {
			m.prevView = m.currentView
			m.currentView = ViewHelp
		}
		return m, nil

	case key.Matches(msg, m.keys.Config):
		if m.currentView == ViewConfig {
			m.currentView = ViewIssueList
		} else {
			m.configInputs = makeConfigInputs(m.cfg)
			m.configCursor = 0
			m.configInputs[0].Focus()
			m.currentView = ViewConfig
		}
		return m, nil
	}

	// View-specific keys
	switch m.currentView {
	case ViewIssueList:
		return m.handleIssueListKey(msg)
	case ViewIssueDetail:
		return m.handleIssueDetailKey(msg)
	case ViewHelp:
		return m.handleHelpKey(msg)
	case ViewFilter:
		return m.handleFilterKey(msg)
	}

	return m, nil
}

// ---------- Issue List keys ----------

func (m Model) handleIssueListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.issues)-1 {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.GotoTop):
		m.cursor = 0
	case key.Matches(msg, m.keys.GotoBot):
		if len(m.issues) > 0 {
			m.cursor = len(m.issues) - 1
		}
	case key.Matches(msg, m.keys.Enter):
		if len(m.issues) > 0 && m.cursor < len(m.issues) {
			// Show detail immediately with list data, fetch event in background
			m.detailIssue = m.issues[m.cursor]
			m.detailEvent = models.Event{}
			m.detailScroll = 0
			m.frameFolds = make(map[string]bool)
			m.frameCursor = 0
			m.frameNavMode = false
			m.currentView = ViewIssueDetail
			m.loading = true
			return m, fetchIssueEvent(m.client, m.issues[m.cursor].ID)
		}
	case key.Matches(msg, m.keys.Search):
		m.searching = true
		m.searchInput.Focus()
		return m, textinput.Blink
	case key.Matches(msg, m.keys.Filter):
		m.prevView = m.currentView
		m.currentView = ViewFilter
		return m, nil
	case key.Matches(msg, m.keys.Refresh):
		m.loading = true
		return m, fetchIssues(m.client, m.filter)
	case key.Matches(msg, m.keys.Resolve):
		if len(m.issues) > 0 && m.cursor < len(m.issues) {
			issue := m.issues[m.cursor]
			if issue.Status == "resolved" {
				return m, unresolveIssue(m.client, issue.ID)
			}
			return m, resolveIssue(m.client, issue.ID)
		}
	case key.Matches(msg, m.keys.Ignore):
		if len(m.issues) > 0 && m.cursor < len(m.issues) {
			return m, ignoreIssue(m.client, m.issues[m.cursor].ID)
		}
	case key.Matches(msg, m.keys.Assign):
		if len(m.issues) > 0 && m.cursor < len(m.issues) {
			m.currentView = ViewAssign
			m.assignInput.SetValue("")
			m.assignInput.Focus()
			m.assignCursor = 0
			return m, textinput.Blink
		}
	case key.Matches(msg, m.keys.NextPage):
		if m.pageCursor.HasNext {
			m.filter.Cursor = m.pageCursor.NextCursor
			m.filter.Page++
			m.loading = true
			m.cursor = 0
			return m, fetchIssues(m.client, m.filter)
		}
	case key.Matches(msg, m.keys.PrevPage):
		if m.pageCursor.HasPrev {
			m.filter.Cursor = m.pageCursor.PrevCursor
			if m.filter.Page > 1 {
				m.filter.Page--
			}
			m.loading = true
			m.cursor = 0
			return m, fetchIssues(m.client, m.filter)
		}
	case key.Matches(msg, m.keys.Open):
		if len(m.issues) > 0 && m.cursor < len(m.issues) && m.issues[m.cursor].Permalink != "" {
			// Open in browser — we just set a status msg since we can't actually launch browser from TUI easily
			m.setStatus("Link: "+m.issues[m.cursor].Permalink, false)
		}
	}
	return m, nil
}

// ---------- Issue Detail keys ----------

func (m Model) handleIssueDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		if m.frameNavMode {
			m.frameNavMode = false
			return m, nil
		}
		m.currentView = ViewIssueList
		return m, nil
	case msg.String() == "tab":
		m.frameNavMode = !m.frameNavMode
		return m, nil
	case m.frameNavMode && (msg.String() == "j" || msg.Type == tea.KeyDown):
		m.frameCursor++
		return m, nil
	case m.frameNavMode && (msg.String() == "k" || msg.Type == tea.KeyUp):
		if m.frameCursor > 0 {
			m.frameCursor--
		}
		return m, nil
	case m.frameNavMode && (msg.String() == " " || msg.Type == tea.KeyEnter):
		if m.frameFolds == nil {
			m.frameFolds = make(map[string]bool)
		}
		key := fmt.Sprintf("%d", m.frameCursor)
		m.frameFolds[key] = !m.frameFolds[key]
		return m, nil
	case key.Matches(msg, m.keys.Down):
		m.detailScroll++
	case key.Matches(msg, m.keys.Up):
		if m.detailScroll > 0 {
			m.detailScroll--
		}
	case msg.String() == "d":
		m.detailScroll += (m.height - 3) / 2
	case msg.String() == "u":
		m.detailScroll -= (m.height - 3) / 2
		if m.detailScroll < 0 {
			m.detailScroll = 0
		}
	case key.Matches(msg, m.keys.GotoTop):
		m.detailScroll = 0
	case key.Matches(msg, m.keys.GotoBot):
		m.detailScroll = 99999
	case key.Matches(msg, m.keys.Resolve):
		if m.detailIssue.Status == "resolved" {
			return m, unresolveIssue(m.client, m.detailIssue.ID)
		}
		return m, resolveIssue(m.client, m.detailIssue.ID)
	case key.Matches(msg, m.keys.Ignore):
		return m, ignoreIssue(m.client, m.detailIssue.ID)
	case key.Matches(msg, m.keys.Assign):
		m.currentView = ViewAssign
		m.assignInput.SetValue("")
		m.assignInput.Focus()
		return m, textinput.Blink
	case key.Matches(msg, m.keys.Open):
		if m.detailIssue.Permalink != "" {
			m.setStatus("Link: "+m.detailIssue.Permalink, false)
		}
	case key.Matches(msg, m.keys.Refresh):
		m.loading = true
		return m, fetchIssueDetail(m.client, m.detailIssue.ID)
	}
	return m, nil
}

// ---------- Search input keys ----------

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.searching = false
		m.searchInput.Blur()
		return m, nil
	case tea.KeyEnter:
		m.searching = false
		m.searchInput.Blur()
		m.filter.Query = m.searchInput.Value()
		m.filter.Cursor = ""
		m.filter.Page = 1
		m.cursor = 0
		m.loading = true
		return m, fetchIssues(m.client, m.filter)
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

// ---------- Assign input keys ----------

func (m Model) handleAssignKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.currentView = ViewIssueList
		m.assignInput.Blur()
		return m, nil
	case tea.KeyEnter:
		assignee := m.assignInput.Value()
		if assignee == "" && m.assignCursor < len(m.members) {
			assignee = m.members[m.assignCursor].Email
		}
		if assignee != "" {
			issueID := ""
			if m.currentView == ViewAssign && len(m.issues) > 0 && m.cursor < len(m.issues) {
				issueID = m.issues[m.cursor].ID
			}
			if m.detailIssue.ID != "" {
				issueID = m.detailIssue.ID
			}
			if issueID != "" {
				m.currentView = ViewIssueList
				m.assignInput.Blur()
				return m, assignIssue(m.client, issueID, assignee)
			}
		}
		m.currentView = ViewIssueList
		return m, nil
	case tea.KeyUp:
		if m.assignCursor > 0 {
			m.assignCursor--
		}
		return m, nil
	case tea.KeyDown:
		if m.assignCursor < len(m.members)-1 {
			m.assignCursor++
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.assignInput, cmd = m.assignInput.Update(msg)
	return m, cmd
}

// ---------- Filter view keys ----------

func (m Model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape), key.Matches(msg, m.keys.Back):
		m.currentView = ViewIssueList
		return m, nil

	case msg.String() == "1":
		m.filter.Status = "unresolved"
		m.filter.Cursor = ""
		m.filter.Page = 1
		m.cursor = 0
		m.loading = true
		m.currentView = ViewIssueList
		return m, fetchIssues(m.client, m.filter)

	case msg.String() == "2":
		m.filter.Status = "resolved"
		m.filter.Cursor = ""
		m.filter.Page = 1
		m.cursor = 0
		m.loading = true
		m.currentView = ViewIssueList
		return m, fetchIssues(m.client, m.filter)

	case msg.String() == "3":
		m.filter.Status = "ignored"
		m.filter.Cursor = ""
		m.filter.Page = 1
		m.cursor = 0
		m.loading = true
		m.currentView = ViewIssueList
		return m, fetchIssues(m.client, m.filter)

	case msg.String() == "4":
		m.filter.Status = ""
		m.filter.Cursor = ""
		m.filter.Page = 1
		m.cursor = 0
		m.loading = true
		m.currentView = ViewIssueList
		return m, fetchIssues(m.client, m.filter)

	case msg.String() == "5":
		// Cycle through sort options
		sorts := []string{"date", "new", "priority", "freq", "user"}
		for i, s := range sorts {
			if s == m.filter.Sort {
				m.filter.Sort = sorts[(i+1)%len(sorts)]
				break
			}
		}
		m.filter.Cursor = ""
		m.filter.Page = 1
		m.cursor = 0
		m.loading = true
		m.currentView = ViewIssueList
		return m, fetchIssues(m.client, m.filter)

	case msg.String() == "6":
		// Open project selector
		m.projectSelectCursor = 0
		m.projectSelectInput.SetValue("")
		m.projectSelectInput.Focus()
		m.currentView = ViewProjectSelect
		return m, textinput.Blink

	case msg.String() == "7":
		// Open environment selector
		m.envSelectCursor = 0
		m.envSelectInput.SetValue("")
		m.envSelectInput.Focus()
		m.currentView = ViewEnvSelect
		return m, textinput.Blink

	case msg.String() == "0":
		// Reset filters
		m.filter = models.DefaultFilter()
		m.searchInput.SetValue("")
		m.cursor = 0
		m.loading = true
		m.currentView = ViewIssueList
		return m, fetchIssues(m.client, m.filter)
	}
	return m, nil
}

// ---------- Project Select keys ----------

func (m Model) handleProjectSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	filtered := m.filteredProjects()
	maxIdx := len(filtered) // 0 = "all", 1..N = projects

	switch msg.Type {
	case tea.KeyEsc:
		m.currentView = ViewFilter
		m.projectSelectInput.Blur()
		return m, nil
	case tea.KeyEnter:
		if m.projectSelectCursor == 0 {
			// "All projects"
			m.filter.Project = ""
			m.filter.ProjectID = ""
		} else {
			idx := m.projectSelectCursor - 1
			if idx < len(filtered) {
				m.filter.Project = filtered[idx].Slug
				m.filter.ProjectID = filtered[idx].ID
			}
		}
		m.filter.Cursor = ""
		m.filter.Page = 1
		m.cursor = 0
		m.loading = true
		m.currentView = ViewIssueList
		m.projectSelectInput.Blur()
		return m, fetchIssues(m.client, m.filter)
	case tea.KeyUp:
		if m.projectSelectCursor > 0 {
			m.projectSelectCursor--
		}
		return m, nil
	case tea.KeyDown:
		if m.projectSelectCursor < maxIdx {
			m.projectSelectCursor++
		}
		return m, nil
	}

	// Pass to text input for filtering
	prevValue := m.projectSelectInput.Value()
	var cmd tea.Cmd
	m.projectSelectInput, cmd = m.projectSelectInput.Update(msg)
	// Reset cursor when search changes
	if m.projectSelectInput.Value() != prevValue {
		m.projectSelectCursor = 0
	}
	return m, cmd
}

// ---------- Environment Select keys ----------

func (m Model) handleEnvSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	filtered := m.filteredEnvironments()
	maxIdx := len(filtered) // 0 = "all", 1..N = environments

	switch msg.Type {
	case tea.KeyEsc:
		m.currentView = ViewFilter
		m.envSelectInput.Blur()
		return m, nil
	case tea.KeyEnter:
		if m.envSelectCursor == 0 {
			m.filter.Environment = ""
		} else {
			idx := m.envSelectCursor - 1
			if idx < len(filtered) {
				m.filter.Environment = filtered[idx].Name
			}
		}
		m.filter.Cursor = ""
		m.filter.Page = 1
		m.cursor = 0
		m.loading = true
		m.currentView = ViewIssueList
		m.envSelectInput.Blur()
		return m, fetchIssues(m.client, m.filter)
	case tea.KeyUp:
		if m.envSelectCursor > 0 {
			m.envSelectCursor--
		}
		return m, nil
	case tea.KeyDown:
		if m.envSelectCursor < maxIdx {
			m.envSelectCursor++
		}
		return m, nil
	}

	prevValue := m.envSelectInput.Value()
	var cmd tea.Cmd
	m.envSelectInput, cmd = m.envSelectInput.Update(msg)
	if m.envSelectInput.Value() != prevValue {
		m.envSelectCursor = 0
	}
	return m, cmd
}

// ---------- Help view keys ----------

func (m Model) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back), key.Matches(msg, m.keys.Help):
		m.currentView = m.prevView
	}
	return m, nil
}

// ---------- Setup wizard keys ----------

func (m Model) handleSetupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	idx := int(m.setupStep)
	switch msg.Type {
	case tea.KeyEnter:
		switch m.setupStep {
		case StepToken:
			if m.setupInputs[0].Value() != "" {
				m.setupStep = StepOrg
				m.setupInputs[0].Blur()
				m.setupInputs[1].Focus()
				return m, textinput.Blink
			}
		case StepOrg:
			if m.setupInputs[1].Value() != "" {
				m.setupStep = StepProject
				m.setupInputs[1].Blur()
				m.setupInputs[2].Focus()
				return m, textinput.Blink
			}
		case StepProject:
			// Save config
			m.cfg.AuthToken = m.setupInputs[0].Value()
			m.cfg.Organization = m.setupInputs[1].Value()
			m.cfg.DefaultProject = m.setupInputs[2].Value()
			if m.cfg.BaseURL == "" {
				m.cfg.BaseURL = "https://sentry.io/api/0"
			}
			if m.cfg.RefreshSeconds == 0 {
				m.cfg.RefreshSeconds = 30
			}
			_ = config.Save(m.cfg)

			m.client = api.NewClient(m.cfg.BaseURL, m.cfg.AuthToken, m.cfg.Organization)
			m.currentView = ViewIssueList
			m.loading = true
			m.refreshInterval = time.Duration(m.cfg.RefreshSeconds) * time.Second
			if m.cfg.DefaultProject != "" {
				m.filter.Project = m.cfg.DefaultProject
				// ProjectID will be resolved when projects are loaded
				for _, p := range m.projects {
					if p.Slug == m.cfg.DefaultProject {
						m.filter.ProjectID = p.ID
						break
					}
				}
			}
			return m, tea.Batch(
				fetchIssues(m.client, m.filter),
				fetchProjects(m.client),
				fetchMembers(m.client),
				fetchEnvironments(m.client),
				tickCmd(m.refreshInterval),
			)
		}
		return m, nil
	case tea.KeyEsc:
		return m, tea.Quit
	}

	if idx < len(m.setupInputs) {
		var cmd tea.Cmd
		m.setupInputs[idx], cmd = m.setupInputs[idx].Update(msg)
		return m, cmd
	}
	return m, nil
}

// ---------- Config view keys ----------

func (m Model) handleConfigKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.currentView = ViewIssueList
		return m, nil
	case tea.KeyTab, tea.KeyDown:
		m.configInputs[m.configCursor].Blur()
		m.configCursor = (m.configCursor + 1) % len(m.configInputs)
		m.configInputs[m.configCursor].Focus()
		return m, textinput.Blink
	case tea.KeyShiftTab, tea.KeyUp:
		m.configInputs[m.configCursor].Blur()
		m.configCursor--
		if m.configCursor < 0 {
			m.configCursor = len(m.configInputs) - 1
		}
		m.configInputs[m.configCursor].Focus()
		return m, textinput.Blink
	case tea.KeyEnter:
		// Save
		m.cfg.AuthToken = m.configInputs[0].Value()
		m.cfg.Organization = m.configInputs[1].Value()
		m.cfg.DefaultProject = m.configInputs[2].Value()
		if rs := m.configInputs[3].Value(); rs != "" {
			if v, err := strconv.Atoi(rs); err == nil && v > 0 {
				m.cfg.RefreshSeconds = v
			}
		}
		_ = config.Save(m.cfg)
		m.client = api.NewClient(m.cfg.BaseURL, m.cfg.AuthToken, m.cfg.Organization)
		m.refreshInterval = time.Duration(m.cfg.RefreshSeconds) * time.Second
		m.setStatus("Configuration saved ✓", false)
		m.currentView = ViewIssueList
		m.loading = true
		return m, tea.Batch(
			fetchIssues(m.client, m.filter),
			fetchProjects(m.client),
			fetchMembers(m.client),
			fetchEnvironments(m.client),
		)
	}

	var cmd tea.Cmd
	m.configInputs[m.configCursor], cmd = m.configInputs[m.configCursor].Update(msg)
	return m, cmd
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
