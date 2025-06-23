package bugmanager

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/types"
	"github.com/kkz6/devtools/internal/ui"
)

// Module implements the bug manager module
type Module struct{}

// New creates a new bug manager module
func New() *Module {
	return &Module{}
}

// Info returns module information
func (m *Module) Info() types.ModuleInfo {
	return types.ModuleInfo{
		ID:          "bugmanager",
		Name:        "Issue Manager (Sentry/Linear)",
		Description: "Sync bugs from Sentry to Linear & Create issues manually",
	}
}

// Execute runs the bug manager module
func (m *Module) Execute(cfg *config.Config) error {
	// Check if API keys are configured
	if cfg.Sentry.APIKey == "" || cfg.Linear.APIKey == "" {
		if err := m.configureAPIs(cfg); err != nil {
			return err
		}
	}

	// Check if there are any configured projects
	if len(cfg.Sentry.Projects) == 0 && cfg.Sentry.APIKey != "" && cfg.Linear.APIKey != "" {
		ui.ShowInfo("No project mappings found. Please configure projects first.")
	}

	// Show main menu
	return m.showMainMenu(cfg)
}

// showMainMenu displays the main bug manager menu
func (m *Module) showMainMenu(cfg *config.Config) error {
	for {
		options := []string{
			"Sync Bugs from Sentry",
			"Create Manual Issue",
			"Configure Projects",
			"Configure APIs",
			"Test API Connectivity",
			"Resolve Sentry Issues",
			"Back",
		}

		choice, err := ui.SelectFromList("Issue Manager", options)
		if err != nil {
			return err
		}

		switch choice {
		case 0: // Sync bugs
			if err := m.syncBugs(cfg); err != nil && err != types.ErrNavigateBack {
				return err
			}
		case 1: // Create manual issue
			if err := m.createManualIssue(cfg); err != nil && err != types.ErrNavigateBack {
				return err
			}
		case 2: // Configure projects
			if err := m.configureProjects(cfg); err != nil && err != types.ErrNavigateBack {
				return err
			}
		case 3: // Configure APIs
			if err := m.configureAPIs(cfg); err != nil && err != types.ErrNavigateBack {
				return err
			}
		case 4: // Test connectivity
			if err := m.testConnectivity(cfg); err != nil && err != types.ErrNavigateBack {
				return err
			}
		case 5: // Resolve Sentry issues
			if err := m.resolveIssues(cfg); err != nil && err != types.ErrNavigateBack {
				return err
			}
		case 6: // Back
			return types.ErrNavigateBack
		}
	}
}

// configureAPIs handles API key configuration
func (m *Module) configureAPIs(cfg *config.Config) error {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	fmt.Println(titleStyle.Render("Configure API Keys"))

	// Configure Sentry API Key (sensitive input)
	sentryKey, err := ui.GetInput("Enter Sentry API Key", cfg.Sentry.APIKey, true, nil)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}
	cfg.Sentry.APIKey = sentryKey

	// Configure Linear API Key (sensitive input)
	linearKey, err := ui.GetInput("Enter Linear API Key", cfg.Linear.APIKey, true, nil)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}
	cfg.Linear.APIKey = linearKey

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	ui.ShowSuccess("API keys configured successfully!")
	return nil
}

// testConnectivity tests the API connectivity for both Sentry and Linear
func (m *Module) testConnectivity(cfg *config.Config) error {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	fmt.Println(titleStyle.Render("Testing API Connectivity"))

	// Test Sentry connectivity
	if cfg.Sentry.APIKey != "" {
		ui.ShowInfo("Testing Sentry API connection...")
		sentryClient := NewSentryClient(cfg.Sentry.APIKey, cfg.Sentry.BaseURL)

		projects, err := sentryClient.GetProjects()
		if err != nil {
			ui.ShowWarning(fmt.Sprintf("Sentry API test failed: %v", err))
		} else {
			ui.ShowSuccess(fmt.Sprintf("Sentry API connected successfully! Found %d projects.", len(projects)))
		}
	} else {
		ui.ShowWarning("Sentry API key not configured")
	}

	// Test Linear connectivity
	if cfg.Linear.APIKey != "" {
		ui.ShowInfo("Testing Linear API connection...")
		linearClient := NewLinearClient(cfg.Linear.APIKey)

		teams, err := linearClient.GetTeams()
		if err != nil {
			ui.ShowWarning(fmt.Sprintf("Linear API test failed: %v", err))
		} else {
			ui.ShowSuccess(fmt.Sprintf("Linear API connected successfully! Found %d teams.", len(teams)))
		}
	} else {
		ui.ShowWarning("Linear API key not configured")
	}

	fmt.Println()
	return nil
}

// createManualIssue handles manual issue creation in Linear
func (m *Module) createManualIssue(cfg *config.Config) error {
	// Check if Linear API key is configured
	if cfg.Linear.APIKey == "" {
		ui.ShowError("Linear API key not configured. Please configure it first.")
		return nil
	}

	// Initialize Linear client
	linearClient := NewLinearClient(cfg.Linear.APIKey)

	// Fetch teams
	ui.ShowInfo("Fetching Linear teams...")
	teams, err := linearClient.GetTeams()
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to fetch teams: %v", err))
		return nil
	}

	if len(teams) == 0 {
		ui.ShowWarning("No teams found in Linear.")
		return nil
	}

	// Select team
	teamOptions := make([]string, len(teams))
	for i, team := range teams {
		teamOptions[i] = fmt.Sprintf("%s (%s)", team.Name, team.Key)
	}

	teamChoice, err := ui.SelectFromList("Select team", teamOptions)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	selectedTeam := teams[teamChoice]

	// Fetch projects for the team
	ui.ShowInfo("Fetching projects...")
	projects, err := linearClient.GetProjects(selectedTeam.ID)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to fetch projects: %v", err))
		return nil
	}

	// Select project (optional)
	var selectedProjectID string
	if len(projects) > 0 {
		projectOptions := []string{"No Project (Team only)"}
		for _, proj := range projects {
			projectOptions = append(projectOptions, proj.Name)
		}

		projectChoice, err := ui.SelectFromList("Select project (optional)", projectOptions)
		if err != nil {
			if err.Error() == "cancelled" {
				return types.ErrNavigateBack
			}
			return err
		}

		if projectChoice > 0 {
			selectedProjectID = projects[projectChoice-1].ID
		}
	}

	// Get issue details
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	fmt.Println(titleStyle.Render("Create New Issue"))

	// Issue type
	issueTypes := []string{"Bug", "Feature", "Task", "Improvement", "Story"}
	typeChoice, err := ui.SelectFromList("Select issue type", issueTypes)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}
	issueType := issueTypes[typeChoice]

	// Title
	title, err := ui.GetInput(
		"Issue Title",
		"",
		false,
		func(s string) error {
			if len(s) < 3 {
				return fmt.Errorf("title must be at least 3 characters")
			}
			return nil
		},
	)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	// Add issue type prefix if not already present
	if !strings.HasPrefix(strings.ToLower(title), strings.ToLower(issueType)) {
		title = fmt.Sprintf("[%s] %s", issueType, title)
	}

	// Description
	descStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	fmt.Println()
	fmt.Println(descStyle.Render("Issue Description"))
	fmt.Println("Enter description line by line. Press Enter on empty line twice to finish.")
	fmt.Println()

	var descLines []string
	emptyCount := 0
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("▸ ")
		if scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				emptyCount++
				if emptyCount >= 2 {
					// Remove the last empty line since it was just to signal end
					if len(descLines) > 0 && descLines[len(descLines)-1] == "" {
						descLines = descLines[:len(descLines)-1]
					}
					break
				}
				// Add empty line to description
				descLines = append(descLines, "")
			} else {
				emptyCount = 0
				descLines = append(descLines, line)
			}
		} else {
			// Handle scanner error or EOF
			break
		}
	}
	description := strings.Join(descLines, "\n")

	// Priority
	fmt.Println()
	priorities := []string{
		"No priority",
		"Urgent",
		"High",
		"Medium",
		"Low",
	}
	priorityChoice, err := ui.SelectFromList("Select priority", priorities)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	// Labels
	fmt.Println()
	labelsInput, err := ui.GetInput(
		"Labels (comma-separated, press Enter to skip)",
		strings.ToLower(issueType),
		false,
		nil,
	)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		// Don't return on other errors, just use default
		labelsInput = strings.ToLower(issueType)
	}

	labels := []string{}
	if labelsInput != "" {
		for _, label := range strings.Split(labelsInput, ",") {
			trimmed := strings.TrimSpace(label)
			if trimmed != "" {
				labels = append(labels, trimmed)
			}
		}
	}

	// Fetch workflow states
	ui.ShowInfo("Fetching workflow states...")
	states, err := linearClient.GetWorkflowStates(selectedTeam.ID)
	if err != nil {
		ui.ShowWarning(fmt.Sprintf("Could not fetch workflow states: %v", err))
		states = []LinearWorkflowState{}
	}

	// Select state
	var selectedStateID string
	if len(states) > 0 {
		stateOptions := make([]string, len(states))
		for i, state := range states {
			stateType := ""
			switch state.Type {
			case "triage":
				stateType = " (Triage)"
			case "backlog":
				stateType = " (Backlog)"
			case "unstarted":
				stateType = " (Todo)"
			case "started":
				stateType = " (In Progress)"
			case "completed":
				stateType = " (Done)"
			case "canceled":
				stateType = " (Canceled)"
			}
			stateOptions[i] = fmt.Sprintf("%s%s", state.Name, stateType)
		}

		stateChoice, err := ui.SelectFromList("Select initial state", stateOptions)
		if err != nil {
			if err.Error() == "cancelled" {
				return types.ErrNavigateBack
			}
			return err
		}
		selectedStateID = states[stateChoice].ID
	}

	// Show summary
	separator := strings.Repeat("─", 60)
	fmt.Println("\n" + separator)
	fmt.Println(lipgloss.NewStyle().Bold(true).Render("Issue Summary:"))
	fmt.Println(separator)

	fmt.Printf("Type: %s\n", issueType)
	fmt.Printf("Title: %s\n", title)
	fmt.Printf("Team: %s\n", selectedTeam.Name)
	if selectedProjectID != "" {
		for _, p := range projects {
			if p.ID == selectedProjectID {
				fmt.Printf("Project: %s\n", p.Name)
				break
			}
		}
	}
	fmt.Printf("Priority: %s\n", priorities[priorityChoice])
	fmt.Printf("Labels: %s\n", strings.Join(labels, ", "))
	if description != "" {
		fmt.Println("\nDescription:")
		fmt.Println(description)
	}
	fmt.Println(separator)

	if !ui.GetConfirmation("Create this issue in Linear?") {
		return types.ErrNavigateBack
	}

	// Create labels
	ui.ShowInfo("Creating labels...")
	var labelIDs []string
	for _, label := range labels {
		labelID, err := linearClient.GetOrCreateLabel(selectedTeam.ID, label, m.getLabelColor(label))
		if err != nil {
			ui.ShowWarning(fmt.Sprintf("Failed to create label '%s': %v", label, err))
			continue
		}
		labelIDs = append(labelIDs, labelID)
	}

	// Create issue
	ui.ShowInfo("Creating issue in Linear...")
	issue, err := linearClient.CreateIssue(
		selectedTeam.ID,
		selectedProjectID,
		title,
		description,
		labelIDs,
		priorityChoice,
		selectedStateID,
	)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to create issue: %v", err))
		return nil
	}

	ui.ShowSuccess(fmt.Sprintf("Issue created successfully!\nURL: %s", issue.URL))
	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
	return nil
}

// configureProjects handles project configuration
func (m *Module) configureProjects(cfg *config.Config) error {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	fmt.Println(titleStyle.Render("Configure Project Mappings"))

	// Initialize clients
	sentryClient := NewSentryClient(cfg.Sentry.APIKey, cfg.Sentry.BaseURL)
	linearClient := NewLinearClient(cfg.Linear.APIKey)

	// Fetch Sentry projects
	ui.ShowInfo("Fetching Sentry projects...")
	sentryProjects, err := sentryClient.GetProjects()
	if err != nil {
		return fmt.Errorf("failed to fetch Sentry projects: %w", err)
	}

	if len(sentryProjects) == 0 {
		ui.ShowWarning("No Sentry projects found. Please create projects in Sentry first.")
		return types.ErrNavigateBack
	}

	// Fetch Linear teams
	ui.ShowInfo("Fetching Linear teams...")
	linearTeams, err := linearClient.GetTeams()
	if err != nil {
		return fmt.Errorf("failed to fetch Linear teams: %w", err)
	}

	if len(linearTeams) == 0 {
		ui.ShowWarning("No Linear teams found. Please create teams in Linear first.")
		return types.ErrNavigateBack
	}

	// Show current mappings
	if len(cfg.Sentry.Projects) > 0 {
		fmt.Println("\nCurrent Project Mappings:")
		for name, project := range cfg.Sentry.Projects {
			linearProj, exists := cfg.Linear.Projects[project.LinearProjectID]
			if exists {
				fmt.Printf("  • %s (%s/%s) → %s\n",
					name, project.OrganizationSlug, project.ProjectSlug, linearProj.ProjectName)
			} else {
				fmt.Printf("  • %s (%s/%s) → [Not Configured]\n",
					name, project.OrganizationSlug, project.ProjectSlug)
			}
		}
		fmt.Println()
	}

	// Menu for project configuration
	for {
		options := []string{
			"Add New Project Mapping",
			"Remove Project Mapping",
			"Back",
		}

		choice, err := ui.SelectFromList("Project Configuration", options)
		if err != nil {
			return err
		}

		switch choice {
		case 0: // Add new mapping
			if err := m.addProjectMapping(cfg, sentryProjects, linearTeams, sentryClient, linearClient); err != nil {
				if err == types.ErrNavigateBack {
					continue
				}
				return err
			}
		case 1: // Remove mapping
			if err := m.removeProjectMapping(cfg); err != nil {
				if err == types.ErrNavigateBack {
					continue
				}
				return err
			}
		case 2: // Back
			return types.ErrNavigateBack
		}
	}
}

// addProjectMapping adds a new project mapping
func (m *Module) addProjectMapping(cfg *config.Config, sentryProjects []SentryProject, linearTeams []LinearTeam, sentryClient *SentryClient, linearClient *LinearClient) error {
	// Select Sentry project
	sentryOptions := make([]string, len(sentryProjects))
	for i, proj := range sentryProjects {
		orgSlug := proj.OrganizationSlug
		if orgSlug == "" && proj.Organization.Slug != "" {
			orgSlug = proj.Organization.Slug
		}
		sentryOptions[i] = fmt.Sprintf("%s (%s/%s)", proj.Name, orgSlug, proj.Slug)
	}

	sentryChoice, err := ui.SelectFromList("Select Sentry Project", sentryOptions)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	selectedSentryProject := sentryProjects[sentryChoice]

	// Select Linear team
	teamOptions := make([]string, len(linearTeams))
	for i, team := range linearTeams {
		teamOptions[i] = fmt.Sprintf("%s (%s)", team.Name, team.Key)
	}

	teamChoice, err := ui.SelectFromList("Select Linear Team", teamOptions)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	selectedTeam := linearTeams[teamChoice]

	// Fetch Linear projects for the selected team
	ui.ShowInfo("Fetching Linear projects...")
	linearProjects, err := linearClient.GetProjects(selectedTeam.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch Linear projects: %w", err)
	}

	if len(linearProjects) == 0 {
		ui.ShowWarning("No projects found in the selected team. Please create projects in Linear first.")
		return types.ErrNavigateBack
	}

	// Select Linear project
	projectOptions := make([]string, len(linearProjects))
	for i, proj := range linearProjects {
		projectOptions[i] = proj.Name
	}

	projectChoice, err := ui.SelectFromList("Select Linear Project", projectOptions)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	selectedLinearProject := linearProjects[projectChoice]

	// Get mapping name
	mappingName, err := ui.GetInput(
		"Enter a name for this mapping",
		fmt.Sprintf("%s-%s", selectedSentryProject.Slug, selectedLinearProject.Name),
		false,
		nil,
	)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	// Configure default labels
	labelsInput, err := ui.GetInput(
		"Enter default labels (comma-separated)",
		"bug,sentry",
		false,
		nil,
	)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	labels := []string{}
	if labelsInput != "" {
		for _, label := range strings.Split(labelsInput, ",") {
			labels = append(labels, strings.TrimSpace(label))
		}
	}

	// Save the mapping
	if cfg.Sentry.Projects == nil {
		cfg.Sentry.Projects = make(map[string]config.SentryProject)
	}
	if cfg.Linear.Projects == nil {
		cfg.Linear.Projects = make(map[string]config.LinearProject)
	}

	cfg.Sentry.Projects[mappingName] = config.SentryProject{
		OrganizationSlug: selectedSentryProject.OrganizationSlug,
		ProjectSlug:      selectedSentryProject.Slug,
		LinearProjectID:  selectedLinearProject.ID,
	}

	cfg.Linear.Projects[selectedLinearProject.ID] = config.LinearProject{
		TeamID:      selectedTeam.ID,
		ProjectID:   selectedLinearProject.ID,
		ProjectName: selectedLinearProject.Name,
		Labels:      labels,
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	ui.ShowSuccess(fmt.Sprintf("Project mapping '%s' created successfully!", mappingName))
	return nil
}

// removeProjectMapping removes an existing project mapping
func (m *Module) removeProjectMapping(cfg *config.Config) error {
	if len(cfg.Sentry.Projects) == 0 {
		ui.ShowWarning("No project mappings configured.")
		return types.ErrNavigateBack
	}

	// Build list of mappings
	var mappingNames []string
	for name := range cfg.Sentry.Projects {
		mappingNames = append(mappingNames, name)
	}

	// Select mapping to remove
	choice, err := ui.SelectFromList("Select mapping to remove", mappingNames)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	selectedMapping := mappingNames[choice]

	// Confirm deletion
	if !ui.GetConfirmation(fmt.Sprintf("Are you sure you want to remove the mapping '%s'?", selectedMapping)) {
		return types.ErrNavigateBack
	}

	// Remove the mapping
	project := cfg.Sentry.Projects[selectedMapping]
	delete(cfg.Sentry.Projects, selectedMapping)
	delete(cfg.Linear.Projects, project.LinearProjectID)

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	ui.ShowSuccess(fmt.Sprintf("Project mapping '%s' removed successfully!", selectedMapping))
	return nil
}

// syncBugs handles the bug syncing process
func (m *Module) syncBugs(cfg *config.Config) error {
	if len(cfg.Sentry.Projects) == 0 {
		ui.ShowWarning("No project mappings configured. Please configure projects first.")
		return types.ErrNavigateBack
	}

	// Select project mapping
	var mappingNames []string
	for name := range cfg.Sentry.Projects {
		mappingNames = append(mappingNames, name)
	}

	mappingChoice, err := ui.SelectFromList("Select project to sync bugs from", mappingNames)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	selectedMapping := mappingNames[mappingChoice]
	sentryProject := cfg.Sentry.Projects[selectedMapping]
	linearProject := cfg.Linear.Projects[sentryProject.LinearProjectID]

	// Validate Sentry project configuration
	if sentryProject.OrganizationSlug == "" {
		ui.ShowError("Sentry organization slug is missing in the configuration!")
		ui.ShowInfo("Please reconfigure the project mapping with the correct organization slug.")
		fmt.Println("\nPress Enter to continue...")
		fmt.Scanln()
		return types.ErrNavigateBack
	}

	// Initialize clients
	sentryClient := NewSentryClient(cfg.Sentry.APIKey, cfg.Sentry.BaseURL)
	linearClient := NewLinearClient(cfg.Linear.APIKey)

	// Fetch unresolved issues from Sentry
	ui.ShowInfo("Fetching unresolved bugs from Sentry...")
	issues, err := sentryClient.GetUnresolvedIssues(sentryProject.OrganizationSlug, sentryProject.ProjectSlug, 5)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to fetch Sentry issues: %v", err))
		ui.ShowInfo("Please check your API key, project configuration, and network connection.")
		fmt.Println("\nPress Enter to continue...")
		fmt.Scanln()
		return types.ErrNavigateBack
	}

	if len(issues) == 0 {
		ui.ShowSuccess("No unresolved bugs found in Sentry!")
		return types.ErrNavigateBack
	}

	// Display issues
	fmt.Println("\nUnresolved Bugs:")
	issueOptions := make([]string, len(issues))
	for i, issue := range issues {
		issueOptions[i] = fmt.Sprintf("[%s] %s (Level: %s, Count: %s, Users: %d)",
			issue.ShortID, issue.Title, issue.Level, issue.Count, issue.UserCount)
	}

	issueChoice, err := ui.SelectFromList("Select bug to sync to Linear", issueOptions)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	selectedIssue := issues[issueChoice]

	// Fetch latest event for stack trace information
	ui.ShowInfo("Fetching detailed error information...")
	var event *SentryEvent
	event, err = sentryClient.GetLatestEvent(selectedIssue.ID)
	if err != nil {
		ui.ShowWarning(fmt.Sprintf("Could not fetch detailed event data: %v", err))
		// Continue without event data
		event = nil
	}

	// Prepare bug details for Linear
	bugDetails := m.prepareBugDetails(selectedIssue, event)

	// Show bug details and confirm
	separator := strings.Repeat("─", 60)
	fmt.Println("\n" + separator)
	fmt.Println(lipgloss.NewStyle().Bold(true).Render("Bug Details to be Created in Linear:"))
	fmt.Println(separator)

	fmt.Printf("Title: %s\n", bugDetails.Title)
	fmt.Printf("Priority: %s\n", m.getPriorityName(bugDetails.Priority))
	fmt.Printf("Labels: %s\n", strings.Join(append(linearProject.Labels, m.getSentryLabels(selectedIssue)...), ", "))
	fmt.Println("\nDescription Preview:")
	fmt.Println(bugDetails.Description[:min(500, len(bugDetails.Description))] + "...")

	fmt.Println(separator)

	if !ui.GetConfirmation("Create this bug in Linear?") {
		return types.ErrNavigateBack
	}

	// Fetch workflow states
	ui.ShowInfo("Fetching workflow states...")
	states, err := linearClient.GetWorkflowStates(linearProject.TeamID)
	if err != nil {
		ui.ShowWarning(fmt.Sprintf("Could not fetch workflow states: %v", err))
		states = []LinearWorkflowState{}
	}

	// Let user select desired state
	var selectedStateID string
	if len(states) > 0 {
		stateOptions := make([]string, len(states))
		for i, state := range states {
			// Show state type to help user choose
			stateType := ""
			switch state.Type {
			case "triage":
				stateType = " (Triage)"
			case "backlog":
				stateType = " (Backlog)"
			case "unstarted":
				stateType = " (Todo)"
			case "started":
				stateType = " (In Progress)"
			case "completed":
				stateType = " (Done)"
			case "canceled":
				stateType = " (Canceled)"
			}
			stateOptions[i] = fmt.Sprintf("%s%s", state.Name, stateType)
		}

		fmt.Println("\nSelect the initial state for the issue:")
		stateChoice, err := ui.SelectFromList("Select issue state", stateOptions)
		if err != nil {
			if err.Error() == "cancelled" {
				// Use default state if user cancels
				ui.ShowInfo("Using default state (Triage)")
			} else {
				return err
			}
		} else {
			selectedStateID = states[stateChoice].ID
		}
	}

	// Create labels in Linear
	ui.ShowInfo("Creating labels in Linear...")
	var labelIDs []string
	allLabels := append(linearProject.Labels, m.getSentryLabels(selectedIssue)...)

	for _, label := range allLabels {
		labelID, err := linearClient.GetOrCreateLabel(linearProject.TeamID, label, m.getLabelColor(label))
		if err != nil {
			ui.ShowWarning(fmt.Sprintf("Failed to create label '%s': %v", label, err))
			continue
		}
		labelIDs = append(labelIDs, labelID)
	}

	// Create issue in Linear
	ui.ShowInfo("Creating issue in Linear...")
	linearIssue, err := linearClient.CreateIssue(
		linearProject.TeamID,
		linearProject.ProjectID,
		bugDetails.Title,
		bugDetails.Description,
		labelIDs,
		bugDetails.Priority,
		selectedStateID,
	)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to create issue in Linear: %v", err))
		fmt.Println("\nPress Enter to continue...")
		fmt.Scanln()
		return types.ErrNavigateBack
	}

	ui.ShowSuccess(fmt.Sprintf("Bug successfully created in Linear!\nURL: %s", linearIssue.URL))
	return nil
}

// BugDetails holds the prepared bug information for Linear
type BugDetails struct {
	Title       string
	Description string
	Priority    int
}

// prepareBugDetails prepares bug details from Sentry issue for Linear
func (m *Module) prepareBugDetails(issue SentryIssue, event *SentryEvent) BugDetails {
	// Build description with all relevant information
	var description strings.Builder

	description.WriteString("## Sentry Bug Report\n\n")

	// Basic information
	description.WriteString(fmt.Sprintf("**Sentry Issue:** [%s](%s)\n", issue.ShortID, issue.Permalink))
	description.WriteString(fmt.Sprintf("**Level:** %s\n", issue.Level))
	description.WriteString(fmt.Sprintf("**Platform:** %s\n", issue.Platform))
	description.WriteString(fmt.Sprintf("**First Seen:** %s\n", issue.FirstSeen.Format("2006-01-02 15:04:05")))
	description.WriteString(fmt.Sprintf("**Last Seen:** %s\n", issue.LastSeen.Format("2006-01-02 15:04:05")))
	description.WriteString(fmt.Sprintf("**Occurrences:** %s\n", issue.Count))
	description.WriteString(fmt.Sprintf("**Users Affected:** %d\n\n", issue.UserCount))

	// Error details
	description.WriteString("## Error Details\n\n")
	description.WriteString(fmt.Sprintf("**Culprit:** `%s`\n\n", issue.Culprit))

	// Metadata
	if len(issue.Metadata) > 0 {
		description.WriteString("## Additional Information\n\n")

		// Extract common metadata fields
		if value, ok := issue.Metadata["value"].(string); ok && value != "" {
			description.WriteString(fmt.Sprintf("**Error Message:** %s\n\n", value))
		}

		if typeInfo, ok := issue.Metadata["type"].(string); ok && typeInfo != "" {
			description.WriteString(fmt.Sprintf("**Error Type:** `%s`\n\n", typeInfo))
		}

		if filename, ok := issue.Metadata["filename"].(string); ok && filename != "" {
			description.WriteString(fmt.Sprintf("**File:** `%s`\n", filename))
		}

		if function, ok := issue.Metadata["function"].(string); ok && function != "" {
			description.WriteString(fmt.Sprintf("**Function:** `%s`\n", function))
		}

		if lineNo, ok := issue.Metadata["lineNo"].(float64); ok && lineNo > 0 {
			description.WriteString(fmt.Sprintf("**Line Number:** %d\n", int(lineNo)))
		}
	}

	// Add stack trace if event data is available
	if event != nil && len(event.Exception.Values) > 0 {
		description.WriteString("\n## Stack Trace\n\n")

		for _, exception := range event.Exception.Values {
			if exception.Type != "" || exception.Value != "" {
				description.WriteString(fmt.Sprintf("**Exception:** `%s: %s`\n\n", exception.Type, exception.Value))
			}

			if len(exception.Stacktrace.Frames) > 0 {
				description.WriteString("```\n")
				// Reverse the frames to show most recent first
				frames := exception.Stacktrace.Frames
				for i := len(frames) - 1; i >= 0; i-- {
					frame := frames[i]
					if frame.InApp {
						description.WriteString(fmt.Sprintf("  at %s in %s:%d:%d\n",
							frame.Function,
							frame.Filename,
							frame.LineNo,
							frame.ColNo))

						// Add code context if available
						if len(frame.Context) > 0 && frame.LineNo > 0 {
							description.WriteString(fmt.Sprintf("     %s\n", frame.AbsPath))
						}
					}
				}
				description.WriteString("```\n\n")
			}
		}

		// Add most recent error location prominently
		if len(event.Exception.Values) > 0 && len(event.Exception.Values[0].Stacktrace.Frames) > 0 {
			// Find the topmost in-app frame
			for i := len(event.Exception.Values[0].Stacktrace.Frames) - 1; i >= 0; i-- {
				frame := event.Exception.Values[0].Stacktrace.Frames[i]
				if frame.InApp && frame.LineNo > 0 {
					description.WriteString(fmt.Sprintf("**Error Location:** `%s:%d` in function `%s`\n\n",
						frame.Filename,
						frame.LineNo,
						frame.Function))
					break
				}
			}
		}
	}

	// Link back to Sentry
	description.WriteString(fmt.Sprintf("\n---\n\n[View in Sentry](%s)", issue.Permalink))

	// Determine priority based on level and impact
	priority := m.calculatePriority(issue)

	return BugDetails{
		Title:       fmt.Sprintf("[Sentry %s] %s", issue.ShortID, issue.Title),
		Description: description.String(),
		Priority:    priority,
	}
}

// calculatePriority calculates Linear priority based on Sentry issue data
func (m *Module) calculatePriority(issue SentryIssue) int {
	// Linear priorities: 0 = No priority, 1 = Urgent, 2 = High, 3 = Medium, 4 = Low

	// Base priority on error level
	basePriority := 3 // Default to Medium

	switch issue.Level {
	case "fatal", "critical":
		basePriority = 1 // Urgent
	case "error":
		basePriority = 2 // High
	case "warning":
		basePriority = 3 // Medium
	case "info", "debug":
		basePriority = 4 // Low
	}

	// Adjust based on user impact
	if issue.UserCount > 100 && basePriority > 1 {
		basePriority = 1 // Upgrade to Urgent if many users affected
	} else if issue.UserCount > 50 && basePriority > 2 {
		basePriority = 2 // Upgrade to High
	}

	return basePriority
}

// getPriorityName returns the human-readable priority name
func (m *Module) getPriorityName(priority int) string {
	switch priority {
	case 1:
		return "Urgent"
	case 2:
		return "High"
	case 3:
		return "Medium"
	case 4:
		return "Low"
	default:
		return "No priority"
	}
}

// getSentryLabels generates labels based on Sentry issue data
func (m *Module) getSentryLabels(issue SentryIssue) []string {
	labels := []string{
		fmt.Sprintf("level:%s", issue.Level),
		fmt.Sprintf("platform:%s", issue.Platform),
	}

	// Add impact label based on user count
	if issue.UserCount > 100 {
		labels = append(labels, "high-impact")
	} else if issue.UserCount > 50 {
		labels = append(labels, "medium-impact")
	}

	return labels
}

// getLabelColor returns a color for the label
func (m *Module) getLabelColor(label string) string {
	// Define some color mappings
	colorMap := map[string]string{
		"bug":           "#e11d48", // Red
		"sentry":        "#8b5cf6", // Purple
		"level:fatal":   "#991b1b", // Dark red
		"level:error":   "#dc2626", // Red
		"level:warning": "#f59e0b", // Amber
		"level:info":    "#3b82f6", // Blue
		"high-impact":   "#ef4444", // Red
		"medium-impact": "#f97316", // Orange
	}

	// Check if we have a predefined color
	if color, exists := colorMap[label]; exists {
		return color
	}

	// Check for prefixes
	if strings.HasPrefix(label, "level:") {
		return "#8b5cf6" // Purple for level labels
	}

	if strings.HasPrefix(label, "platform:") {
		return "#10b981" // Green for platform labels
	}

	// Default color
	return "#6b7280" // Gray
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// resolveIssues handles resolving Sentry issues
func (m *Module) resolveIssues(cfg *config.Config) error {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	fmt.Println(titleStyle.Render("Resolve Sentry Issues"))

	// Check if API is configured
	if cfg.Sentry.APIKey == "" {
		ui.ShowWarning("Sentry API key not configured. Please configure APIs first.")
		return types.ErrNavigateBack
	}

	if len(cfg.Sentry.Projects) == 0 {
		ui.ShowWarning("No project mappings configured. Please configure projects first.")
		return types.ErrNavigateBack
	}

	// Select project
	var mappingNames []string
	for name := range cfg.Sentry.Projects {
		mappingNames = append(mappingNames, name)
	}

	mappingChoice, err := ui.SelectFromList("Select project to view issues from", mappingNames)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	selectedMapping := mappingNames[mappingChoice]
	sentryProject := cfg.Sentry.Projects[selectedMapping]

	// Validate Sentry project configuration
	if sentryProject.OrganizationSlug == "" {
		ui.ShowError("Sentry organization slug is missing in the configuration!")
		return types.ErrNavigateBack
	}

	// Initialize Sentry client
	sentryClient := NewSentryClient(cfg.Sentry.APIKey, cfg.Sentry.BaseURL)

	// Fetch unresolved issues
	ui.ShowInfo("Fetching unresolved issues from Sentry...")
	issues, err := sentryClient.GetUnresolvedIssues(sentryProject.OrganizationSlug, sentryProject.ProjectSlug, 20)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to fetch Sentry issues: %v", err))
		return types.ErrNavigateBack
	}

	if len(issues) == 0 {
		ui.ShowSuccess("No unresolved issues found in Sentry!")
		return types.ErrNavigateBack
	}

	// Display issues with multi-select option
	fmt.Println("\nUnresolved Issues:")
	issueOptions := make([]string, len(issues))
	for i, issue := range issues {
		issueOptions[i] = fmt.Sprintf("[%s] %s (Level: %s, Count: %s, Users: %d, Last seen: %s)",
			issue.ShortID, issue.Title, issue.Level, issue.Count, issue.UserCount, issue.LastSeen)
	}

	// Add option to select multiple or single
	fmt.Println("\nSelect resolution mode:")
	modeOptions := []string{"Resolve Single Issue", "Resolve Multiple Issues", "Back"}
	modeChoice, err := ui.SelectFromList("Resolution Mode", modeOptions)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	switch modeChoice {
	case 0: // Single issue
		issueChoice, err := ui.SelectFromList("Select issue to resolve", issueOptions)
		if err != nil {
			if err.Error() == "cancelled" {
				return types.ErrNavigateBack
			}
			return err
		}

		selectedIssue := issues[issueChoice]

		// Show issue details
		separator := strings.Repeat("─", 60)
		fmt.Println("\n" + separator)
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("Issue Details:"))
		fmt.Println(separator)
		fmt.Printf("ID: %s\n", selectedIssue.ShortID)
		fmt.Printf("Title: %s\n", selectedIssue.Title)
		fmt.Printf("Level: %s\n", selectedIssue.Level)
		fmt.Printf("Count: %s\n", selectedIssue.Count)
		fmt.Printf("Users Affected: %d\n", selectedIssue.UserCount)
		fmt.Printf("First Seen: %s\n", selectedIssue.FirstSeen)
		fmt.Printf("Last Seen: %s\n", selectedIssue.LastSeen)
		fmt.Printf("Permalink: %s\n", selectedIssue.Permalink)
		fmt.Println(separator)

		if !ui.GetConfirmation("Are you sure you want to mark this issue as resolved?") {
			return types.ErrNavigateBack
		}

		ui.ShowInfo("Resolving issue...")
		if err := sentryClient.ResolveIssue(selectedIssue.ID); err != nil {
			ui.ShowError(fmt.Sprintf("Failed to resolve issue: %v", err))
			return nil
		}

		ui.ShowSuccess(fmt.Sprintf("Issue %s resolved successfully!", selectedIssue.ShortID))

	case 1: // Multiple issues
		fmt.Println("\nSelect issues to resolve (select one at a time, choose 'Done' when finished):")

		selected := make(map[int]bool)

		for {
			// Build options list with selection indicators
			currentOptions := []string{"Done Selecting"}
			for i := range issues {
				prefix := "  "
				if selected[i] {
					prefix = "✓ "
				}
				currentOptions = append(currentOptions, fmt.Sprintf("%s%s", prefix, issueOptions[i]))
			}

			fmt.Printf("\n%d issue(s) selected\n", len(selected))

			choice, err := ui.SelectFromList("Select issues to resolve", currentOptions)
			if err != nil {
				if err.Error() == "cancelled" {
					return types.ErrNavigateBack
				}
				return err
			}

			if choice == 0 { // Done selecting
				if len(selected) == 0 {
					ui.ShowWarning("No issues selected")
					continue
				}
				break
			}

			// Toggle selection
			issueIndex := choice - 1
			if selected[issueIndex] {
				delete(selected, issueIndex)
				ui.ShowInfo(fmt.Sprintf("Deselected: %s", issues[issueIndex].ShortID))
			} else {
				selected[issueIndex] = true
				ui.ShowSuccess(fmt.Sprintf("Selected: %s", issues[issueIndex].ShortID))
			}
		}

		// Show summary
		separator := strings.Repeat("─", 60)
		fmt.Println("\n" + separator)
		fmt.Printf("Selected %d issue(s) to resolve:\n", len(selected))
		fmt.Println(separator)
		for idx := range selected {
			fmt.Printf("• %s\n", issueOptions[idx])
		}
		fmt.Println(separator)

		if !ui.GetConfirmation(fmt.Sprintf("Are you sure you want to mark %d issue(s) as resolved?", len(selected))) {
			return types.ErrNavigateBack
		}

		// Resolve selected issues
		resolved := 0
		failed := 0
		for idx := range selected {
			ui.ShowInfo(fmt.Sprintf("Resolving issue %s...", issues[idx].ShortID))
			if err := sentryClient.ResolveIssue(issues[idx].ID); err != nil {
				ui.ShowWarning(fmt.Sprintf("Failed to resolve %s: %v", issues[idx].ShortID, err))
				failed++
			} else {
				resolved++
			}
		}

		if failed > 0 {
			ui.ShowWarning(fmt.Sprintf("Resolved %d issue(s), %d failed", resolved, failed))
		} else {
			ui.ShowSuccess(fmt.Sprintf("Successfully resolved %d issue(s)!", resolved))
		}

	case 2: // Back
		return types.ErrNavigateBack
	}

	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
	return nil
}
