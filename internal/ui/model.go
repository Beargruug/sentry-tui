package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Beargruug/sentry-tui/internal/api"
	"github.com/Beargruug/sentry-tui/internal/config"
	"github.com/Beargruug/sentry-tui/internal/models"
	"github.com/Beargruug/sentry-tui/internal/ui/keys"
)

// View identifies the current screen.
type View int

const (
	ViewSetup View = iota
	ViewIssueList
	ViewIssueDetail
	ViewHelp
	ViewFilter
	ViewAssign
	ViewConfig
	ViewProjectSelect
	ViewEnvSelect
)

// SetupStep tracks the wizard step.
type SetupStep int

const (
	StepToken SetupStep = iota
	StepOrg
	StepProject
	StepDone
)

// Model is the root Bubble Tea model.
type Model struct {
	// State
	currentView View
	prevView    View
	width       int
	height      int
	ready       bool

	// Config
	cfg    config.Config
	client *api.Client
	keys   keys.KeyMap

	// Issues list
	issues       []models.Issue
	cursor       int
	filter       models.FilterState
	pageCursor   models.IssueListCursor
	loading      bool
	statusMsg    string
	statusIsErr  bool
	statusExpiry time.Time

	// Issue detail
	detailIssue  models.Issue
	detailEvent  models.Event
	detailScroll int
	// Stack trace fold state: map of "exceptionIdx:frameIdx" -> expanded
	frameFolds    map[string]bool
	frameCursor   int  // cursor position in the stack trace frames
	frameNavMode  bool // whether we're navigating frames

	// Projects & members caches
	projects     []models.Project
	members      []models.Member
	environments []models.Environment

	// Search / filter input
	searchInput textinput.Model
	searching   bool

	// Assign input
	assignInput    textinput.Model
	assignCursor   int

	// Project selector
	projectSelectInput  textinput.Model
	projectSelectCursor int

	// Environment selector
	envSelectInput  textinput.Model
	envSelectCursor int

	// Setup wizard
	setupStep    SetupStep
	setupInputs  []textinput.Model

	// Config view
	configInputs []textinput.Model
	configCursor int

	// Widgets
	spinner spinner.Model

	// Auto-refresh
	refreshInterval time.Duration
}

// NewModel creates a new root Model.
func NewModel(cfg config.Config) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	si := textinput.New()
	si.Placeholder = "Search issues..."
	si.CharLimit = 256

	ai := textinput.New()
	ai.Placeholder = "user email or username"
	ai.CharLimit = 256

	pi := textinput.New()
	pi.Placeholder = "Filter projects..."
	pi.CharLimit = 128

	ei := textinput.New()
	ei.Placeholder = "Filter environments..."
	ei.CharLimit = 128

	m := Model{
		cfg:                cfg,
		keys:               keys.DefaultKeyMap(),
		filter:             models.DefaultFilter(),
		spinner:            sp,
		searchInput:        si,
		assignInput:        ai,
		projectSelectInput: pi,
		envSelectInput:     ei,
		refreshInterval:    time.Duration(cfg.RefreshSeconds) * time.Second,
	}

	if cfg.DefaultProject != "" {
		m.filter.Project = cfg.DefaultProject
	}

	if cfg.IsValid() {
		m.client = api.NewClient(cfg.BaseURL, cfg.AuthToken, cfg.Organization)
		m.currentView = ViewIssueList
	} else {
		m.currentView = ViewSetup
		m.setupStep = StepToken
		m.setupInputs = makeSetupInputs(cfg)
	}

	return m
}

func makeSetupInputs(cfg config.Config) []textinput.Model {
	inputs := make([]textinput.Model, 3)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "sentry auth token (sntrys_...)"
	inputs[0].CharLimit = 256
	inputs[0].Focus()
	if cfg.AuthToken != "" {
		inputs[0].SetValue(cfg.AuthToken)
	}

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "organization slug"
	inputs[1].CharLimit = 128
	if cfg.Organization != "" {
		inputs[1].SetValue(cfg.Organization)
	}

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "default project slug (optional)"
	inputs[2].CharLimit = 128
	if cfg.DefaultProject != "" {
		inputs[2].SetValue(cfg.DefaultProject)
	}

	return inputs
}

func makeConfigInputs(cfg config.Config) []textinput.Model {
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "auth token"
	inputs[0].SetValue(cfg.AuthToken)
	inputs[0].CharLimit = 256

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "organization slug"
	inputs[1].SetValue(cfg.Organization)
	inputs[1].CharLimit = 128

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "default project (optional)"
	inputs[2].SetValue(cfg.DefaultProject)
	inputs[2].CharLimit = 128

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "refresh seconds"
	inputs[3].SetValue(fmt.Sprintf("%d", cfg.RefreshSeconds))
	inputs[3].CharLimit = 4

	return inputs
}

// Init starts the Bubble Tea program.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.spinner.Tick}

	if m.currentView == ViewIssueList && m.client != nil {
		m.loading = true
		cmds = append(cmds,
			fetchIssues(m.client, m.filter),
			fetchProjects(m.client),
			fetchMembers(m.client),
			fetchEnvironments(m.client),
			tickCmd(m.refreshInterval),
		)
	}
	return tea.Batch(cmds...)
}
