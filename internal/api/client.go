// Package api provides the Sentry REST API client.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Beargruug/sentry-tui/internal/models"
)

// Client communicates with the Sentry API.
type Client struct {
	baseURL    string
	authToken  string
	org        string
	httpClient *http.Client
}

// NewClient creates a new Sentry API client.
func NewClient(baseURL, authToken, org string) *Client {
	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		authToken: authToken,
		org:       org,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ---------- Low-level helpers ----------

func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	reqURL := c.baseURL + path
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.authToken)
	req.Header.Set("Content-Type", "application/json")
	return c.httpClient.Do(req)
}

func (c *Client) get(path string) (*http.Response, error) {
	return c.doRequest(http.MethodGet, path, nil)
}

func (c *Client) put(path string, body io.Reader) (*http.Response, error) {
	return c.doRequest(http.MethodPut, path, body)
}

func decodeJSON[T any](resp *http.Response) (T, error) {
	var result T
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return result, fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("decoding response: %w", err)
	}
	return result, nil
}

// parseLinkHeader extracts cursor values from the Sentry Link header.
func parseLinkHeader(header string) models.IssueListCursor {
	cursor := models.IssueListCursor{}
	if header == "" {
		return cursor
	}

	re := regexp.MustCompile(`<[^>]+[?&]cursor=([^&>]+)[^>]*>;\s*rel="(\w+)";\s*results="(\w+)"`)
	matches := re.FindAllStringSubmatch(header, -1)
	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		cursorVal := m[1]
		rel := m[2]
		hasResults := m[3] == "true"

		switch rel {
		case "next":
			cursor.NextCursor = cursorVal
			cursor.HasNext = hasResults
		case "previous":
			cursor.PrevCursor = cursorVal
			cursor.HasPrev = hasResults
		}
	}
	return cursor
}

// ---------- Organizations ----------

// ListOrganizations fetches all organizations the token can access.
func (c *Client) ListOrganizations() ([]models.Organization, error) {
	resp, err := c.get("/organizations/")
	if err != nil {
		return nil, err
	}
	return decodeJSON[[]models.Organization](resp)
}

// ---------- Projects ----------

// ListProjects fetches all projects for the configured organization.
func (c *Client) ListProjects() ([]models.Project, error) {
	resp, err := c.get(fmt.Sprintf("/organizations/%s/projects/", c.org))
	if err != nil {
		return nil, err
	}
	return decodeJSON[[]models.Project](resp)
}

// ---------- Environments ----------

// ListEnvironments fetches all environments for the configured organization.
func (c *Client) ListEnvironments() ([]models.Environment, error) {
	resp, err := c.get(fmt.Sprintf("/organizations/%s/environments/", c.org))
	if err != nil {
		return nil, err
	}
	return decodeJSON[[]models.Environment](resp)
}

// ---------- Teams / Members ----------

// ListMembers fetches organization members.
func (c *Client) ListMembers() ([]models.Member, error) {
	resp, err := c.get(fmt.Sprintf("/organizations/%s/members/", c.org))
	if err != nil {
		return nil, err
	}
	return decodeJSON[[]models.Member](resp)
}

// ---------- Issues ----------

// IssueResult contains issues and pagination cursor.
type IssueResult struct {
	Issues []models.Issue
	Cursor models.IssueListCursor
}

// ListIssues fetches issues with filters.
func (c *Client) ListIssues(filter models.FilterState) (IssueResult, error) {
	params := url.Values{}
	if filter.Query != "" {
		params.Set("query", filter.Query)
	}
	if filter.ProjectID != "" {
		params.Set("project", filter.ProjectID)
	}
	if filter.Environment != "" {
		params.Set("environment", filter.Environment)
	}
	if filter.Status != "" {
		// Sentry uses "is:unresolved" style search queries
		existing := params.Get("query")
		statusQuery := "is:" + filter.Status
		if existing != "" {
			params.Set("query", existing+" "+statusQuery)
		} else {
			params.Set("query", statusQuery)
		}
	}
	if filter.Sort != "" {
		params.Set("sort", filter.Sort)
	}
	if filter.Cursor != "" {
		params.Set("cursor", filter.Cursor)
	}

	path := fmt.Sprintf("/organizations/%s/issues/?%s", c.org, params.Encode())
	resp, err := c.get(path)
	if err != nil {
		return IssueResult{}, err
	}

	linkHeader := resp.Header.Get("Link")
	cursor := parseLinkHeader(linkHeader)

	issues, err := decodeJSON[[]models.Issue](resp)
	if err != nil {
		return IssueResult{}, err
	}

	return IssueResult{Issues: issues, Cursor: cursor}, nil
}

// GetIssue fetches a single issue by ID.
func (c *Client) GetIssue(issueID string) (models.Issue, error) {
	// Try org-scoped endpoint first, fall back to global
	resp, err := c.get(fmt.Sprintf("/organizations/%s/issues/%s/", c.org, issueID))
	if err != nil {
		return models.Issue{}, err
	}
	if resp.StatusCode == 404 {
		resp.Body.Close()
		// Fall back to global endpoint
		resp, err = c.get(fmt.Sprintf("/issues/%s/", issueID))
		if err != nil {
			return models.Issue{}, err
		}
	}
	return decodeJSON[models.Issue](resp)
}

// GetLatestEvent fetches the latest event for an issue.
func (c *Client) GetLatestEvent(issueID string) (models.Event, error) {
	resp, err := c.get(fmt.Sprintf("/organizations/%s/issues/%s/events/latest/", c.org, issueID))
	if err != nil {
		return models.Event{}, err
	}
	if resp.StatusCode == 404 {
		resp.Body.Close()
		resp, err = c.get(fmt.Sprintf("/issues/%s/events/latest/", issueID))
		if err != nil {
			return models.Event{}, err
		}
	}
	return decodeJSON[models.Event](resp)
}

// ---------- Issue Actions ----------

// ResolveIssue marks an issue as resolved.
func (c *Client) ResolveIssue(issueID string) error {
	body := strings.NewReader(`{"status":"resolved"}`)
	resp, err := c.put(fmt.Sprintf("/organizations/%s/issues/%s/", c.org, issueID), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resolve failed %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// UnresolveIssue marks an issue as unresolved.
func (c *Client) UnresolveIssue(issueID string) error {
	body := strings.NewReader(`{"status":"unresolved"}`)
	resp, err := c.put(fmt.Sprintf("/organizations/%s/issues/%s/", c.org, issueID), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unresolve failed %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// IgnoreIssue marks an issue as ignored.
func (c *Client) IgnoreIssue(issueID string) error {
	body := strings.NewReader(`{"status":"ignored"}`)
	resp, err := c.put(fmt.Sprintf("/organizations/%s/issues/%s/", c.org, issueID), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ignore failed %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// AssignIssue assigns an issue to a user by email or member ID.
func (c *Client) AssignIssue(issueID, assignee string) error {
	payload := fmt.Sprintf(`{"assignedTo":"%s"}`, assignee)
	body := strings.NewReader(payload)
	resp, err := c.put(fmt.Sprintf("/organizations/%s/issues/%s/", c.org, issueID), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("assign failed %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// TestConnection makes a lightweight API call to verify credentials.
func (c *Client) TestConnection() error {
	resp, err := c.get("/")
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return fmt.Errorf("authentication failed — check your auth token")
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	return nil
}
