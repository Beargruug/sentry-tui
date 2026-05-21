// Package models defines Sentry data types used throughout the application.
package models

import "time"

// Organization represents a Sentry organization.
type Organization struct {
	ID       string `json:"id"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Avatar   Avatar `json:"avatar"`
	Status   Status `json:"status"`
	DateCreated time.Time `json:"dateCreated"`
}

// Avatar represents an entity avatar.
type Avatar struct {
	Type string `json:"avatarType"`
	UUID string `json:"avatarUuid"`
}

// Status represents an entity status.
type Status struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Project represents a Sentry project.
type Project struct {
	ID           string       `json:"id"`
	Slug         string       `json:"slug"`
	Name         string       `json:"name"`
	Platform     string       `json:"platform"`
	DateCreated  time.Time    `json:"dateCreated"`
	IsBookmarked bool         `json:"isBookmarked"`
	Color        string       `json:"color"`
	Status       string       `json:"status"`
	Organization Organization `json:"organization"`
}

// Environment represents a Sentry environment.
type Environment struct {
	Name string `json:"name"`
}

// Team represents a Sentry team.
type Team struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	DateCreated time.Time `json:"dateCreated"`
	MemberCount int       `json:"memberCount"`
}

// Member represents a Sentry organization member.
type Member struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	User  User   `json:"user"`
	Role  string `json:"role"`
}

// User represents a Sentry user.
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   Avatar `json:"avatar"`
}

// Issue represents a Sentry issue (group).
type Issue struct {
	ID             string         `json:"id"`
	ShortID        string         `json:"shortId"`
	Title          string         `json:"title"`
	Culprit        string         `json:"culprit"`
	Level          string         `json:"level"`
	Status         string         `json:"status"`
	StatusDetails  map[string]any `json:"statusDetails"`
	IsPublic       bool           `json:"isPublic"`
	IsBookmarked   bool           `json:"isBookmarked"`
	Project        ProjectRef     `json:"project"`
	Type           string         `json:"type"`
	Platform       string         `json:"platform"`
	Count          string         `json:"count"`
	UserCount      int            `json:"userCount"`
	FirstSeen      time.Time      `json:"firstSeen"`
	LastSeen       time.Time      `json:"lastSeen"`
	AssignedTo     *AssignedTo    `json:"assignedTo"`
	Metadata       IssueMetadata  `json:"metadata"`
	HasSeen        bool           `json:"hasSeen"`
	Permalink      string         `json:"permalink"`
	NumComments    int            `json:"numComments"`
	Annotations    []string       `json:"annotations"`
	IsUnhandled    bool           `json:"isUnhandled"`
	Logger         string         `json:"logger"`
}

// ProjectRef is a lightweight project reference embedded in issues.
type ProjectRef struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Platform string `json:"platform"`
}

// AssignedTo represents the assigned user/team for an issue.
type AssignedTo struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// IssueMetadata carries extra metadata for an issue.
type IssueMetadata struct {
	Value    string `json:"value"`
	Type     string `json:"type"`
	Filename string `json:"filename"`
	Function string `json:"function"`
}

// Event represents a Sentry event (the latest event for an issue).
type Event struct {
	EventID     string            `json:"eventID"`
	ID          string            `json:"id"`
	GroupID     string            `json:"groupID"`
	Title       string            `json:"title"`
	Message     string            `json:"message"`
	Platform    string            `json:"platform"`
	DateCreated time.Time         `json:"dateCreated"`
	Tags        []Tag             `json:"tags"`
	Context     map[string]any    `json:"context"`
	Contexts    map[string]any    `json:"contexts"`
	Entries     []EventEntry      `json:"entries"`
	Sdk         SdkInfo           `json:"sdk"`
	User        *EventUser        `json:"user"`
}

// Tag is a Sentry event tag.
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// EventEntry represents a section of event data (exception, breadcrumbs, request, etc.).
type EventEntry struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

// SdkInfo holds SDK metadata.
type SdkInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// EventUser represents user context from an event.
type EventUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	IPAddr   string `json:"ip_address"`
}

// ----- Parsed helpers for stack trace display -----

// StackFrame represents a single frame in a stack trace.
type StackFrame struct {
	Filename    string   `json:"filename"`
	Function    string   `json:"function"`
	Module      string   `json:"module"`
	LineNo      int      `json:"lineNo"`
	ColNo       int      `json:"colNo"`
	AbsPath     string   `json:"absPath"`
	InApp       bool     `json:"inApp"`
	Context     [][]any  `json:"context"`
	PreContext  []string `json:"preContext"`
	PostContext []string `json:"postContext"`
}

// ExceptionValue represents a single exception in a chain.
type ExceptionValue struct {
	Type       string       `json:"type"`
	Value      string       `json:"value"`
	Module     string       `json:"module"`
	Mechanism  map[string]any `json:"mechanism"`
	Stacktrace *Stacktrace `json:"stacktrace"`
}

// Stacktrace holds a list of frames.
type Stacktrace struct {
	Frames []StackFrame `json:"frames"`
}

// Breadcrumb represents a single breadcrumb entry.
type Breadcrumb struct {
	Timestamp time.Time      `json:"timestamp"`
	Type      string         `json:"type"`
	Category  string         `json:"category"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Data      map[string]any `json:"data"`
}

// IssueListCursor holds pagination state.
type IssueListCursor struct {
	NextCursor string
	PrevCursor string
	HasNext    bool
	HasPrev    bool
}

// FilterState holds the current filter/search criteria.
type FilterState struct {
	Query       string
	Project     string // project slug (for display)
	ProjectID   string // numeric project ID (for API calls)
	Environment string
	Status      string // "unresolved", "resolved", "ignored"
	Sort        string // "date", "new", "priority", "freq", "user"
	Page        int
	Cursor      string
}

// DefaultFilter returns the default filter state.
func DefaultFilter() FilterState {
	return FilterState{
		Status: "unresolved",
		Sort:   "date",
		Page:   1,
	}
}
