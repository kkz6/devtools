package bugmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LinearClient handles Linear API interactions
type LinearClient struct {
	apiKey string
	client *http.Client
}

// NewLinearClient creates a new Linear API client
func NewLinearClient(apiKey string) *LinearClient {
	return &LinearClient{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// LinearTeam represents a Linear team
type LinearTeam struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

// LinearProject represents a Linear project
type LinearProject struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	State       string `json:"state"`
}

// LinearLabel represents a Linear label
type LinearLabel struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// LinearIssue represents a Linear issue
type LinearIssue struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	State       struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"state"`
	Labels struct {
		Nodes []LinearLabel `json:"nodes"`
	} `json:"labels"`
	URL string `json:"url"`
}

// LinearWorkflowState represents a Linear workflow state
type LinearWorkflowState struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Color string `json:"color"`
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string `json:"message"`
}

// executeGraphQL executes a GraphQL query
func (c *LinearClient) executeGraphQL(query string, variables map[string]interface{}) (json.RawMessage, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.linear.app/graphql", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var graphQLResp GraphQLResponse
	if err := json.Unmarshal(respBody, &graphQLResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", graphQLResp.Errors[0].Message)
	}

	return graphQLResp.Data, nil
}

// GetTeams fetches all teams
func (c *LinearClient) GetTeams() ([]LinearTeam, error) {
	query := `
		query {
			teams {
				nodes {
					id
					name
					key
				}
			}
		}
	`

	data, err := c.executeGraphQL(query, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Teams struct {
			Nodes []LinearTeam `json:"nodes"`
		} `json:"teams"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal teams: %w", err)
	}

	return result.Teams.Nodes, nil
}

// GetProjects fetches all projects for a team
func (c *LinearClient) GetProjects(teamID string) ([]LinearProject, error) {
	query := `
		query GetProjects($teamId: String!) {
			team(id: $teamId) {
				projects {
					nodes {
						id
						name
						description
						state
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"teamId": teamID,
	}

	data, err := c.executeGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Team struct {
			Projects struct {
				Nodes []LinearProject `json:"nodes"`
			} `json:"projects"`
		} `json:"team"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal projects: %w", err)
	}

	return result.Team.Projects.Nodes, nil
}

// GetOrCreateLabel gets or creates a label
func (c *LinearClient) GetOrCreateLabel(teamID, name, color string) (string, error) {
	// First, try to find existing label
	query := `
		query GetLabel($teamId: String!, $name: String!) {
			team(id: $teamId) {
				labels(filter: { name: { eq: $name } }) {
					nodes {
						id
						name
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"teamId": teamID,
		"name":   name,
	}

	data, err := c.executeGraphQL(query, variables)
	if err != nil {
		return "", err
	}

	var result struct {
		Team struct {
			Labels struct {
				Nodes []LinearLabel `json:"nodes"`
			} `json:"labels"`
		} `json:"team"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal labels: %w", err)
	}

	if len(result.Team.Labels.Nodes) > 0 {
		return result.Team.Labels.Nodes[0].ID, nil
	}

	// Create new label if not found
	createQuery := `
		mutation CreateLabel($teamId: String!, $name: String!, $color: String!) {
			issueLabelCreate(input: {
				teamId: $teamId
				name: $name
				color: $color
			}) {
				success
				issueLabel {
					id
				}
			}
		}
	`

	createVariables := map[string]interface{}{
		"teamId": teamID,
		"name":   name,
		"color":  color,
	}

	createData, err := c.executeGraphQL(createQuery, createVariables)
	if err != nil {
		return "", err
	}

	var createResult struct {
		IssueLabelCreate struct {
			Success    bool `json:"success"`
			IssueLabel struct {
				ID string `json:"id"`
			} `json:"issueLabel"`
		} `json:"issueLabelCreate"`
	}

	if err := json.Unmarshal(createData, &createResult); err != nil {
		return "", fmt.Errorf("failed to unmarshal create result: %w", err)
	}

	if !createResult.IssueLabelCreate.Success {
		return "", fmt.Errorf("failed to create label")
	}

	return createResult.IssueLabelCreate.IssueLabel.ID, nil
}

// CreateIssue creates a new issue in Linear
func (c *LinearClient) CreateIssue(teamID, projectID, title, description string, labelIDs []string, priority int, stateID string) (*LinearIssue, error) {
	query := `
		mutation CreateIssue($teamId: String!, $projectId: String!, $title: String!, $description: String!, $labelIds: [String!], $priority: Int!, $stateId: String) {
			issueCreate(input: {
				teamId: $teamId
				projectId: $projectId
				title: $title
				description: $description
				labelIds: $labelIds
				priority: $priority
				stateId: $stateId
			}) {
				success
				issue {
					id
					title
					description
					priority
					url
					state {
						id
						name
					}
					labels {
						nodes {
							id
							name
							color
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"teamId":      teamID,
		"projectId":   projectID,
		"title":       title,
		"description": description,
		"labelIds":    labelIDs,
		"priority":    priority,
	}

	// Only add stateId if provided
	if stateID != "" {
		variables["stateId"] = stateID
	}

	data, err := c.executeGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		IssueCreate struct {
			Success bool        `json:"success"`
			Issue   LinearIssue `json:"issue"`
		} `json:"issueCreate"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal issue: %w", err)
	}

	if !result.IssueCreate.Success {
		return nil, fmt.Errorf("failed to create issue")
	}

	return &result.IssueCreate.Issue, nil
}

// GetWorkflowStates fetches workflow states for a team
func (c *LinearClient) GetWorkflowStates(teamID string) ([]LinearWorkflowState, error) {
	query := `
		query GetWorkflowStates($teamId: String!) {
			team(id: $teamId) {
				states {
					nodes {
						id
						name
						type
						color
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"teamId": teamID,
	}

	data, err := c.executeGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Team struct {
			States struct {
				Nodes []LinearWorkflowState `json:"nodes"`
			} `json:"states"`
		} `json:"team"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow states: %w", err)
	}

	return result.Team.States.Nodes, nil
}
