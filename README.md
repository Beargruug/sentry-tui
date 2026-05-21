# 🔭 Sentry TUI

A feature-rich terminal user interface for monitoring and managing [Sentry](https://sentry.io) issues, built with Go and the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-blue)

---

### Features

- **📋 Issue List** — Browse recent errors/issues with level badges, event counts, user counts, and relative timestamps
- **🔍 Search** — Full-text search across issue titles and messages
- **🎛️ Filters** — Filter by status (unresolved/resolved/ignored), project, and sort order
- **📄 Issue Detail** — View full stack traces with source context, breadcrumbs, HTTP request info, tags, and metadata
- **✅ Resolve/Unresolve** — Mark issues as resolved directly from the terminal
- **🚫 Ignore** — Ignore noisy issues
- **👤 Assign** — Assign issues to team members with an interactive member picker
- **🔄 Auto-Refresh** — Configurable automatic polling for real-time updates
- **📑 Pagination** — Navigate through pages of issues with cursor-based pagination
- **⌨️ Vim-Style Keys** — Navigate with `j`/`k`, `g`/`G`, and more
- **⚙️ Configuration** — YAML config file + environment variable support + first-run setup wizard

---

### Installation

#### Prerequisites

- **Go 1.22+** installed ([download](https://go.dev/dl/))
- A **Sentry account** with an **Auth Token** ([create one](https://sentry.io/settings/auth-tokens/))

#### Build from source

```bash
git clone https://github.com/user/sentry-tui.git
cd sentry-tui
go mod tidy
go build -o sentry-tui .
```

#### Run directly

```bash
go run .
```

---

### Quick Start

#### 1. Get a Sentry Auth Token

1. Go to [Sentry → Settings → Auth Tokens](https://sentry.io/settings/auth-tokens/)
2. Click **Create New Token**
3. Grant scopes: `event:read`, `event:write`, `project:read`, `org:read`, `member:read`
4. Copy the token

#### 2. Run the app

```bash
./sentry-tui
```

On first run, the **setup wizard** will prompt you for:
- Auth token
- Organization slug (the slug from your Sentry URL, e.g., `my-org`)
- Default project (optional)

Configuration is saved to `~/.config/sentry-tui/sentry-tui.yaml`.

#### 3. Or use environment variables

```bash
export SENTRY_AUTH_TOKEN="sntrys_your_token_here"
export SENTRY_ORG="my-organization"
export SENTRY_PROJECT="my-project"  # optional

./sentry-tui
```

---

### Configuration

Configuration is stored in `~/.config/sentry-tui/sentry-tui.yaml`:

```yaml
auth_token: "sntrys_your_token_here"
organization: "my-org"
default_project: "my-project"     # optional
base_url: "https://sentry.io/api/0"  # change for self-hosted
refresh_seconds: 30
theme: "dark"
```

Environment variables override file values:

| Variable | Description |
|---|---|
| `SENTRY_AUTH_TOKEN` | Sentry auth token |
| `SENTRY_ORG` | Organization slug |
| `SENTRY_PROJECT` | Default project filter |
| `SENTRY_BASE_URL` | API base URL (for self-hosted Sentry) |

---

### Keyboard Shortcuts

#### Issue List

| Key | Action |
|---|---|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `g` | Go to top |
| `G` | Go to bottom |
| `enter` | Open issue detail |
| `n` / `ctrl+f` | Next page |
| `p` / `ctrl+b` | Previous page |
| `/` | Search issues |
| `f` | Open filter panel |
| `r` | Refresh |
| `R` | Resolve / unresolve |
| `i` | Ignore issue |
| `a` | Assign to member |
| `o` | Show permalink |

#### Issue Detail

| Key | Action |
|---|---|
| `↑` / `k` | Scroll up |
| `↓` / `j` | Scroll down |
| `g` | Scroll to top |
| `esc` | Back to list |
| `R` | Resolve / unresolve |
| `a` | Assign |
| `r` | Refresh detail |

#### General

| Key | Action |
|---|---|
| `?` | Toggle help |
| `C` | Open configuration |
| `q` / `ctrl+c` | Quit |

#### Filter Panel

| Key | Action |
|---|---|
| `1` | Unresolved only |
| `2` | Resolved only |
| `3` | Ignored only |
| `4` | All statuses |
| `5` | Cycle sort mode |
| `6` | Cycle project |
| `0` | Reset all filters |

---

### Project Structure

```
sentry-tui/
├── main.go                      # Entry point
├── go.mod
├── go.sum
├── README.md
└── internal/
    ├── api/
    │   └── client.go            # Sentry REST API client
    ├── config/
    │   └── config.go            # YAML config + env vars
    ├── models/
    │   └── models.go            # Sentry data types
    └── ui/
        ├── model.go             # Root Bubble Tea model
        ├── update.go            # Update logic & key handlers
        ├── view.go              # View dispatcher & shared rendering
        ├── view_issues.go       # Issue list view
        ├── view_detail.go       # Issue detail view (stack traces, breadcrumbs, tags)
        ├── view_help.go         # Help view
        ├── view_filter.go       # Filter panel view
        ├── view_assign.go       # Assign dialog view
        ├── view_setup.go        # First-run setup wizard
        ├── view_config.go       # Configuration editor view
        ├── messages.go          # Custom tea.Msg types
        ├── commands.go          # Async tea.Cmd functions
        ├── keys/
        │   └── keys.go          # Keybinding definitions
        └── styles/
            └── styles.go        # Lipgloss styles & theme
```

---

### Self-Hosted Sentry

For self-hosted Sentry instances, set the base URL:

```yaml
# ~/.config/sentry-tui/sentry-tui.yaml
base_url: "https://sentry.mycompany.com/api/0"
```

Or via environment variable:

```bash
export SENTRY_BASE_URL="https://sentry.mycompany.com/api/0"
```

---

### Development

```bash
# Run with live reload (requires air: go install github.com/air-verse/air@latest)
air

# Build for release
go build -ldflags="-s -w" -o sentry-tui .

# Cross-compile
GOOS=darwin GOARCH=arm64 go build -o sentry-tui-darwin-arm64 .
GOOS=linux GOARCH=amd64 go build -o sentry-tui-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o sentry-tui-windows-amd64.exe .
```

---

### Required Sentry Token Scopes

When creating your auth token, ensure these scopes are enabled:

| Scope | Used for |
|---|---|
| `event:read` | Fetching issues and events |
| `event:write` | Resolving, ignoring, assigning issues |
| `project:read` | Listing projects |
| `org:read` | Listing organizations |
| `member:read` | Listing team members for assignment |

---

### License

MIT
