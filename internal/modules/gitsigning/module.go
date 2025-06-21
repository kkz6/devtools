package gitsigning

import (
	"fmt"
	"strings"
	"time"

	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/types"
	"github.com/kkz6/devtools/internal/ui"
)

// Module implements the git signing configuration module
type Module struct{}

// New creates a new git signing module
func New() *Module {
	return &Module{}
}

// Info returns module information
func (m *Module) Info() types.ModuleInfo {
	return types.ModuleInfo{
		ID:          "git-signing",
		Name:        "Git Commit Signing Setup",
		Description: "Configure SSH or GPG signing for Git commits",
	}
}

// Execute runs the git signing setup
func (m *Module) Execute(cfg *config.Config) error {
	ui.ShowBanner()
	
	title := ui.GetGradientTitle("🔐 Git Commit Signing Setup")
	fmt.Println(title)
	fmt.Println()

	options := []string{
		"SSH signing (recommended)",
		"GPG signing",
		"Export existing SSH key to GitHub",
		"Import existing SSH signing key",
		"Back to main menu",
	}
	
	choice, err := ui.SelectFromList("Select signing method:", options)
	if err != nil {
		return nil
	}

	switch choice {
	case 0:
		return m.setupSSHSigning(cfg, false, false)
	case 1:
		return m.setupGPGSigning(cfg)
	case 2:
		return m.exportSSHKeyToGitHub(cfg)
	case 3:
		return m.setupSSHSigning(cfg, false, true)
	case 4:
		return nil
	default:
		ui.ShowError("Invalid choice")
		return m.Execute(cfg)
	}
}

// setupSSHSigning configures SSH-based commit signing
func (m *Module) setupSSHSigning(cfg *config.Config, force bool, importMode bool) error {
	sshSigner := NewSSHSigner(cfg)
	
	// Check if we need to force regenerate
	if !force && !importMode {
		if ui.GetConfirmation("⚠️  Force regenerate SSH key?") {
			force = true
		}
	}

	// Setup the SSH key with animation
	err := ui.ShowLoadingAnimation("Setting up SSH key", func() error {
		return sshSigner.SetupKey(force, importMode)
	})
	if err != nil {
		return fmt.Errorf("failed to setup SSH key: %w", err)
	}

	// Configure Git with animation
	err = ui.ShowLoadingAnimation("Configuring Git", func() error {
		return sshSigner.ConfigureGit()
	})
	if err != nil {
		return fmt.Errorf("failed to configure Git: %w", err)
	}

	// Ask if user wants to upload to GitHub
	if ui.GetConfirmation("📤 Upload SSH key to GitHub?") {
		if err := m.promptGitHubCredentials(cfg); err != nil {
			return err
		}
		
		err = ui.ShowLoadingAnimation("Uploading to GitHub", func() error {
			return sshSigner.UploadToGitHub()
		})
		if err != nil {
			ui.ShowWarning(fmt.Sprintf("Failed to upload to GitHub: %v", err))
		}
	}

	time.Sleep(500 * time.Millisecond)
	ui.ShowSuccess("SSH signing setup complete!")
	return nil
}

// setupGPGSigning configures GPG-based commit signing
func (m *Module) setupGPGSigning(cfg *config.Config) error {
	gpgSigner := NewGPGSigner(cfg)

	// Get email if not configured
	if cfg.GPG.Email == "" {
		email, err := ui.GetInput(
			"📧 Enter email for GPG key",
			"your-email@example.com",
			false,
			func(s string) error {
				if !strings.Contains(s, "@") {
					return fmt.Errorf("please enter a valid email address")
				}
				return nil
			},
		)
		if err != nil {
			return err
		}
		cfg.GPG.Email = strings.TrimSpace(email)
	}

	// Setup GPG with animation
	err := ui.ShowLoadingAnimation("Setting up GPG key", func() error {
		return gpgSigner.SetupKey()
	})
	if err != nil {
		return fmt.Errorf("failed to setup GPG key: %w", err)
	}

	// Configure Git with animation
	err = ui.ShowLoadingAnimation("Configuring Git", func() error {
		return gpgSigner.ConfigureGit()
	})
	if err != nil {
		return fmt.Errorf("failed to configure Git: %w", err)
	}

	// Export public key
	var publicKey string
	err = ui.ShowLoadingAnimation("Exporting public key", func() error {
		var err error
		publicKey, err = gpgSigner.ExportPublicKey()
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to export public key: %w", err)
	}

	// Display public key in a nice box
	keyBox := ui.CreateBox("🔐 GPG PUBLIC KEY", publicKey)
	fmt.Println(keyBox)

	ui.ShowInfo("📋 Please copy the above key and add it to GitHub:")
	ui.ShowInfo("👉 https://github.com/settings/keys")

	if ui.GetConfirmation("🌐 Open GitHub in browser?") {
		ui.ShowInfo("Please open https://github.com/settings/keys in your browser")
	}

	time.Sleep(500 * time.Millisecond)
	ui.ShowSuccess("GPG signing setup complete!")
	return nil
}

// exportSSHKeyToGitHub exports an existing SSH key to GitHub
func (m *Module) exportSSHKeyToGitHub(cfg *config.Config) error {
	if err := m.promptGitHubCredentials(cfg); err != nil {
		return err
	}

	sshSigner := NewSSHSigner(cfg)
	if err := sshSigner.UploadToGitHub(); err != nil {
		return fmt.Errorf("failed to upload to GitHub: %w", err)
	}

	fmt.Println("\n✅ SSH key uploaded to GitHub!")
	return nil
}

// promptGitHubCredentials prompts for GitHub credentials if not configured
func (m *Module) promptGitHubCredentials(cfg *config.Config) error {
	if cfg.GitHub.Username == "" {
		username, err := ui.GetInput(
			"👤 GitHub username",
			"your-github-username",
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
		cfg.GitHub.Username = strings.TrimSpace(username)
	}

	if cfg.GitHub.Token == "" {
		token, err := ui.GetInput(
			"🔑 GitHub personal access token",
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
		cfg.GitHub.Token = strings.TrimSpace(token)
	}

	if cfg.GitHub.Username == "" || cfg.GitHub.Token == "" {
		return fmt.Errorf("GitHub credentials are required")
	}

	return nil
} 