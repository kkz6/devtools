package bugmanager

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/types"
	"github.com/kkz6/devtools/internal/ui"
)

// manageConnections handles connection management between Linear and Sentry
func (m *Module) manageConnections(cfg *config.Config) error {
	for {
		// Build options list
		options := []string{"Add New Connection"}

		// Add existing connections
		for _, conn := range cfg.BugManager.Connections {
			linearName := "Unknown"
			sentryName := "Unknown"

			if linear, ok := cfg.Linear.Instances[conn.LinearInstance]; ok {
				linearName = linear.Name
			}
			if sentry, ok := cfg.Sentry.Instances[conn.SentryInstance]; ok {
				sentryName = sentry.Name
			}

			options = append(options, fmt.Sprintf("%s: %s ↔ %s (%d mappings)",
				conn.Name, sentryName, linearName, len(conn.ProjectMappings)))
		}

		options = append(options, "Back")

		choice, err := ui.SelectFromList("Sentry-Linear Connections", options)
		if err != nil {
			return err
		}

		if choice == 0 {
			// Add new connection
			if err := m.addConnection(cfg); err != nil && err != types.ErrNavigateBack {
				return err
			}
		} else if choice < len(options)-1 {
			// Edit existing connection
			connIndex := choice - 1
			if err := m.editConnection(cfg, connIndex); err != nil && err != types.ErrNavigateBack {
				return err
			}
		} else {
			// Back
			return types.ErrNavigateBack
		}
	}
}

// addConnection adds a new connection between Linear and Sentry
func (m *Module) addConnection(cfg *config.Config) error {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	fmt.Println(titleStyle.Render("Add New Connection"))

	// Check if we have instances
	if len(cfg.Linear.Instances) == 0 {
		ui.ShowError("No Linear instances configured. Please add a Linear instance first.")
		return types.ErrNavigateBack
	}
	if len(cfg.Sentry.Instances) == 0 {
		ui.ShowError("No Sentry instances configured. Please add a Sentry instance first.")
		return types.ErrNavigateBack
	}

	// Get connection name
	name, err := ui.GetInput(
		"Connection name",
		"",
		false,
		func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("name must not be empty")
			}
			// Check for duplicate names
			for _, conn := range cfg.BugManager.Connections {
				if conn.Name == s {
					return fmt.Errorf("connection with name '%s' already exists", s)
				}
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

	// Select Sentry instance
	sentryOptions := make([]string, 0, len(cfg.Sentry.Instances))
	sentryKeys := make([]string, 0, len(cfg.Sentry.Instances))
	for key := range cfg.Sentry.Instances {
		sentryKeys = append(sentryKeys, key)
	}
	sort.Strings(sentryKeys)

	for _, key := range sentryKeys {
		instance := cfg.Sentry.Instances[key]
		sentryOptions = append(sentryOptions, fmt.Sprintf("%s (%s)", instance.Name, key))
	}

	sentryChoice, err := ui.SelectFromList("Select Sentry instance", sentryOptions)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}
	selectedSentryKey := sentryKeys[sentryChoice]

	// Select Linear instance
	linearOptions := make([]string, 0, len(cfg.Linear.Instances))
	linearKeys := make([]string, 0, len(cfg.Linear.Instances))
	for key := range cfg.Linear.Instances {
		linearKeys = append(linearKeys, key)
	}
	sort.Strings(linearKeys)

	for _, key := range linearKeys {
		instance := cfg.Linear.Instances[key]
		linearOptions = append(linearOptions, fmt.Sprintf("%s (%s)", instance.Name, key))
	}

	linearChoice, err := ui.SelectFromList("Select Linear instance", linearOptions)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}
	selectedLinearKey := linearKeys[linearChoice]

	// Create connection
	connection := config.BugManagerConnection{
		Name:            name,
		LinearInstance:  selectedLinearKey,
		SentryInstance:  selectedSentryKey,
		ProjectMappings: []config.BugManagerProjectMapping{},
	}

	cfg.BugManager.Connections = append(cfg.BugManager.Connections, connection)

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	ui.ShowSuccess(fmt.Sprintf("Connection '%s' created successfully!", name))

	// Ask if user wants to add project mappings now
	if ui.GetConfirmation("Would you like to add project mappings now?") {
		return m.editConnection(cfg, len(cfg.BugManager.Connections)-1)
	}

	return nil
}

// editConnection edits an existing connection
func (m *Module) editConnection(cfg *config.Config, index int) error {
	if index < 0 || index >= len(cfg.BugManager.Connections) {
		return fmt.Errorf("invalid connection index")
	}

	conn := &cfg.BugManager.Connections[index]

	for {
		linearName := "Unknown"
		sentryName := "Unknown"

		if linear, ok := cfg.Linear.Instances[conn.LinearInstance]; ok {
			linearName = linear.Name
		}
		if sentry, ok := cfg.Sentry.Instances[conn.SentryInstance]; ok {
			sentryName = sentry.Name
		}

		options := []string{
			fmt.Sprintf("Edit Name (current: %s)", conn.Name),
			fmt.Sprintf("Change Sentry Instance (current: %s)", sentryName),
			fmt.Sprintf("Change Linear Instance (current: %s)", linearName),
			fmt.Sprintf("Manage Project Mappings (%d)", len(conn.ProjectMappings)),
			"Test Connection",
			"Remove Connection",
			"Back",
		}

		choice, err := ui.SelectFromList(fmt.Sprintf("Edit Connection: %s", conn.Name), options)
		if err != nil {
			return err
		}

		switch choice {
		case 0: // Edit name
			name, err := ui.GetInput("Connection name", conn.Name, false, nil)
			if err != nil {
				if err.Error() == "cancelled" {
					continue
				}
				return err
			}
			conn.Name = name
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
			ui.ShowSuccess("Name updated successfully!")

		case 1: // Change Sentry instance
			sentryOptions := make([]string, 0, len(cfg.Sentry.Instances))
			sentryKeys := make([]string, 0, len(cfg.Sentry.Instances))
			for key := range cfg.Sentry.Instances {
				sentryKeys = append(sentryKeys, key)
			}
			sort.Strings(sentryKeys)

			for _, key := range sentryKeys {
				instance := cfg.Sentry.Instances[key]
				sentryOptions = append(sentryOptions, fmt.Sprintf("%s (%s)", instance.Name, key))
			}

			sentryChoice, err := ui.SelectFromList("Select Sentry instance", sentryOptions)
			if err != nil {
				if err.Error() == "cancelled" {
					continue
				}
				return err
			}
			conn.SentryInstance = sentryKeys[sentryChoice]
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
			ui.ShowSuccess("Sentry instance updated successfully!")

		case 2: // Change Linear instance
			linearOptions := make([]string, 0, len(cfg.Linear.Instances))
			linearKeys := make([]string, 0, len(cfg.Linear.Instances))
			for key := range cfg.Linear.Instances {
				linearKeys = append(linearKeys, key)
			}
			sort.Strings(linearKeys)

			for _, key := range linearKeys {
				instance := cfg.Linear.Instances[key]
				linearOptions = append(linearOptions, fmt.Sprintf("%s (%s)", instance.Name, key))
			}

			linearChoice, err := ui.SelectFromList("Select Linear instance", linearOptions)
			if err != nil {
				if err.Error() == "cancelled" {
					continue
				}
				return err
			}
			conn.LinearInstance = linearKeys[linearChoice]
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
			ui.ShowSuccess("Linear instance updated successfully!")

		case 3: // Manage project mappings
			if err := m.manageProjectMappings(cfg, conn); err != nil && err != types.ErrNavigateBack {
				return err
			}

		case 4: // Test connection
			if err := m.testConnection(cfg, conn); err != nil {
				ui.ShowError(fmt.Sprintf("Connection test failed: %v", err))
			} else {
				ui.ShowSuccess("Connection test passed!")
			}

		case 5: // Remove connection
			if ui.GetConfirmation(fmt.Sprintf("Remove connection '%s'?", conn.Name)) {
				// Remove the connection
				cfg.BugManager.Connections = append(
					cfg.BugManager.Connections[:index],
					cfg.BugManager.Connections[index+1:]...,
				)
				if err := config.Save(cfg); err != nil {
					return fmt.Errorf("failed to save configuration: %w", err)
				}
				ui.ShowSuccess("Connection removed successfully!")
				return types.ErrNavigateBack
			}

		case 6: // Back
			return types.ErrNavigateBack
		}
	}
}

// manageProjectMappings manages project mappings for a connection
func (m *Module) manageProjectMappings(cfg *config.Config, conn *config.BugManagerConnection) error {
	// Get instances
	sentryInstance := cfg.Sentry.Instances[conn.SentryInstance]
	linearInstance := cfg.Linear.Instances[conn.LinearInstance]

	if sentryInstance == nil || linearInstance == nil {
		ui.ShowError("Invalid instance configuration")
		return types.ErrNavigateBack
	}

	for {
		// Build options list
		options := []string{"Add New Project Mapping"}

		// Add existing mappings
		for _, mapping := range conn.ProjectMappings {
			options = append(options, fmt.Sprintf("%s/%s → %s",
				mapping.SentryOrganization, mapping.SentryProject, mapping.LinearProjectName))
		}

		options = append(options, "Back")

		choice, err := ui.SelectFromList("Project Mappings", options)
		if err != nil {
			return err
		}

		if choice == 0 {
			// Add new mapping
			if err := m.addProjectMapping(cfg, conn, sentryInstance, linearInstance); err != nil && err != types.ErrNavigateBack {
				return err
			}
		} else if choice < len(options)-1 {
			// Edit/remove existing mapping
			mappingIndex := choice - 1
			if err := m.editProjectMapping(cfg, conn, mappingIndex); err != nil && err != types.ErrNavigateBack {
				return err
			}
		} else {
			// Back
			return types.ErrNavigateBack
		}
	}
}

// addProjectMapping adds a new project mapping
func (m *Module) addProjectMapping(cfg *config.Config, conn *config.BugManagerConnection,
	sentryInstance *config.SentryInstance, linearInstance *config.LinearInstance) error {

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	fmt.Println(titleStyle.Render("Add Project Mapping"))

	// Initialize clients
	sentryClient := NewSentryClient(sentryInstance.APIKey, sentryInstance.BaseURL)
	linearClient := NewLinearClient(linearInstance.APIKey)

	// Fetch Sentry projects
	ui.ShowInfo("Fetching Sentry projects...")
	sentryProjects, err := sentryClient.GetProjects()
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to fetch Sentry projects: %v", err))
		return types.ErrNavigateBack
	}

	if len(sentryProjects) == 0 {
		ui.ShowWarning("No Sentry projects found.")
		return types.ErrNavigateBack
	}

	// Select Sentry project
	sentryOptions := make([]string, len(sentryProjects))
	for i, proj := range sentryProjects {
		sentryOptions[i] = fmt.Sprintf("%s/%s", proj.Organization.Slug, proj.Slug)
	}

	sentryChoice, err := ui.SelectFromList("Select Sentry project", sentryOptions)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	selectedSentryProject := sentryProjects[sentryChoice]

	// Check if this Sentry project is already mapped
	for _, mapping := range conn.ProjectMappings {
		if mapping.SentryOrganization == selectedSentryProject.Organization.Slug &&
			mapping.SentryProject == selectedSentryProject.Slug {
			ui.ShowError("This Sentry project is already mapped in this connection.")
			return types.ErrNavigateBack
		}
	}

	// Fetch Linear teams
	ui.ShowInfo("Fetching Linear teams...")
	linearTeams, err := linearClient.GetTeams()
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to fetch Linear teams: %v", err))
		return types.ErrNavigateBack
	}

	if len(linearTeams) == 0 {
		ui.ShowWarning("No Linear teams found.")
		return types.ErrNavigateBack
	}

	// Select Linear team
	teamOptions := make([]string, len(linearTeams))
	for i, team := range linearTeams {
		teamOptions[i] = fmt.Sprintf("%s (%s)", team.Name, team.Key)
	}

	teamChoice, err := ui.SelectFromList("Select Linear team", teamOptions)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	selectedTeam := linearTeams[teamChoice]

	// Fetch Linear projects for the team
	ui.ShowInfo("Fetching Linear projects...")
	linearProjects, err := linearClient.GetProjects(selectedTeam.ID)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to fetch Linear projects: %v", err))
		return types.ErrNavigateBack
	}

	// Select Linear project (optional)
	var selectedProjectID string
	var selectedProjectName string

	projectOptions := []string{"No Project (Team only)"}
	for _, proj := range linearProjects {
		projectOptions = append(projectOptions, proj.Name)
	}

	projectChoice, err := ui.SelectFromList("Select Linear project (optional)", projectOptions)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	if projectChoice > 0 {
		selectedProjectID = linearProjects[projectChoice-1].ID
		selectedProjectName = linearProjects[projectChoice-1].Name
	} else {
		selectedProjectName = fmt.Sprintf("%s (Team)", selectedTeam.Name)
	}

	// Get default labels
	labelsInput, err := ui.GetInput(
		"Default labels for synced bugs (comma-separated)",
		"bug,sentry",
		false,
		nil,
	)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		labelsInput = "bug,sentry"
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

	// Create mapping
	mapping := config.BugManagerProjectMapping{
		SentryOrganization: selectedSentryProject.Organization.Slug,
		SentryProject:      selectedSentryProject.Slug,
		LinearTeamID:       selectedTeam.ID,
		LinearProjectID:    selectedProjectID,
		LinearProjectName:  selectedProjectName,
		DefaultLabels:      labels,
	}

	conn.ProjectMappings = append(conn.ProjectMappings, mapping)

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	ui.ShowSuccess("Project mapping added successfully!")
	return nil
}

// editProjectMapping edits/removes a project mapping
func (m *Module) editProjectMapping(cfg *config.Config, conn *config.BugManagerConnection, index int) error {
	if index < 0 || index >= len(conn.ProjectMappings) {
		return fmt.Errorf("invalid mapping index")
	}

	mapping := &conn.ProjectMappings[index]

	options := []string{
		"Edit Default Labels",
		"Remove Mapping",
		"Back",
	}

	choice, err := ui.SelectFromList(fmt.Sprintf("Edit Mapping: %s/%s",
		mapping.SentryOrganization, mapping.SentryProject), options)
	if err != nil {
		return err
	}

	switch choice {
	case 0: // Edit labels
		currentLabels := strings.Join(mapping.DefaultLabels, ", ")
		labelsInput, err := ui.GetInput("Default labels (comma-separated)", currentLabels, false, nil)
		if err != nil {
			if err.Error() == "cancelled" {
				return types.ErrNavigateBack
			}
			return err
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
		mapping.DefaultLabels = labels

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}
		ui.ShowSuccess("Labels updated successfully!")

	case 1: // Remove mapping
		if ui.GetConfirmation("Remove this project mapping?") {
			conn.ProjectMappings = append(
				conn.ProjectMappings[:index],
				conn.ProjectMappings[index+1:]...,
			)
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
			ui.ShowSuccess("Mapping removed successfully!")
			return types.ErrNavigateBack
		}

	case 2: // Back
		return types.ErrNavigateBack
	}

	return nil
}

// testConnection tests a connection between Linear and Sentry
func (m *Module) testConnection(cfg *config.Config, conn *config.BugManagerConnection) error {
	sentryInstance := cfg.Sentry.Instances[conn.SentryInstance]
	linearInstance := cfg.Linear.Instances[conn.LinearInstance]

	if sentryInstance == nil || linearInstance == nil {
		return fmt.Errorf("invalid instance configuration")
	}

	ui.ShowInfo("Testing Sentry connection...")
	sentryClient := NewSentryClient(sentryInstance.APIKey, sentryInstance.BaseURL)
	_, err := sentryClient.GetProjects()
	if err != nil {
		return fmt.Errorf("sentry connection failed: %w", err)
	}
	ui.ShowSuccess("Sentry connection successful!")

	ui.ShowInfo("Testing Linear connection...")
	linearClient := NewLinearClient(linearInstance.APIKey)
	_, err = linearClient.GetTeams()
	if err != nil {
		return fmt.Errorf("linear connection failed: %w", err)
	}
	ui.ShowSuccess("Linear connection successful!")

	return nil
}
