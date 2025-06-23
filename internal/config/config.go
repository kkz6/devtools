package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	GitHub   GitHubConfig   `yaml:"github"`
	SSH      SSHConfig      `yaml:"ssh"`
	GPG      GPGConfig      `yaml:"gpg"`
	Cursor   CursorConfig   `yaml:"cursor"`
	Sentry   SentryConfig   `yaml:"sentry"`
	Linear   LinearConfig   `yaml:"linear"`
	Settings GlobalSettings `yaml:"settings"`
}

// GitHubConfig holds GitHub-related configuration
type GitHubConfig struct {
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
	Email    string `yaml:"email"`
}

// SSHConfig holds SSH-related configuration
type SSHConfig struct {
	SigningKeyPath string `yaml:"signing_key_path"`
	KeyComment     string `yaml:"key_comment"`
}

// GPGConfig holds GPG-related configuration
type GPGConfig struct {
	KeyID string `yaml:"key_id"`
	Email string `yaml:"email"`
}

// CursorConfig holds Cursor AI-related configuration
type CursorConfig struct {
	APIKey      string `yaml:"api_key"`
	APIEndpoint string `yaml:"api_endpoint"`
	CurrentPlan string `yaml:"current_plan"` // "free", "pro", "business"
}

// GlobalSettings holds global application settings
type GlobalSettings struct {
	PreferredSigningMethod string `yaml:"preferred_signing_method"` // "ssh" or "gpg"
}

// SentryConfig holds Sentry-related configuration
type SentryConfig struct {
	APIKey   string                   `yaml:"api_key"`
	BaseURL  string                   `yaml:"base_url"`
	Projects map[string]SentryProject `yaml:"projects"`
}

// SentryProject represents a Sentry project configuration
type SentryProject struct {
	OrganizationSlug string `yaml:"organization_slug"`
	ProjectSlug      string `yaml:"project_slug"`
	LinearProjectID  string `yaml:"linear_project_id"` // Mapped Linear project
}

// LinearConfig holds Linear-related configuration
type LinearConfig struct {
	APIKey   string                   `yaml:"api_key"`
	Projects map[string]LinearProject `yaml:"projects"`
}

// LinearProject represents a Linear project configuration
type LinearProject struct {
	TeamID      string   `yaml:"team_id"`
	ProjectID   string   `yaml:"project_id"`
	ProjectName string   `yaml:"project_name"`
	Labels      []string `yaml:"labels"` // Default labels for bugs
}

// New creates a new default configuration
func New() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		SSH: SSHConfig{
			SigningKeyPath: filepath.Join(homeDir, ".ssh", "git-ssh-signing-key"),
			KeyComment:     "git-ssh-signing-key",
		},
		Sentry: SentryConfig{
			BaseURL:  "https://sentry.io/api/0",
			Projects: make(map[string]SentryProject),
		},
		Linear: LinearConfig{
			Projects: make(map[string]LinearProject),
		},
		Settings: GlobalSettings{
			PreferredSigningMethod: "ssh",
		},
	}
}

// Load loads configuration from file
func Load() (*Config, error) {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config if it doesn't exist
			cfg := New()
			if err := Save(cfg); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults for missing values
	if cfg.SSH.SigningKeyPath == "" {
		homeDir, _ := os.UserHomeDir()
		cfg.SSH.SigningKeyPath = filepath.Join(homeDir, ".ssh", "git-ssh-signing-key")
	}
	if cfg.SSH.KeyComment == "" {
		cfg.SSH.KeyComment = "git-ssh-signing-key"
	}
	if cfg.Settings.PreferredSigningMethod == "" {
		cfg.Settings.PreferredSigningMethod = "ssh"
	}
	// Set Sentry defaults
	if cfg.Sentry.BaseURL == "" {
		cfg.Sentry.BaseURL = "https://sentry.io/api/0"
	}
	if cfg.Sentry.Projects == nil {
		cfg.Sentry.Projects = make(map[string]SentryProject)
	}
	// Set Linear defaults
	if cfg.Linear.Projects == nil {
		cfg.Linear.Projects = make(map[string]LinearProject)
	}

	return &cfg, nil
}

// Save saves configuration to file
func Save(cfg *Config) error {
	configPath := getConfigPath()

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// getConfigPath returns the path to the configuration file
func getConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".devtools", "config.yaml")
}

// GetConfigPath returns the path to the configuration file (exported)
func GetConfigPath() string {
	return getConfigPath()
}
