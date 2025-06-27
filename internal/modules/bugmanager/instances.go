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

// manageInstances handles instance management for Linear and Sentry
func (m *Module) manageInstances(cfg *config.Config) error {
	for {
		options := []string{
			"Manage Linear Instances",
			"Manage Sentry Instances",
			"Back",
		}

		choice, err := ui.SelectFromList("Instance Management", options)
		if err != nil {
			return err
		}

		switch choice {
		case 0: // Linear instances
			if err := m.manageLinearInstances(cfg); err != nil && err != types.ErrNavigateBack {
				return err
			}
		case 1: // Sentry instances
			if err := m.manageSentryInstances(cfg); err != nil && err != types.ErrNavigateBack {
				return err
			}
		case 2: // Back
			return types.ErrNavigateBack
		}
	}
}

// manageLinearInstances handles Linear instance management
func (m *Module) manageLinearInstances(cfg *config.Config) error {
	for {
		// Build options list
		options := []string{"Add New Linear Instance"}

		// Add existing instances
		instanceKeys := make([]string, 0, len(cfg.Linear.Instances))
		for key := range cfg.Linear.Instances {
			instanceKeys = append(instanceKeys, key)
		}

		sort.Strings(instanceKeys)

		for _, key := range instanceKeys {
			instance := cfg.Linear.Instances[key]
			options = append(options, fmt.Sprintf("Edit: %s (%s)", instance.Name, key))
		}

		options = append(options, "Back")

		choice, err := ui.SelectFromList("Linear Instances", options)
		if err != nil {
			return err
		}

		if choice == 0 {
			// Add new instance
			if err := m.addLinearInstance(cfg); err != nil && err != types.ErrNavigateBack {
				return err
			}
		} else if choice < len(options)-1 {
			// Edit existing instance
			instanceKey := instanceKeys[choice-1]
			if err := m.editLinearInstance(cfg, instanceKey); err != nil && err != types.ErrNavigateBack {
				return err
			}
		} else {
			// Back
			return types.ErrNavigateBack
		}
	}
}

// addLinearInstance adds a new Linear instance
func (m *Module) addLinearInstance(cfg *config.Config) error {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	fmt.Println(titleStyle.Render("Add Linear Instance"))

	// Get instance key
	key, err := ui.GetInput(
		"Instance key (e.g., 'work', 'personal')",
		"",
		false,
		func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("key must not be empty")
			}
			if strings.Contains(s, " ") {
				return fmt.Errorf("key must not contain spaces")
			}
			if _, exists := cfg.Linear.Instances[s]; exists {
				return fmt.Errorf("instance with key '%s' already exists", s)
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

	// Get instance name
	name, err := ui.GetInput(
		"Instance name (display name)",
		"",
		false,
		func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("name must not be empty")
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

	// Get API key
	apiKey, err := ui.GetInput("Linear API Key", "", true, nil)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	// Test connection
	ui.ShowInfo("Testing Linear API connection...")
	linearClient := NewLinearClient(apiKey)
	teams, err := linearClient.GetTeams()
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to connect to Linear: %v", err))
		if !ui.GetConfirmation("Continue anyway?") {
			return types.ErrNavigateBack
		}
	} else {
		ui.ShowSuccess(fmt.Sprintf("Connected successfully! Found %d teams.", len(teams)))
	}

	// Save instance
	cfg.Linear.Instances[key] = &config.LinearInstance{
		Name:   name,
		APIKey: apiKey,
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	ui.ShowSuccess(fmt.Sprintf("Linear instance '%s' added successfully!", name))
	return nil
}

// editLinearInstance edits an existing Linear instance
func (m *Module) editLinearInstance(cfg *config.Config, key string) error {
	instance := cfg.Linear.Instances[key]
	if instance == nil {
		return fmt.Errorf("instance not found")
	}

	for {
		options := []string{
			fmt.Sprintf("Edit Name (current: %s)", instance.Name),
			"Update API Key",
			"Test Connection",
			"Remove Instance",
			"Back",
		}

		choice, err := ui.SelectFromList(fmt.Sprintf("Edit Linear Instance: %s", instance.Name), options)
		if err != nil {
			return err
		}

		switch choice {
		case 0: // Edit name
			name, err := ui.GetInput("Instance name", instance.Name, false, nil)
			if err != nil {
				if err.Error() == "cancelled" {
					continue
				}
				return err
			}
			instance.Name = name
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
			ui.ShowSuccess("Name updated successfully!")

		case 1: // Update API key
			apiKey, err := ui.GetInput("Linear API Key", "", true, nil)
			if err != nil {
				if err.Error() == "cancelled" {
					continue
				}
				return err
			}
			instance.APIKey = apiKey
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
			ui.ShowSuccess("API key updated successfully!")

		case 2: // Test connection
			ui.ShowInfo("Testing Linear API connection...")
			linearClient := NewLinearClient(instance.APIKey)
			teams, err := linearClient.GetTeams()
			if err != nil {
				ui.ShowError(fmt.Sprintf("Connection failed: %v", err))
			} else {
				ui.ShowSuccess(fmt.Sprintf("Connected successfully! Found %d teams.", len(teams)))
			}

		case 3: // Remove instance
			// Check if instance is used in any connections
			isUsed := false
			for _, conn := range cfg.BugManager.Connections {
				if conn.LinearInstance == key {
					isUsed = true
					break
				}
			}

			if isUsed {
				ui.ShowError("Cannot remove instance: it is used in one or more connections")
				continue
			}

			if ui.GetConfirmation(fmt.Sprintf("Remove Linear instance '%s'?", instance.Name)) {
				delete(cfg.Linear.Instances, key)
				if err := config.Save(cfg); err != nil {
					return fmt.Errorf("failed to save configuration: %w", err)
				}
				ui.ShowSuccess("Instance removed successfully!")
				return types.ErrNavigateBack
			}

		case 4: // Back
			return types.ErrNavigateBack
		}
	}
}

// manageSentryInstances handles Sentry instance management
func (m *Module) manageSentryInstances(cfg *config.Config) error {
	for {
		// Build options list
		options := []string{"Add New Sentry Instance"}

		// Add existing instances
		instanceKeys := make([]string, 0, len(cfg.Sentry.Instances))
		for key := range cfg.Sentry.Instances {
			instanceKeys = append(instanceKeys, key)
		}

		sort.Strings(instanceKeys)

		for _, key := range instanceKeys {
			instance := cfg.Sentry.Instances[key]
			options = append(options, fmt.Sprintf("Edit: %s (%s)", instance.Name, key))
		}

		options = append(options, "Back")

		choice, err := ui.SelectFromList("Sentry Instances", options)
		if err != nil {
			return err
		}

		if choice == 0 {
			// Add new instance
			if err := m.addSentryInstance(cfg); err != nil && err != types.ErrNavigateBack {
				return err
			}
		} else if choice < len(options)-1 {
			// Edit existing instance
			instanceKey := instanceKeys[choice-1]
			if err := m.editSentryInstance(cfg, instanceKey); err != nil && err != types.ErrNavigateBack {
				return err
			}
		} else {
			// Back
			return types.ErrNavigateBack
		}
	}
}

// addSentryInstance adds a new Sentry instance
func (m *Module) addSentryInstance(cfg *config.Config) error {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	fmt.Println(titleStyle.Render("Add Sentry Instance"))

	// Get instance key
	key, err := ui.GetInput(
		"Instance key (e.g., 'work', 'personal')",
		"",
		false,
		func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("key must not be empty")
			}
			if strings.Contains(s, " ") {
				return fmt.Errorf("key must not contain spaces")
			}
			if _, exists := cfg.Sentry.Instances[s]; exists {
				return fmt.Errorf("instance with key '%s' already exists", s)
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

	// Get instance name
	name, err := ui.GetInput(
		"Instance name (display name)",
		"",
		false,
		func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("name must not be empty")
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

	// Get base URL
	baseURL, err := ui.GetInput(
		"Sentry Base URL",
		"https://sentry.io/api/0",
		false,
		func(s string) error {
			if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
				return fmt.Errorf("URL must start with http:// or https://")
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

	// Get API key
	apiKey, err := ui.GetInput("Sentry API Key", "", true, nil)
	if err != nil {
		if err.Error() == "cancelled" {
			return types.ErrNavigateBack
		}
		return err
	}

	// Test connection
	ui.ShowInfo("Testing Sentry API connection...")
	sentryClient := NewSentryClient(apiKey, baseURL)
	projects, err := sentryClient.GetProjects()
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to connect to Sentry: %v", err))
		if !ui.GetConfirmation("Continue anyway?") {
			return types.ErrNavigateBack
		}
	} else {
		ui.ShowSuccess(fmt.Sprintf("Connected successfully! Found %d projects.", len(projects)))
	}

	// Save instance
	cfg.Sentry.Instances[key] = &config.SentryInstance{
		Name:    name,
		APIKey:  apiKey,
		BaseURL: baseURL,
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	ui.ShowSuccess(fmt.Sprintf("Sentry instance '%s' added successfully!", name))
	return nil
}

// editSentryInstance edits an existing Sentry instance
func (m *Module) editSentryInstance(cfg *config.Config, key string) error {
	instance := cfg.Sentry.Instances[key]
	if instance == nil {
		return fmt.Errorf("instance not found")
	}

	for {
		options := []string{
			fmt.Sprintf("Edit Name (current: %s)", instance.Name),
			fmt.Sprintf("Edit Base URL (current: %s)", instance.BaseURL),
			"Update API Key",
			"Test Connection",
			"Remove Instance",
			"Back",
		}

		choice, err := ui.SelectFromList(fmt.Sprintf("Edit Sentry Instance: %s", instance.Name), options)
		if err != nil {
			return err
		}

		switch choice {
		case 0: // Edit name
			name, err := ui.GetInput("Instance name", instance.Name, false, nil)
			if err != nil {
				if err.Error() == "cancelled" {
					continue
				}
				return err
			}
			instance.Name = name
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
			ui.ShowSuccess("Name updated successfully!")

		case 1: // Edit base URL
			baseURL, err := ui.GetInput("Sentry Base URL", instance.BaseURL, false, nil)
			if err != nil {
				if err.Error() == "cancelled" {
					continue
				}
				return err
			}
			instance.BaseURL = baseURL
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
			ui.ShowSuccess("Base URL updated successfully!")

		case 2: // Update API key
			apiKey, err := ui.GetInput("Sentry API Key", "", true, nil)
			if err != nil {
				if err.Error() == "cancelled" {
					continue
				}
				return err
			}
			instance.APIKey = apiKey
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
			ui.ShowSuccess("API key updated successfully!")

		case 3: // Test connection
			ui.ShowInfo("Testing Sentry API connection...")
			sentryClient := NewSentryClient(instance.APIKey, instance.BaseURL)
			projects, err := sentryClient.GetProjects()
			if err != nil {
				ui.ShowError(fmt.Sprintf("Connection failed: %v", err))
			} else {
				ui.ShowSuccess(fmt.Sprintf("Connected successfully! Found %d projects.", len(projects)))
			}

		case 4: // Remove instance
			// Check if instance is used in any connections
			isUsed := false
			for _, conn := range cfg.BugManager.Connections {
				if conn.SentryInstance == key {
					isUsed = true
					break
				}
			}

			if isUsed {
				ui.ShowError("Cannot remove instance: it is used in one or more connections")
				continue
			}

			if ui.GetConfirmation(fmt.Sprintf("Remove Sentry instance '%s'?", instance.Name)) {
				delete(cfg.Sentry.Instances, key)
				if err := config.Save(cfg); err != nil {
					return fmt.Errorf("failed to save configuration: %w", err)
				}
				ui.ShowSuccess("Instance removed successfully!")
				return types.ErrNavigateBack
			}

		case 5: // Back
			return types.ErrNavigateBack
		}
	}
}
