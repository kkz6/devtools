package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	GitHub     GitHubConfig     `yaml:"github"`
	SSH        SSHConfig        `yaml:"ssh"`
	GPG        GPGConfig        `yaml:"gpg"`
	Cursor     CursorConfig     `yaml:"cursor"`
	Sentry     SentryConfig     `yaml:"sentry"`
	Linear     LinearConfig     `yaml:"linear"`
	Flutter    FlutterConfig    `yaml:"flutter"`
	Settings   GlobalSettings   `yaml:"settings"`
	BugManager BugManagerConfig `yaml:"bug_manager"`
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
	APIKey    string                     `yaml:"api_key"`   // Deprecated: for backward compatibility
	BaseURL   string                     `yaml:"base_url"`  // Deprecated: for backward compatibility
	Projects  map[string]SentryProject   `yaml:"projects"`  // Deprecated: for backward compatibility
	Instances map[string]*SentryInstance `yaml:"instances"` // New: multiple Sentry instances
}

// SentryInstance represents a single Sentry instance configuration
type SentryInstance struct {
	Name    string `yaml:"name"`
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
}

// SentryProject represents a Sentry project configuration
type SentryProject struct {
	OrganizationSlug string `yaml:"organization_slug"`
	ProjectSlug      string `yaml:"project_slug"`
	LinearProjectID  string `yaml:"linear_project_id"` // Mapped Linear project
}

// LinearConfig holds Linear-related configuration
type LinearConfig struct {
	APIKey    string                     `yaml:"api_key"`   // Deprecated: for backward compatibility
	Projects  map[string]LinearProject   `yaml:"projects"`  // Deprecated: for backward compatibility
	Instances map[string]*LinearInstance `yaml:"instances"` // New: multiple Linear instances
}

// LinearInstance represents a single Linear instance configuration
type LinearInstance struct {
	Name   string `yaml:"name"`
	APIKey string `yaml:"api_key"`
}

// LinearProject represents a Linear project configuration
type LinearProject struct {
	TeamID      string   `yaml:"team_id"`
	ProjectID   string   `yaml:"project_id"`
	ProjectName string   `yaml:"project_name"`
	Labels      []string `yaml:"labels"` // Default labels for bugs
}

// BugManagerConfig holds bug manager specific configuration
type BugManagerConfig struct {
	Connections []BugManagerConnection `yaml:"connections"`
}

// BugManagerConnection represents a connection between Linear and Sentry instances
type BugManagerConnection struct {
	Name            string                     `yaml:"name"`
	LinearInstance  string                     `yaml:"linear_instance"`
	SentryInstance  string                     `yaml:"sentry_instance"`
	ProjectMappings []BugManagerProjectMapping `yaml:"project_mappings"`
}

// BugManagerProjectMapping represents a mapping between Sentry and Linear projects
type BugManagerProjectMapping struct {
	SentryOrganization string   `yaml:"sentry_organization"`
	SentryProject      string   `yaml:"sentry_project"`
	LinearTeamID       string   `yaml:"linear_team_id"`
	LinearProjectID    string   `yaml:"linear_project_id"`
	LinearProjectName  string   `yaml:"linear_project_name"`
	DefaultLabels      []string `yaml:"default_labels"`
}

// FlutterConfig holds Flutter-related configuration
type FlutterConfig struct {
	AndroidSDKPath   string                     `yaml:"android_sdk_path"`
	KeystoreDir      string                     `yaml:"keystore_dir"`
	BackupDir        string                     `yaml:"backup_dir"`
	DefaultBuildMode string                     `yaml:"default_build_mode"` // "debug" or "release"
	Projects         map[string]*FlutterProject `yaml:"projects"`
}

// FlutterProject represents a Flutter project configuration
type FlutterProject struct {
	Path         string `yaml:"path"`
	KeystorePath string `yaml:"keystore_path"`
	KeyAlias     string `yaml:"key_alias"`
	LastVersion  string `yaml:"last_version"`
	LastBuildNum string `yaml:"last_build_num"`
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
			BaseURL:   "https://sentry.io/api/0",
			Projects:  make(map[string]SentryProject),
			Instances: make(map[string]*SentryInstance),
		},
		Linear: LinearConfig{
			Projects:  make(map[string]LinearProject),
			Instances: make(map[string]*LinearInstance),
		},
		Flutter: FlutterConfig{
			KeystoreDir:      filepath.Join(homeDir, ".devtools", "flutter", "keystores"),
			BackupDir:        filepath.Join(homeDir, ".devtools", "flutter", "backups"),
			DefaultBuildMode: "release",
			Projects:         make(map[string]*FlutterProject),
		},
		Settings: GlobalSettings{
			PreferredSigningMethod: "ssh",
		},
		BugManager: BugManagerConfig{
			Connections: []BugManagerConnection{},
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
	if cfg.Sentry.Instances == nil {
		cfg.Sentry.Instances = make(map[string]*SentryInstance)
	}
	// Migrate old single API config to instances if needed
	if cfg.Sentry.APIKey != "" && len(cfg.Sentry.Instances) == 0 {
		cfg.Sentry.Instances["default"] = &SentryInstance{
			Name:    "Default Sentry",
			APIKey:  cfg.Sentry.APIKey,
			BaseURL: cfg.Sentry.BaseURL,
		}
	}
	// Set Linear defaults
	if cfg.Linear.Projects == nil {
		cfg.Linear.Projects = make(map[string]LinearProject)
	}
	if cfg.Linear.Instances == nil {
		cfg.Linear.Instances = make(map[string]*LinearInstance)
	}
	// Migrate old single API config to instances if needed
	if cfg.Linear.APIKey != "" && len(cfg.Linear.Instances) == 0 {
		cfg.Linear.Instances["default"] = &LinearInstance{
			Name:   "Default Linear",
			APIKey: cfg.Linear.APIKey,
		}
	}
	// Migrate old project mappings to connections if needed
	if len(cfg.BugManager.Connections) == 0 && len(cfg.Sentry.Projects) > 0 {
		mappings := []BugManagerProjectMapping{}
		for _, sentryProj := range cfg.Sentry.Projects {
			if linearProj, ok := cfg.Linear.Projects[sentryProj.LinearProjectID]; ok {
				mappings = append(mappings, BugManagerProjectMapping{
					SentryOrganization: sentryProj.OrganizationSlug,
					SentryProject:      sentryProj.ProjectSlug,
					LinearTeamID:       linearProj.TeamID,
					LinearProjectID:    linearProj.ProjectID,
					LinearProjectName:  linearProj.ProjectName,
					DefaultLabels:      linearProj.Labels,
				})
			}
		}
		if len(mappings) > 0 {
			cfg.BugManager.Connections = append(cfg.BugManager.Connections, BugManagerConnection{
				Name:            "Default Connection",
				LinearInstance:  "default",
				SentryInstance:  "default",
				ProjectMappings: mappings,
			})
		}
	}
	// Set Flutter defaults
	if cfg.Flutter.KeystoreDir == "" {
		homeDir, _ := os.UserHomeDir()
		cfg.Flutter.KeystoreDir = filepath.Join(homeDir, ".devtools", "flutter", "keystores")
	}
	if cfg.Flutter.BackupDir == "" {
		homeDir, _ := os.UserHomeDir()
		cfg.Flutter.BackupDir = filepath.Join(homeDir, ".devtools", "flutter", "backups")
	}
	if cfg.Flutter.DefaultBuildMode == "" {
		cfg.Flutter.DefaultBuildMode = "release"
	}
	if cfg.Flutter.Projects == nil {
		cfg.Flutter.Projects = make(map[string]*FlutterProject)
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
