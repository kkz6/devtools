package githubmanager

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/ui"
)

// WorkflowRun represents a GitHub Actions workflow run
type WorkflowRun struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Conclusion string    `json:"conclusion"`
	CreatedAt  time.Time `json:"created_at"`
}

// WorkflowRunsResponse represents the response from GitHub API for workflow runs
type WorkflowRunsResponse struct {
	TotalCount   int           `json:"total_count"`
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}

// Deployment represents a GitHub deployment
type Deployment struct {
	ID          int64     `json:"id"`
	SHA         string    `json:"sha"`
	Ref         string    `json:"ref"`
	Task        string    `json:"task"`
	Environment string    `json:"environment"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// deleteActionLogs deletes all action logs from a repository
func deleteActionLogs(cfg *config.Config) error {
	// Get repository information
	owner, repo, err := getRepositoryInfo()
	if err != nil {
		return err
	}

	ui.ShowInfo(fmt.Sprintf("Fetching workflow runs for %s/%s...", owner, repo))

	// Fetch all workflow runs
	runs, err := fetchAllWorkflowRuns(cfg, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to fetch workflow runs: %w", err)
	}

	if len(runs) == 0 {
		ui.ShowInfo("No workflow runs found.")
		return nil
	}

	ui.ShowWarning(fmt.Sprintf("Found %d workflow runs", len(runs)))

	if !ui.GetConfirmation("Are you sure you want to delete ALL workflow run logs? This action cannot be undone.") {
		ui.ShowInfo("Operation cancelled.")
		return nil
	}

	// Delete each workflow run
	progressBar := ui.NewProgressBar("Deleting workflow runs", len(runs))
	deletedCount := 0
	failedCount := 0

	for _, run := range runs {
		progressBar.UpdateTitle(fmt.Sprintf("Deleting run #%d (%s)", run.ID, run.Name))

		if err := deleteWorkflowRun(cfg, owner, repo, run.ID); err != nil {
			failedCount++
			// Continue with other deletions even if one fails
		} else {
			deletedCount++
		}

		progressBar.Increment()
	}

	progressBar.Finish()

	ui.ShowSuccess(fmt.Sprintf("Successfully deleted %d workflow runs", deletedCount))
	if failedCount > 0 {
		ui.ShowWarning(fmt.Sprintf("Failed to delete %d workflow runs", failedCount))
	}

	return nil
}

// deleteDeployments deletes all deployments from a repository
func deleteDeployments(cfg *config.Config) error {
	// Get repository information
	owner, repo, err := getRepositoryInfo()
	if err != nil {
		return err
	}

	ui.ShowInfo(fmt.Sprintf("Fetching deployments for %s/%s...", owner, repo))

	// Fetch all deployments
	deployments, err := fetchAllDeployments(cfg, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to fetch deployments: %w", err)
	}

	if len(deployments) == 0 {
		ui.ShowInfo("No deployments found.")
		return nil
	}

	ui.ShowWarning(fmt.Sprintf("Found %d deployments", len(deployments)))

	if !ui.GetConfirmation("Are you sure you want to delete ALL deployments? This action cannot be undone.") {
		ui.ShowInfo("Operation cancelled.")
		return nil
	}

	// Delete each deployment
	progressBar := ui.NewProgressBar("Deleting deployments", len(deployments))
	deletedCount := 0
	failedCount := 0

	for _, deployment := range deployments {
		progressBar.UpdateTitle(fmt.Sprintf("Deleting deployment #%d (%s)", deployment.ID, deployment.Environment))

		if err := deleteDeployment(cfg, owner, repo, deployment.ID); err != nil {
			failedCount++
			// Continue with other deletions even if one fails
		} else {
			deletedCount++
		}

		progressBar.Increment()
	}

	progressBar.Finish()

	ui.ShowSuccess(fmt.Sprintf("Successfully deleted %d deployments", deletedCount))
	if failedCount > 0 {
		ui.ShowWarning(fmt.Sprintf("Failed to delete %d deployments", failedCount))
	}

	return nil
}

// getRepositoryInfo prompts the user for repository information
func getRepositoryInfo() (string, string, error) {
	repoPath, err := ui.GetInput("Enter repository (owner/repo):", "", false, func(s string) error {
		if s == "" {
			return fmt.Errorf("repository path cannot be empty")
		}
		parts := strings.Split(s, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository format. Expected: owner/repo")
		}
		return nil
	})
	if err != nil {
		return "", "", err
	}

	parts := strings.Split(repoPath, "/")
	return parts[0], parts[1], nil
}

// fetchAllWorkflowRuns fetches all workflow runs from a repository
func fetchAllWorkflowRuns(cfg *config.Config, owner, repo string) ([]WorkflowRun, error) {
	var allRuns []WorkflowRun
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs?per_page=%d&page=%d", owner, repo, perPage, page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "token "+cfg.GitHub.Token)
		req.Header.Set("Accept", "application/vnd.github+json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
		}

		var response WorkflowRunsResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		allRuns = append(allRuns, response.WorkflowRuns...)

		// Check if there are more pages
		if len(response.WorkflowRuns) < perPage {
			break
		}

		page++
	}

	return allRuns, nil
}

// deleteWorkflowRun deletes a specific workflow run
func deleteWorkflowRun(cfg *config.Config, owner, repo string, runID int64) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs/%d", owner, repo, runID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+cfg.GitHub.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// fetchAllDeployments fetches all deployments from a repository
func fetchAllDeployments(cfg *config.Config, owner, repo string) ([]Deployment, error) {
	var allDeployments []Deployment
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/deployments?per_page=%d&page=%d", owner, repo, perPage, page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "token "+cfg.GitHub.Token)
		req.Header.Set("Accept", "application/vnd.github+json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
		}

		var deployments []Deployment
		if err := json.NewDecoder(resp.Body).Decode(&deployments); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		allDeployments = append(allDeployments, deployments...)

		// Check if there are more pages
		if len(deployments) < perPage {
			break
		}

		page++
	}

	return allDeployments, nil
}

// deleteDeployment deletes a specific deployment
func deleteDeployment(cfg *config.Config, owner, repo string, deploymentID int64) error {
	// First, we need to set the deployment status to inactive
	// GitHub requires deployments to be inactive before deletion
	if err := setDeploymentInactive(cfg, owner, repo, deploymentID); err != nil {
		return fmt.Errorf("failed to set deployment inactive: %w", err)
	}

	// Now delete the deployment
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/deployments/%d", owner, repo, deploymentID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+cfg.GitHub.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// setDeploymentInactive sets a deployment status to inactive
func setDeploymentInactive(cfg *config.Config, owner, repo string, deploymentID int64) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/deployments/%d/statuses", owner, repo, deploymentID)

	payload := map[string]interface{}{
		"state":       "inactive",
		"description": "Deployment marked as inactive for deletion",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+cfg.GitHub.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
