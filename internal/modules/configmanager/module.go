package configmanager

import (
	"fmt"
	"strings"
	"time"

	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/types"
	"github.com/kkz6/devtools/internal/ui"
)

// Module implements the configuration manager module
type Module struct{}

// New creates a new configuration manager module
func New() *Module {
	return &Module{}
}

// Info returns module information
func (m *Module) Info() types.ModuleInfo {
	return types.ModuleInfo{
		ID:          "config-manager",
		Name:        "Configuration Manager",
		Description: "Manage application configuration settings",
	}
}

// Execute runs the configuration manager
func (m *Module) Execute(cfg *config.Config) error {
	ui.ShowBanner()

	title := ui.GetGradientTitle("⚙️  Configuration Manager")
	fmt.Println(title)
	fmt.Println()

	for {
		options := []string{
			"GitHub Configuration",
			"SSH Configuration",
			"GPG Configuration",
			"Cursor AI Configuration",
			"Sentry Configuration",
			"Linear Configuration",
			"Global Settings",
			"View Configuration Path",
			"Back to main menu",
		}

		choice, err := ui.SelectFromList("Select configuration to manage:", options)
		if err != nil || choice == 8 {
			return types.ErrNavigateBack
		}

		switch choice {
		case 0:
			if err := m.configureGitHub(cfg); err != nil {
				ui.ShowError(fmt.Sprintf("Failed to configure GitHub: %v", err))
			}
		case 1:
			if err := m.configureSSH(cfg); err != nil {
				ui.ShowError(fmt.Sprintf("Failed to configure SSH: %v", err))
			}
		case 2:
			if err := m.configureGPG(cfg); err != nil {
				ui.ShowError(fmt.Sprintf("Failed to configure GPG: %v", err))
			}
		case 3:
			if err := m.configureCursor(cfg); err != nil {
				ui.ShowError(fmt.Sprintf("Failed to configure Cursor: %v", err))
			}
		case 4:
			if err := m.configureSentry(cfg); err != nil {
				ui.ShowError(fmt.Sprintf("Failed to configure Sentry: %v", err))
			}
		case 5:
			if err := m.configureLinear(cfg); err != nil {
				ui.ShowError(fmt.Sprintf("Failed to configure Linear: %v", err))
			}
		case 6:
			if err := m.configureGlobalSettings(cfg); err != nil {
				ui.ShowError(fmt.Sprintf("Failed to configure global settings: %v", err))
			}
		case 7:
			m.showConfigPath()
		}

		// Save configuration after each change
		if choice >= 0 && choice <= 6 {
			err = ui.ShowLoadingAnimation("Saving configuration", func() error {
				return config.Save(cfg)
			})
			if err != nil {
				ui.ShowError(fmt.Sprintf("Failed to save configuration: %v", err))
			} else {
				ui.ShowSuccess("Configuration saved successfully!")
				time.Sleep(1 * time.Second)
			}
		}
	}
}

// configureGitHub handles GitHub configuration
func (m *Module) configureGitHub(cfg *config.Config) error {
	fmt.Println()
	ui.ShowInfo("Configure GitHub settings for API access and authentication")
	fmt.Println()

	// Username
	username, err := ui.GetInput(
		"👤 GitHub Username",
		cfg.GitHub.Username,
		false,
		func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("username cannot be empty")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	cfg.GitHub.Username = username

	// Personal Access Token
	ui.ShowInfo("Create a token at: https://github.com/settings/tokens")
	ui.ShowInfo("Required scopes: write:ssh_signing_key (for SSH key upload)")
	fmt.Println()

	token, err := ui.GetInput(
		"🔑 GitHub Personal Access Token",
		"ghp_...",
		true, // password mode
		func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("token cannot be empty")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	cfg.GitHub.Token = token

	// Email
	email, err := ui.GetInput(
		"📧 GitHub Email",
		cfg.GitHub.Email,
		false,
		func(s string) error {
			if len(s) > 0 && !strings.Contains(s, "@") {
				return fmt.Errorf("please enter a valid email address")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	cfg.GitHub.Email = email

	return nil
}

// configureSSH handles SSH configuration
func (m *Module) configureSSH(cfg *config.Config) error {
	fmt.Println()
	ui.ShowInfo("Configure SSH signing key settings")
	fmt.Println()

	// Key path
	keyPath, err := ui.GetInput(
		"🔐 SSH Signing Key Path",
		cfg.SSH.SigningKeyPath,
		false,
		func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("key path cannot be empty")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	cfg.SSH.SigningKeyPath = keyPath

	// Key comment
	comment, err := ui.GetInput(
		"💬 SSH Key Comment",
		cfg.SSH.KeyComment,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	if comment != "" {
		cfg.SSH.KeyComment = comment
	}

	return nil
}

// configureGPG handles GPG configuration
func (m *Module) configureGPG(cfg *config.Config) error {
	fmt.Println()
	ui.ShowInfo("Configure GPG signing settings")
	fmt.Println()

	// Email
	email, err := ui.GetInput(
		"📧 GPG Email",
		cfg.GPG.Email,
		false,
		func(s string) error {
			if len(s) > 0 && !strings.Contains(s, "@") {
				return fmt.Errorf("please enter a valid email address")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	cfg.GPG.Email = email

	// Key ID (optional)
	keyID, err := ui.GetInput(
		"🔑 GPG Key ID (optional)",
		cfg.GPG.KeyID,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	cfg.GPG.KeyID = keyID

	return nil
}

// configureCursor handles Cursor AI configuration
func (m *Module) configureCursor(cfg *config.Config) error {
	fmt.Println()
	ui.ShowInfo("Configure Cursor AI settings for usage tracking and analysis")
	fmt.Println()

	// API Key
	apiKey, err := ui.GetInput(
		"🔑 Cursor API Key",
		"cur_...",
		true, // password mode
		func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("API key cannot be empty")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	cfg.Cursor.APIKey = apiKey

	// API Endpoint (optional, with default)
	endpoint, err := ui.GetInput(
		"🌐 API Endpoint (press Enter for default)",
		cfg.Cursor.APIEndpoint,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	if endpoint == "" {
		cfg.Cursor.APIEndpoint = "https://api.cursor.sh/v1"
	} else {
		cfg.Cursor.APIEndpoint = endpoint
	}

	// Current Plan
	plans := []string{
		"Free",
		"Pro ($20/month)",
		"Business ($40/month)",
	}

	choice, err := ui.SelectFromList("Select your current Cursor plan:", plans)
	if err != nil {
		return err
	}

	switch choice {
	case 0:
		cfg.Cursor.CurrentPlan = "free"
	case 1:
		cfg.Cursor.CurrentPlan = "pro"
	case 2:
		cfg.Cursor.CurrentPlan = "business"
	}

	return nil
}

// configureSentry handles Sentry configuration
func (m *Module) configureSentry(cfg *config.Config) error {
	fmt.Println()
	ui.ShowInfo("Configure Sentry settings for bug tracking integration")
	ui.ShowWarning("Note: For multiple Sentry instances, use the Bug Manager's 'Manage Instances' option")
	fmt.Println()

	// Check if instances exist
	if len(cfg.Sentry.Instances) > 0 {
		ui.ShowInfo("You have configured Sentry instances in Bug Manager:")
		for key, instance := range cfg.Sentry.Instances {
			fmt.Printf("  • %s (%s)\n", instance.Name, key)
		}
		fmt.Println()

		if !ui.GetConfirmation("Configure legacy single-instance settings? (For backward compatibility only)") {
			return nil
		}
	}

	// API Key
	ui.ShowInfo("Create an API token at: https://sentry.io/settings/account/api/auth-tokens/")
	ui.ShowInfo("Required scopes: project:read, org:read, issue:read")
	fmt.Println()

	apiKey, err := ui.GetInput(
		"🔑 Sentry API Key",
		cfg.Sentry.APIKey,
		true, // password mode
		func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("API key cannot be empty")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	cfg.Sentry.APIKey = apiKey

	// Base URL (for self-hosted Sentry or different regions)
	ui.ShowInfo("Default: https://sentry.io/api/0")
	ui.ShowInfo("For US region: https://sentry.io/api/0")
	ui.ShowInfo("For EU region: https://de.sentry.io/api/0")
	ui.ShowInfo("For self-hosted: https://your-sentry-instance.com/api/0")
	fmt.Println()

	baseURL, err := ui.GetInput(
		"🌐 Sentry Base URL",
		cfg.Sentry.BaseURL,
		false,
		func(s string) error {
			if len(s) > 0 && !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
				return fmt.Errorf("base URL must start with http:// or https://")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	if baseURL == "" {
		cfg.Sentry.BaseURL = "https://sentry.io/api/0"
	} else {
		cfg.Sentry.BaseURL = strings.TrimRight(baseURL, "/")
	}

	// Suggest using Bug Manager for full functionality
	ui.ShowInfo("\n💡 Tip: Use Bug Manager → Manage Instances for multiple Sentry accounts")

	return nil
}

// configureLinear handles Linear configuration
func (m *Module) configureLinear(cfg *config.Config) error {
	fmt.Println()
	ui.ShowInfo("Configure Linear settings for issue tracking")
	ui.ShowWarning("Note: For multiple Linear instances, use the Bug Manager's 'Manage Instances' option")
	fmt.Println()

	// Check if instances exist
	if len(cfg.Linear.Instances) > 0 {
		ui.ShowInfo("You have configured Linear instances in Bug Manager:")
		for key, instance := range cfg.Linear.Instances {
			fmt.Printf("  • %s (%s)\n", instance.Name, key)
		}
		fmt.Println()

		if !ui.GetConfirmation("Configure legacy single-instance settings? (For backward compatibility only)") {
			return nil
		}
	}

	// API Key
	ui.ShowInfo("Create an API key at: https://linear.app/settings/api")
	fmt.Println()

	apiKey, err := ui.GetInput(
		"🔑 Linear API Key",
		cfg.Linear.APIKey,
		true, // password mode
		func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("API key cannot be empty")
			}
			if !strings.HasPrefix(s, "lin_api_") {
				return fmt.Errorf("Linear API key should start with 'lin_api_'")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	cfg.Linear.APIKey = apiKey

	// Suggest using Bug Manager for full functionality
	ui.ShowInfo("\n💡 Tip: Use Bug Manager → Manage Instances for multiple Linear accounts")

	return nil
}

// configureGlobalSettings handles global application settings
func (m *Module) configureGlobalSettings(cfg *config.Config) error {
	fmt.Println()
	ui.ShowInfo("Configure global application settings")
	fmt.Println()

	options := []string{
		"SSH (recommended)",
		"GPG",
	}

	choice, err := ui.SelectFromList("Select preferred signing method:", options)
	if err != nil {
		return err
	}

	switch choice {
	case 0:
		cfg.Settings.PreferredSigningMethod = "ssh"
	case 1:
		cfg.Settings.PreferredSigningMethod = "gpg"
	}

	return nil
}

// showConfigPath displays the configuration file path
func (m *Module) showConfigPath() {
	fmt.Println()
	configPath := config.GetConfigPath()
	box := ui.CreateBox("📁 Configuration File Location", configPath)
	fmt.Println(box)

	ui.ShowInfo("You can manually edit this file if needed")
	ui.ShowInfo("Press Enter to continue...")
	fmt.Scanln()
}
