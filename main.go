// Sentry TUI — a terminal user interface for monitoring and managing Sentry issues.
//
// Usage:
//
//	sentry-tui              Run the TUI (launches setup wizard on first run)
//	sentry-tui --version    Print version
//	sentry-tui --help       Print help
package main

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Beargruug/sentry-tui/internal/config"
	"github.com/Beargruug/sentry-tui/internal/ui"
)

var version = "0.1.0"

func main() {
	// Simple arg handling
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("sentry-tui v%s\n", version)
			os.Exit(0)
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		case "update":
			runUpdate()
			os.Exit(0)
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load config: %s\n", err)
		cfg = config.DefaultConfig()
	}

	// Create the Bubble Tea model
	model := ui.NewModel(cfg)

	// Run the program
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running sentry-tui: %s\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf(`sentry-tui v%s — Terminal UI for Sentry

USAGE:
  sentry-tui              Launch the TUI
  sentry-tui update       Update to latest version
  sentry-tui --version    Print version
  sentry-tui --help       Print this help

CONFIGURATION:
  Config file: ~/.config/sentry-tui/sentry-tui.yaml

  On first run, a setup wizard will guide you through configuration.
  You can also set environment variables:

    SENTRY_AUTH_TOKEN    Your Sentry auth token
    SENTRY_ORG           Organization slug
    SENTRY_PROJECT       Default project slug (optional)
    SENTRY_BASE_URL      API base URL (default: https://sentry.io/api/0)

KEYBOARD SHORTCUTS:
  ↑/k, ↓/j       Navigate issues
  enter           View issue detail
  esc             Go back
  /               Search
  f               Filter panel
  R               Resolve/unresolve
  a               Assign issue
  i               Ignore issue
  r               Refresh
  n/p             Next/previous page
  ?               Help
  C               Configuration
  q               Quit
`, version)
}

func runUpdate() {
	fmt.Println("Updating sentry-tui...")
	cmd := exec.Command("bash", "-c",
		"curl -fsSL https://raw.githubusercontent.com/Beargruug/sentry-tui/main/install.sh | bash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
