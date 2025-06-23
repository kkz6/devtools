package bugmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/getsentry/sentry-go"
)

// SentryClient handles Sentry API interactions
type SentryClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewSentryClient creates a new Sentry API client
func NewSentryClient(apiKey, baseURL string) *SentryClient {
	// Initialize Sentry SDK for error reporting (optional)
	sentry.Init(sentry.ClientOptions{
		Dsn:              "", // We're not sending errors to Sentry, just using the client
		AttachStacktrace: true,
	})

	return &SentryClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// SentryProject represents a Sentry project
type SentryProject struct {
	ID           string `json:"id"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Organization struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
		Name string `json:"name"`
	} `json:"organization"`
	// Keep this for backward compatibility
	OrganizationSlug string `json:"organization_slug"`
}

// SentryIssue represents a Sentry issue
type SentryIssue struct {
	ID            string                 `json:"id"`
	ShortID       string                 `json:"shortId"`
	Title         string                 `json:"title"`
	Culprit       string                 `json:"culprit"`
	Permalink     string                 `json:"permalink"`
	Count         string                 `json:"count"`
	UserCount     int                    `json:"userCount"`
	FirstSeen     time.Time              `json:"firstSeen"`
	LastSeen      time.Time              `json:"lastSeen"`
	Level         string                 `json:"level"`
	Status        string                 `json:"status"`
	StatusDetails map[string]interface{} `json:"statusDetails"`
	IsPublic      bool                   `json:"isPublic"`
	Platform      string                 `json:"platform"`
	Project       struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"project"`
	Type     string                 `json:"type"`
	Metadata map[string]interface{} `json:"metadata"`
}

// SentryEvent represents a Sentry event with stack trace
type SentryEvent struct {
	ID       string    `json:"id"`
	EventID  string    `json:"eventID"`
	Title    string    `json:"title"`
	Message  string    `json:"message"`
	Platform string    `json:"platform"`
	DateTime time.Time `json:"dateTime"`
	Entries  []struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	} `json:"entries"`
	Exception struct {
		Values []struct {
			Type       string `json:"type"`
			Value      string `json:"value"`
			Stacktrace struct {
				Frames []struct {
					Filename string          `json:"filename"`
					Function string          `json:"function"`
					Module   string          `json:"module"`
					LineNo   int             `json:"lineNo"`
					ColNo    int             `json:"colNo"`
					AbsPath  string          `json:"absPath"`
					Context  [][]interface{} `json:"context"`
					InApp    bool            `json:"inApp"`
				} `json:"frames"`
			} `json:"stacktrace"`
		} `json:"values"`
	} `json:"exception"`
}

// GetProjects fetches all projects from Sentry
func (c *SentryClient) GetProjects() ([]SentryProject, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/", c.baseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var projects []SentryProject
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Ensure organization slug is populated from nested structure
	for i := range projects {
		if projects[i].OrganizationSlug == "" && projects[i].Organization.Slug != "" {
			projects[i].OrganizationSlug = projects[i].Organization.Slug
		}
	}

	return projects, nil
}

// GetUnresolvedIssues fetches unresolved issues for a specific project
func (c *SentryClient) GetUnresolvedIssues(organizationSlug, projectSlug string, limit int) ([]SentryIssue, error) {
	// Build query parameters
	params := url.Values{}
	params.Set("query", "is:unresolved")
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("sort", "date")
	params.Set("statsPeriod", "24h")

	url := fmt.Sprintf("%s/projects/%s/%s/issues/?%s",
		c.baseURL, organizationSlug, projectSlug, params.Encode())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var issues []SentryIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return issues, nil
}

// GetIssueDetails fetches detailed information about a specific issue
func (c *SentryClient) GetIssueDetails(issueID string) (*SentryIssue, error) {
	url := fmt.Sprintf("%s/issues/%s/", c.baseURL, issueID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issue details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var issue SentryIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &issue, nil
}

// GetLatestEvent fetches the latest event for an issue to get stack trace
func (c *SentryClient) GetLatestEvent(issueID string) (*SentryEvent, error) {
	url := fmt.Sprintf("%s/issues/%s/events/latest/", c.baseURL, issueID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var event SentryEvent
	if err := json.NewDecoder(resp.Body).Decode(&event); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &event, nil
}

// ResolveIssue marks an issue as resolved in Sentry
func (c *SentryClient) ResolveIssue(issueID string) error {
	url := fmt.Sprintf("%s/issues/%s/", c.baseURL, issueID)

	// Create request body
	reqBody := map[string]interface{}{
		"status": "resolved",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to resolve issue: %s (status: %d)", string(respBody), resp.StatusCode)
	}

	return nil
}
