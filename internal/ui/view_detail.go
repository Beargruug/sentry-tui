package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/Beargruug/sentry-tui/internal/models"
	"github.com/Beargruug/sentry-tui/internal/ui/styles"
)

func (m Model) viewIssueDetail() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())

	issue := m.detailIssue
	halfWidth := m.width / 2
	if halfWidth < 40 {
		halfWidth = 40
	}

	// Title bar with level badge
	levelBadge := styles.LevelStyle(issue.Level).Render(" " + strings.ToUpper(issue.Level) + " ")
	statusBadge := renderStatusBadge(issue.Status)
	b.WriteString("\n " + levelBadge + " " + statusBadge + "  " + styles.DetailHeader.Render(issue.ShortID) + "\n")
	b.WriteString(" " + lipgloss.NewStyle().Bold(true).Foreground(styles.BrightText).Render(issue.Title) + "\n")
	if issue.Culprit != "" {
		b.WriteString(" " + styles.Subtitle.Render(issue.Culprit) + "\n")
	}
	b.WriteString("\n")

	// Two-column metadata
	leftCol := []struct{ label, value string }{
		{"Project", issue.Project.Name},
		{"Platform", issue.Platform},
		{"First Seen", issue.FirstSeen.Format("Jan 02 15:04")},
		{"Last Seen", issue.LastSeen.Format("Jan 02 15:04")},
	}
	rightCol := []struct{ label, value string }{
		{"Events", issue.Count},
		{"Users", fmt.Sprintf("%d", issue.UserCount)},
	}
	if issue.AssignedTo != nil {
		name := issue.AssignedTo.Name
		if name == "" {
			name = issue.AssignedTo.Email
		}
		rightCol = append(rightCol, struct{ label, value string }{"Assigned", name})
	} else {
		rightCol = append(rightCol, struct{ label, value string }{"Assigned", styles.Subtitle.Render("unassigned")})
	}
	if issue.Logger != "" {
		rightCol = append(rightCol, struct{ label, value string }{"Logger", issue.Logger})
	}

	// Render columns
	maxRows := len(leftCol)
	if len(rightCol) > maxRows {
		maxRows = len(rightCol)
	}
	colWidth := 38
	for i := 0; i < maxRows; i++ {
		left := ""
		if i < len(leftCol) {
			left = fmt.Sprintf("  %s %s",
				styles.DetailLabel.Render(padRight(leftCol[i].label+":", 12)),
				styles.DetailValue.Render(leftCol[i].value))
		}
		right := ""
		if i < len(rightCol) {
			right = fmt.Sprintf("%s %s",
				styles.DetailLabel.Render(padRight(rightCol[i].label+":", 12)),
				styles.DetailValue.Render(rightCol[i].value))
		}
		// Pad left column to fixed width
		leftPadded := left + strings.Repeat(" ", max(0, colWidth-visibleLen(left)))
		b.WriteString(leftPadded + right + "\n")
	}

	event := m.detailEvent

	// Tags in multi-column layout
	if len(event.Tags) > 0 {
		b.WriteString("\n" + styles.SectionHeader("Tags", m.width-2) + "\n\n")
		tagCols := 3
		if m.width < 100 {
			tagCols = 2
		}
		if m.width < 60 {
			tagCols = 1
		}
		tagWidth := (m.width - 4) / tagCols
		row := ""
		for i, tag := range event.Tags {
			entry := fmt.Sprintf(" %s %s",
				styles.TagKey.Render(tag.Key+":"),
				truncate(tag.Value, tagWidth-len(tag.Key)-4))
			entry += strings.Repeat(" ", max(0, tagWidth-visibleLen(entry)))
			row += entry
			if (i+1)%tagCols == 0 || i == len(event.Tags)-1 {
				b.WriteString(row + "\n")
				row = ""
			}
		}
	}

	// User context (compact, inline with tags area)
	if event.User != nil {
		u := event.User
		userParts := []string{}
		if u.Email != "" {
			userParts = append(userParts, u.Email)
		}
		if u.Username != "" {
			userParts = append(userParts, "@"+u.Username)
		}
		if u.IPAddr != "" {
			userParts = append(userParts, u.IPAddr)
		}
		if len(userParts) > 0 {
			b.WriteString("\n" + styles.SectionHeader("User", m.width-2) + "\n")
			b.WriteString("  " + styles.DetailValue.Render(strings.Join(userParts, "  ·  ")) + "\n")
		}
	}

	// Stack Traces (from exception entries)
	frameIdx := 0
	for _, entry := range event.Entries {
		switch entry.Type {
		case "exception":
			b.WriteString("\n" + styles.SectionHeader("Exception / Stack Trace", m.width-2) + "\n\n")
			frameIdx = m.renderExceptionsV2(&b, entry.Data, frameIdx)

		case "breadcrumbs":
			b.WriteString("\n" + styles.SectionHeader("Breadcrumbs", m.width-2) + "\n\n")
			m.renderBreadcrumbsV2(&b, entry.Data)

		case "request":
			b.WriteString("\n" + styles.SectionHeader("HTTP Request", m.width-2) + "\n\n")
			m.renderRequestV2(&b, entry.Data)

		case "message":
			b.WriteString("\n" + styles.SectionHeader("Message", m.width-2) + "\n\n")
			if msg, ok := entry.Data["formatted"].(string); ok {
				b.WriteString("  " + msg + "\n")
			}
		}
	}

	// SDK info (compact)
	if event.Sdk.Name != "" {
		b.WriteString("\n " + styles.Subtitle.Render(fmt.Sprintf("SDK: %s %s", event.Sdk.Name, event.Sdk.Version)) + "\n")
	}

	// Permalink
	if issue.Permalink != "" {
		b.WriteString(" " + styles.Subtitle.Render(issue.Permalink) + "\n")
	}

	// Navigation hint
	if m.frameNavMode {
		b.WriteString("\n " + styles.SuccessMsg.Render("FRAME NAV") + " " +
			styles.Subtitle.Render("j/k navigate · space/enter toggle fold · tab exit") + "\n")
	} else {
		b.WriteString("\n " + styles.Subtitle.Render("d/u half-page · tab frame nav · G bottom · gg top") + "\n")
	}

	// Apply scroll
	lines := strings.Split(b.String(), "\n")
	start := m.detailScroll
	if start > len(lines) {
		start = len(lines) - 1
	}
	if start < 0 {
		start = 0
	}
	end := start + m.height - 3
	if end > len(lines) {
		end = len(lines)
	}

	content := strings.Join(lines[start:end], "\n")

	visible := strings.Count(content, "\n") + 1
	for visible < m.height-1 {
		content += "\n"
		visible++
	}

	return content + m.renderFooter()
}

func renderStatusBadge(status string) string {
	switch status {
	case "resolved":
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#fff")).
			Background(styles.Success).
			Padding(0, 1).
			Render(status)
	case "ignored":
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#fff")).
			Background(styles.Muted).
			Padding(0, 1).
			Render(status)
	default: // unresolved
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#fff")).
			Background(styles.Warning).
			Padding(0, 1).
			Render(status)
	}
}

// renderExceptionsV2 renders exception values with foldable stack trace frames.
// Returns the next frameIdx for continued numbering.
func (m Model) renderExceptionsV2(b *strings.Builder, data map[string]any, startFrameIdx int) int {
	valuesRaw, ok := data["values"]
	if !ok {
		return startFrameIdx
	}
	valuesSlice, ok := valuesRaw.([]any)
	if !ok {
		return startFrameIdx
	}

	frameIdx := startFrameIdx

	for _, vRaw := range valuesSlice {
		vMap, ok := vRaw.(map[string]any)
		if !ok {
			continue
		}

		excType, _ := vMap["type"].(string)
		excValue, _ := vMap["value"].(string)

		b.WriteString("  " + styles.ErrorMsg.Render(excType) + " " + styles.DetailValue.Render(excValue) + "\n\n")

		// Stack trace
		stRaw, ok := vMap["stacktrace"]
		if !ok {
			continue
		}
		stMap, ok := stRaw.(map[string]any)
		if !ok {
			continue
		}
		framesRaw, ok := stMap["frames"]
		if !ok {
			continue
		}
		framesSlice, ok := framesRaw.([]any)
		if !ok {
			continue
		}

		frames := parseFrames(framesSlice)

		// Render frames in reverse (most recent first) - like Sentry web UI
		for i := len(frames) - 1; i >= 0; i-- {
			f := frames[i]
			foldKey := fmt.Sprintf("%d", frameIdx)
			isExpanded := m.frameFolds[foldKey]
			isCurrent := m.frameNavMode && frameIdx == m.frameCursor

			// Frame header line (always visible - the "fold" line)
			filename := f.Filename
			if filename == "" {
				filename = f.AbsPath
			}
			funcName := f.Function
			if funcName == "" {
				funcName = "<unknown>"
			}

			// Fold indicator
			foldIcon := "▶"
			if isExpanded {
				foldIcon = "▼"
			}

			// Cursor indicator
			cursor := "  "
			if isCurrent {
				cursor = styles.SuccessMsg.Render("▸ ")
			}

			// Style based on inApp
			var frameLine string
			lineNoStr := fmt.Sprintf("%d", f.LineNo)
			locationStr := fmt.Sprintf("%s:", filename) // filepath:
			if f.InApp {
				frameLine = fmt.Sprintf("%s%s %s  %s%s",
					cursor,
					lipgloss.NewStyle().Foreground(styles.Accent).Render(foldIcon),
					styles.StackFrameApp.Render(funcName),
					styles.Subtitle.Render(locationStr),
					styles.CodeCriticalLineNo.Render(lineNoStr))
			} else {
				frameLine = fmt.Sprintf("%s%s %s  %s%s",
					cursor,
					lipgloss.NewStyle().Foreground(styles.Muted).Render(foldIcon),
					styles.StackFrameLib.Render(funcName),
					styles.Subtitle.Render(locationStr),
					styles.CodeCriticalLineNo.Render(lineNoStr))
			}
			b.WriteString(frameLine + "\n")

			// Expanded content - clean code view like Sentry web UI
			if isExpanded {
				if f.Module != "" {
					b.WriteString("      " + styles.TagKey.Render("module:") + " " + styles.DetailValue.Render(f.Module) + "\n")
				}

				hasContext := len(f.PreContext) > 0 || len(f.Context) > 0 || len(f.PostContext) > 0

				if hasContext {
					// Build a unified list of lines with their line numbers
					// so we can highlight f.LineNo correctly
					type codeLine struct {
						num  int
						text string
					}
					var lines []codeLine

					// Pre-context
					if len(f.PreContext) > 0 {
						startLine := f.LineNo - len(f.PreContext)
						for idx, cl := range f.PreContext {
							lines = append(lines, codeLine{startLine + idx, cl})
						}
					}

					// Critical line from Context field
					if len(f.Context) > 0 {
						for _, ctx := range f.Context {
							if len(ctx) >= 2 {
								num := f.LineNo
								if n, ok := ctx[0].(float64); ok {
									num = int(n)
								}
								lines = append(lines, codeLine{num, fmt.Sprintf("%v", ctx[1])})
							}
						}
					} else {
						// No explicit context - insert a placeholder for the critical line
						lines = append(lines, codeLine{f.LineNo, ""})
					}

					// Post-context
					if len(f.PostContext) > 0 {
						startLine := f.LineNo + 1
						for idx, cl := range f.PostContext {
							lines = append(lines, codeLine{startLine + idx, cl})
						}
					}

					// Render all lines, highlighting f.LineNo
					for _, cl := range lines {
						ln := fmt.Sprintf("%5d", cl.num)
						if cl.num == f.LineNo {
							// This is THE line - highlight it
							if cl.text == "" {
								cl.text = "(source not available)"
							}
							b.WriteString("    " +
								styles.CodeCriticalGutter.Render("▎ ") +
								styles.CodeCriticalLineNo.Render(ln) + "  " +
								styles.CodeCriticalLine.Render(cl.text) + "\n")
						} else {
							// Context line - dimmed
							b.WriteString("      " +
								styles.CodeLineNo.Render(ln) + "  " +
								styles.CodeContextLine.Render(cl.text) + "\n")
						}
					}
				} else {
					// No context at all
					if f.LineNo > 0 {
						ln := fmt.Sprintf("%5d", f.LineNo)
						b.WriteString("    " +
							styles.CodeCriticalGutter.Render("▎ ") +
							styles.CodeCriticalLineNo.Render(ln) + "  " +
							styles.CodeCriticalLine.Render("(source not available)") + "\n")
					} else {
						b.WriteString("      " + styles.Subtitle.Render("(no source context available)") + "\n")
					}
				}
				b.WriteString("\n")
			}

			frameIdx++
		}
		b.WriteString("\n")
	}
	return frameIdx
}

// renderBreadcrumbsV2 renders breadcrumb entries in a compact table format.
func (m Model) renderBreadcrumbsV2(b *strings.Builder, data map[string]any) {
	valuesRaw, ok := data["values"]
	if !ok {
		return
	}
	valuesSlice, ok := valuesRaw.([]any)
	if !ok {
		return
	}

	// Show last 20 breadcrumbs max
	start := 0
	if len(valuesSlice) > 20 {
		start = len(valuesSlice) - 20
		b.WriteString(fmt.Sprintf("  %s\n\n", styles.Subtitle.Render(fmt.Sprintf("(%d earlier breadcrumbs hidden)", start))))
	}

	for i := start; i < len(valuesSlice); i++ {
		bc, ok := valuesSlice[i].(map[string]any)
		if !ok {
			continue
		}
		category, _ := bc["category"].(string)
		level, _ := bc["level"].(string)
		message, _ := bc["message"].(string)
		bcType, _ := bc["type"].(string)
		timestamp, _ := bc["timestamp"].(string)

		// Compact timestamp (just time portion)
		timeStr := ""
		if len(timestamp) > 11 {
			timeStr = timestamp[11:]
			if len(timeStr) > 8 {
				timeStr = timeStr[:8]
			}
		}

		levelStyle := styles.LevelStyle(level)
		levelIndicator := levelStyle.Render("●")

		cat := styles.TagKey.Render(padRight(category, 14))

		msg := message
		if msg == "" && bcType != "" {
			msg = "[" + bcType + "]"
		}
		if msg == "" {
			if d, ok := bc["data"].(map[string]any); ok {
				dataJSON, _ := json.Marshal(d)
				msg = string(dataJSON)
			}
		}

		maxMsg := m.width - 40
		if maxMsg < 20 {
			maxMsg = 20
		}

		b.WriteString(fmt.Sprintf("  %s %s %s %s\n",
			styles.Subtitle.Render(timeStr),
			levelIndicator,
			cat,
			truncate(msg, maxMsg)))
	}
}

// renderRequestV2 renders HTTP request details in a cleaner layout.
func (m Model) renderRequestV2(b *strings.Builder, data map[string]any) {
	method, _ := data["method"].(string)
	url, _ := data["url"].(string)

	methodStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Accent)
	b.WriteString("  " + methodStyle.Render(method) + " " + styles.DetailValue.Render(url) + "\n")

	if query, ok := data["query"].(string); ok && query != "" {
		b.WriteString("  " + styles.TagKey.Render("Query:") + " " + query + "\n")
	}

	if headers, ok := data["headers"].([]any); ok && len(headers) > 0 {
		b.WriteString("\n  " + styles.Subtitle.Render("Headers:") + "\n")
		for _, h := range headers {
			if pair, ok := h.([]any); ok && len(pair) == 2 {
				k := fmt.Sprintf("%v", pair[0])
				v := fmt.Sprintf("%v", pair[1])
				b.WriteString(fmt.Sprintf("    %s %s\n", styles.TagKey.Render(k+":"), v))
			}
		}
	}

	if env, ok := data["env"].(map[string]any); ok {
		b.WriteString("\n  " + styles.Subtitle.Render("Environment:") + "\n")
		for k, v := range env {
			b.WriteString(fmt.Sprintf("    %s %v\n", styles.TagKey.Render(k+":"), v))
		}
	}
}

// visibleLen returns approximate visible character count (strips ANSI).
func visibleLen(s string) int {
	// Simple approximation: strip ANSI escape sequences
	inEscape := false
	count := 0
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		count++
	}
	return count
}

func parseFrames(framesSlice []any) []models.StackFrame {
	var frames []models.StackFrame
	for _, fRaw := range framesSlice {
		fMap, ok := fRaw.(map[string]any)
		if !ok {
			continue
		}
		f := models.StackFrame{}
		f.Filename, _ = fMap["filename"].(string)
		f.Function, _ = fMap["function"].(string)
		f.Module, _ = fMap["module"].(string)
		f.AbsPath, _ = fMap["absPath"].(string)
		f.InApp, _ = fMap["inApp"].(bool)
		if ln, ok := fMap["lineNo"].(float64); ok {
			f.LineNo = int(ln)
		}
		if cn, ok := fMap["colNo"].(float64); ok {
			f.ColNo = int(cn)
		}

		if preCtx, ok := fMap["preContext"].([]any); ok {
			for _, c := range preCtx {
				if s, ok := c.(string); ok {
					f.PreContext = append(f.PreContext, s)
				}
			}
		}
		if postCtx, ok := fMap["postContext"].([]any); ok {
			for _, c := range postCtx {
				if s, ok := c.(string); ok {
					f.PostContext = append(f.PostContext, s)
				}
			}
		}
		if ctx, ok := fMap["context"].([]any); ok {
			for _, c := range ctx {
				if arr, ok := c.([]any); ok {
					f.Context = append(f.Context, arr)
				}
			}
		}

		frames = append(frames, f)
	}
	return frames
}
