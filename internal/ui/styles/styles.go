// Package styles defines lipgloss styles used across the TUI.
package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	Primary     = lipgloss.Color("#7c3aed") // Purple – Sentry brand
	Secondary   = lipgloss.Color("#6366f1")
	Accent      = lipgloss.Color("#22d3ee")
	Success     = lipgloss.Color("#22c55e")
	Warning     = lipgloss.Color("#eab308")
	Error       = lipgloss.Color("#ef4444")
	Fatal       = lipgloss.Color("#dc2626")
	Muted       = lipgloss.Color("#6b7280")
	Subtle      = lipgloss.Color("#374151")
	Text        = lipgloss.Color("#e5e7eb")
	BrightText  = lipgloss.Color("#f9fafb")
	DimText     = lipgloss.Color("#9ca3af")
	Background  = lipgloss.Color("#111827")
	Surface     = lipgloss.Color("#1f2937")

	// Base styles
	App = lipgloss.NewStyle().
		Background(Background)

	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(BrightText).
		Background(Primary).
		Padding(0, 1)

	Subtitle = lipgloss.NewStyle().
		Foreground(DimText).
		Italic(true)

	// Status bar at the bottom
	StatusBar = lipgloss.NewStyle().
		Foreground(Text).
		Background(Surface).
		Padding(0, 1)

	StatusKey = lipgloss.NewStyle().
		Bold(true).
		Foreground(Accent).
		Background(Surface).
		Padding(0, 1)

	// Issue level badges
	LevelError = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#fff")).
		Background(Error).
		Padding(0, 1)

	LevelWarning = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#000")).
		Background(Warning).
		Padding(0, 1)

	LevelInfo = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#fff")).
		Background(Secondary).
		Padding(0, 1)

	LevelFatal = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#fff")).
		Background(Fatal).
		Padding(0, 1)

	LevelDebug = lipgloss.NewStyle().
		Foreground(DimText).
		Padding(0, 1)

	// Issue list
	SelectedItem = lipgloss.NewStyle().
		Bold(true).
		Foreground(BrightText).
		Background(Subtle).
		Padding(0, 1)

	NormalItem = lipgloss.NewStyle().
		Foreground(Text).
		Padding(0, 1)

	// Detail view
	DetailHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(BrightText).
		Background(Subtle).
		Padding(0, 1).
		MarginBottom(1)

	DetailLabel = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary).
		Width(14)

	DetailValue = lipgloss.NewStyle().
		Foreground(Text)

	// Code styles - clean editor look (Rosé Pine Moon palette)
	// The goal: plain readable code, only the error line stands out

	// Normal code context line - just dimmed, no background
	CodeContextLine = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#908caa")) // muted/subtle

	// The critical/error line - normal brightness, stands out by contrast
	CodeCriticalLine = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0def4")) // full brightness text

	// Line numbers - dim
	CodeLineNo = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6e6a86")) // muted

	// Critical line number - highlighted
	CodeCriticalLineNo = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#eb6f92")). // love/rose
		Bold(true)

	// Gutter marker for the critical line (left border indicator)
	CodeCriticalGutter = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#eb6f92")). // love/rose
		Bold(true)

	// Unused but kept for compatibility
	CodeBlock = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0def4"))
	CodeGutter = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#393552"))

	StackFrameApp = lipgloss.NewStyle().
		Foreground(BrightText).
		Bold(true)

	StackFrameLib = lipgloss.NewStyle().
		Foreground(DimText)

	// Tags
	TagKey = lipgloss.NewStyle().
		Foreground(Accent).
		Bold(true)

	TagValue = lipgloss.NewStyle().
		Foreground(Text)

	// Help
	HelpKey = lipgloss.NewStyle().
		Bold(true).
		Foreground(Accent)

	HelpDesc = lipgloss.NewStyle().
		Foreground(DimText)

	// Borders
	BorderedBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Subtle).
		Padding(1, 2)

	// Tab
	ActiveTab = lipgloss.NewStyle().
		Bold(true).
		Foreground(BrightText).
		Background(Primary).
		Padding(0, 2)

	InactiveTab = lipgloss.NewStyle().
		Foreground(DimText).
		Background(Surface).
		Padding(0, 2)

	// Spinner / loading
	Spinner = lipgloss.NewStyle().
		Foreground(Primary)

	// Error message
	ErrorMsg = lipgloss.NewStyle().
		Foreground(Error).
		Bold(true).
		Padding(0, 1)

	// Success message
	SuccessMsg = lipgloss.NewStyle().
		Foreground(Success).
		Bold(true).
		Padding(0, 1)
)

// LevelStyle returns the style for a Sentry issue level.
func LevelStyle(level string) lipgloss.Style {
	switch level {
	case "fatal":
		return LevelFatal
	case "error":
		return LevelError
	case "warning":
		return LevelWarning
	case "info":
		return LevelInfo
	default:
		return LevelDebug
	}
}

// SectionHeader renders a full-width section header with background.
func SectionHeader(title string, width int) string {
	padding := width - len(title) - 4
	if padding < 0 {
		padding = 0
	}
	content := " " + title + " " + strings.Repeat("─", padding)
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(BrightText).
		Background(Subtle).
		Padding(0, 1).
		Render(content)
}
