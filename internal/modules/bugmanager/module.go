package bugmanager

import (
	"bufio"
	"fmt"
	"os"
	"sort"
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
	// Show main menu
	return m.showMainMenu(cfg)
}

// showMainMenu displays the main bug manager menu
func (m *Module) showMainMenu(cfg *config.Config) error {
	for {
		options := []string{
			"Sync Bugs from Sentry",
			"Create Manual Issue",
			"Manage Instances",
			"Manage Connections",
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
		case 2: // Manage instances
			if err := m.manageInstances(cfg); err != nil && err != types.ErrNavigateBack {
				return err
			}
		case 3: // Manage connections
			if err := m.manageConnections(cfg); err != nil && err != types.ErrNavigateBack {
				return err
			}
		case 4: // Back
			return types.ErrNavigateBack
		}
	}
}

// createManualIssue handles manual issue creation in Linear
func (m *Module) createManualIssue(cfg *config.Config) error {
	// Check if we have any Linear instances
	if len(cfg.Linear.Instances) == 0 {
		ui.ShowError("No Linear instances configured. Please add a Linear instance first.")
		return nil
	}

	// Select Linear instance
	var selectedInstance *config.LinearInstance
	var selectedInstanceKey string

	if len(cfg.Linear.Instances) == 1 {
		// If only one instance, use it automatically
		for key, instance := range cfg.Linear.Instances {
			selectedInstance = instance
			selectedInstanceKey = key
			break
		}
	} else {
		// Multiple instances, let user choose
		instanceOptions := make([]string, 0, len(cfg.Linear.Instances))
		instanceKeys := make([]string, 0, len(cfg.Linear.Instances))

		for key := range cfg.Linear.Instances {
			instanceKeys = append(instanceKeys, key)
		}
		sort.Strings(instanceKeys)

		for _, key := range instanceKeys {
			instance := cfg.Linear.Instances[key]
			instanceOptions = append(instanceOptions, fmt.Sprintf("%s (%s)", instance.Name, key))
		}

		choice, err := ui.SelectFromList("Select Linear instance", instanceOptions)
		if err != nil {
			if err.Error() == "cancelled" {
				return types.ErrNavigateBack
			}
			return err
		}

		selectedInstanceKey = instanceKeys[choice]
		selectedInstance = cfg.Linear.Instances[selectedInstanceKey]
	}

	// Initialize Linear client
	linearClient := NewLinearClient(selectedInstance.APIKey)

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

	fmt.Printf("Linear Instance: %s\n", selectedInstance.Name)
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

// syncBugs handles the bug syncing process
func (m *Module) syncBugs(cfg *config.Config) error {
	// Check if we have any connections
	if len(cfg.BugManager.Connections) == 0 {
		ui.ShowError("No Sentry-Linear connections configured. Please add a connection first.")
		return nil
	}

	// Select connection
	var selectedConnection *config.BugManagerConnection

	if len(cfg.BugManager.Connections) == 1 {
		// If only one connection, use it automatically
		selectedConnection = &cfg.BugManager.Connections[0]
	} else {
		// Multiple connections, let user choose
		connectionOptions := make([]string, len(cfg.BugManager.Connections))

		for i, conn := range cfg.BugManager.Connections {
			linearName := "Unknown"
			sentryName := "Unknown"

			if linear, ok := cfg.Linear.Instances[conn.LinearInstance]; ok {
				linearName = linear.Name
			}
			if sentry, ok := cfg.Sentry.Instances[conn.SentryInstance]; ok {
				sentryName = sentry.Name
			}

			connectionOptions[i] = fmt.Sprintf("%s: %s → %s (%d mappings)",
				conn.Name, sentryName, linearName, len(conn.ProjectMappings))
		}

		choice, err := ui.SelectFromList("Select connection to sync", connectionOptions)
		if err != nil {
			if err.Error() == "cancelled" {
				return types.ErrNavigateBack
			}
			return err
		}

		selectedConnection = &cfg.BugManager.Connections[choice]
	}

	// Check if connection has project mappings
	if len(selectedConnection.ProjectMappings) == 0 {
		ui.ShowError("Selected connection has no project mappings. Please configure project mappings first.")
		return nil
	}

	// Get instances
	sentryInstance := cfg.Sentry.Instances[selectedConnection.SentryInstance]
	linearInstance := cfg.Linear.Instances[selectedConnection.LinearInstance]

	if sentryInstance == nil || linearInstance == nil {
		ui.ShowError("Invalid instance configuration in connection.")
		return nil
	}

	// Initialize clients
	sentryClient := NewSentryClient(sentryInstance.APIKey, sentryInstance.BaseURL)
	linearClient := NewLinearClient(linearInstance.APIKey)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	fmt.Println(titleStyle.Render(fmt.Sprintf("Sync Bugs: %s", selectedConnection.Name)))

	// Select project mapping if multiple exist
	var selectedMapping *config.BugManagerProjectMapping

	if len(selectedConnection.ProjectMappings) == 1 {
		selectedMapping = &selectedConnection.ProjectMappings[0]
	} else {
		mappingOptions := make([]string, len(selectedConnection.ProjectMappings))
		for i, mapping := range selectedConnection.ProjectMappings {
			mappingOptions[i] = fmt.Sprintf("%s/%s → %s",
				mapping.SentryOrganization, mapping.SentryProject, mapping.LinearProjectName)
		}

		choice, err := ui.SelectFromList("Select project to sync from", mappingOptions)
		if err != nil {
			if err.Error() == "cancelled" {
				return types.ErrNavigateBack
			}
			return err
		}

		selectedMapping = &selectedConnection.ProjectMappings[choice]
	}

	// Fetch unresolved issues from Sentry
	ui.ShowInfo(fmt.Sprintf("Fetching unresolved issues from %s/%s...",
		selectedMapping.SentryOrganization, selectedMapping.SentryProject))

	issues, err := sentryClient.GetUnresolvedIssues(
		selectedMapping.SentryOrganization,
		selectedMapping.SentryProject,
		20, // Fetch up to 20 issues
	)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to fetch issues: %v", err))
		return nil
	}

	if len(issues) == 0 {
		ui.ShowSuccess("No unresolved issues found in Sentry!")
		return nil
	}

	// Display issues for selection
	fmt.Println(fmt.Sprintf("\nFound %d unresolved issues:", len(issues)))
	issueOptions := make([]string, len(issues))
	for i, issue := range issues {
		issueOptions[i] = fmt.Sprintf("[%s] %s (Level: %s, Count: %s, Users: %d)",
			issue.ShortID, issue.Title, issue.Level, issue.Count, issue.UserCount)
	}

	// Select issue to sync
	issueChoice, err := ui.SelectFromList("Select issue to sync to Linear", issueOptions)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	selectedIssue := issues[issueChoice]

	// Get issue details
	ui.ShowInfo("Fetching issue details...")
	issueDetails, err := sentryClient.GetIssueDetails(selectedIssue.ID)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to get issue details: %v", err))
		return nil
	}

	// Get latest event for stack trace
	ui.ShowInfo("Fetching error event details...")
	event, err := sentryClient.GetLatestEvent(selectedIssue.ID)
	if err != nil {
		ui.ShowWarning(fmt.Sprintf("Could not fetch event details: %v", err))
		// Continue without event data
		event = nil
	}

	// Prepare bug details
	bugDetails := m.prepareBugDetails(*issueDetails, event)

	// Show bug preview
	separator := strings.Repeat("─", 60)
	fmt.Println("\n" + separator)
	fmt.Println(lipgloss.NewStyle().Bold(true).Render("Bug Details to be Created in Linear:"))
	fmt.Println(separator)

	fmt.Printf("Title: %s\n", bugDetails.Title)
	fmt.Printf("Priority: %s\n", m.getPriorityName(bugDetails.Priority))
	fmt.Printf("Target: %s\n", selectedMapping.LinearProjectName)
	fmt.Printf("Labels: %s\n", strings.Join(append(selectedMapping.DefaultLabels, m.getSentryLabels(*issueDetails)...), ", "))
	fmt.Println("\nDescription Preview:")
	// Show first 500 chars of description
	descPreview := bugDetails.Description
	if len(descPreview) > 500 {
		descPreview = descPreview[:500] + "..."
	}
	fmt.Println(descPreview)

	fmt.Println(separator)

	if !ui.GetConfirmation("Create this issue in Linear?") {
		return types.ErrNavigateBack
	}

	// Fetch workflow states
	ui.ShowInfo("Fetching workflow states...")
	states, err := linearClient.GetWorkflowStates(selectedMapping.LinearTeamID)
	if err != nil {
		ui.ShowWarning(fmt.Sprintf("Could not fetch workflow states: %v", err))
		states = []LinearWorkflowState{}
	}

	// Select initial state
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

		fmt.Println("\nSelect the initial state for the issue:")
		stateChoice, err := ui.SelectFromList("Select issue state", stateOptions)
		if err != nil {
			if err.Error() == "cancelled" {
				// Use default state if user cancels
				ui.ShowInfo("Using default state")
			} else {
				return err
			}
		} else {
			selectedStateID = states[stateChoice].ID
		}
	}

	// Create labels
	ui.ShowInfo("Creating labels in Linear...")
	var labelIDs []string
	allLabels := append(selectedMapping.DefaultLabels, m.getSentryLabels(*issueDetails)...)

	for _, label := range allLabels {
		labelID, err := linearClient.GetOrCreateLabel(
			selectedMapping.LinearTeamID,
			label,
			m.getLabelColor(label),
		)
		if err != nil {
			ui.ShowWarning(fmt.Sprintf("Failed to create label '%s': %v", label, err))
			continue
		}
		labelIDs = append(labelIDs, labelID)
	}

	// Create Linear issue
	ui.ShowInfo("Creating issue in Linear...")
	linearIssue, err := linearClient.CreateIssue(
		selectedMapping.LinearTeamID,
		selectedMapping.LinearProjectID,
		bugDetails.Title,
		bugDetails.Description,
		labelIDs,
		bugDetails.Priority,
		selectedStateID,
	)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to create Linear issue: %v", err))
		return nil
	}

	ui.ShowSuccess(fmt.Sprintf("Issue created successfully!\nURL: %s", linearIssue.URL))

	// Ask if user wants to resolve in Sentry
	if ui.GetConfirmation("\nMark this issue as resolved in Sentry?") {
		ui.ShowInfo("Resolving issue in Sentry...")
		if err := sentryClient.ResolveIssue(selectedIssue.ID); err != nil {
			ui.ShowError(fmt.Sprintf("Failed to resolve issue in Sentry: %v", err))
		} else {
			ui.ShowSuccess("Issue marked as resolved in Sentry")
		}
	}

	// Ask if user wants to sync another issue
	if ui.GetConfirmation("\nSync another issue from the same project?") {
		return m.syncBugs(cfg)
	}

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
