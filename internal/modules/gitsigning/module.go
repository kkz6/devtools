package gitsigning

import (
	"fmt"
	"os"
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
		"View Current Signing Status",
		"Enable/Disable Git Signing",
		"SSH signing setup (recommended)",
		"GPG signing setup",
		"Export existing SSH key to GitHub",
		"Import existing SSH signing key",
		"Clean up GPG/SSH signing (Remove all)",
		"Back to main menu",
	}
	
	choice, err := ui.SelectFromList("Select signing method:", options)
	if err != nil {
		return nil
	}

	switch choice {
	case 0:
		if err := showGitSigningStatus(); err != nil {
			ui.ShowError(fmt.Sprintf("Failed to show status: %v", err))
		}
		fmt.Print("\nPress Enter to continue...")
		fmt.Scanln()
		return m.Execute(cfg)
	case 1:
		return m.toggleGitSigning(cfg)
	case 2:
		return m.setupSSHSigning(cfg, false, false)
	case 3:
		return m.setupGPGSigning(cfg)
	case 4:
		return m.exportSSHKeyToGitHub(cfg)
	case 5:
		return m.setupSSHSigning(cfg, false, true)
	case 6:
		return m.cleanupSigning(cfg)
	case 7:
		return nil
	default:
		ui.ShowError("Invalid choice")
		return m.Execute(cfg)
	}
}

// toggleGitSigning enables or disables git commit signing
func (m *Module) toggleGitSigning(cfg *config.Config) error {
	// Check current signing status
	isEnabled, signingMethod, err := checkGitSigningStatus()
	if err != nil {
		ui.ShowWarning("Could not determine current signing status")
	}

	fmt.Println()
	if isEnabled {
		ui.ShowInfo(fmt.Sprintf("Git signing is currently ENABLED using %s", signingMethod))
		
		if ui.GetConfirmation("Do you want to DISABLE git signing?") {
			err := ui.ShowLoadingAnimation("Disabling git signing", func() error {
				return disableGitSigning()
			})
			if err != nil {
				return fmt.Errorf("failed to disable git signing: %w", err)
			}
			ui.ShowSuccess("Git signing has been disabled")
		}
	} else {
		ui.ShowInfo("Git signing is currently DISABLED")
		
		if ui.GetConfirmation("Do you want to ENABLE git signing?") {
			// Check if signing is already configured
			format, key, err := getGitSigningConfig()
			if err != nil || key == "" {
				ui.ShowWarning("No signing key configured. Please set up SSH or GPG signing first.")
				return nil
			}
			
			err = ui.ShowLoadingAnimation("Enabling git signing", func() error {
				return enableGitSigning()
			})
			if err != nil {
				return fmt.Errorf("failed to enable git signing: %w", err)
			}
			ui.ShowSuccess(fmt.Sprintf("Git signing has been enabled using %s", format))
		}
	}
	
	return nil
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

	// Offer to copy SSH public key to clipboard
	if ui.GetConfirmation("📋 Copy SSH public key to clipboard?") {
		pubKey, err := sshSigner.GetPublicKey()
		if err == nil {
			err = ui.ShowLoadingAnimation("Copying to clipboard", func() error {
				return copyToClipboard(pubKey)
			})
			if err != nil {
				ui.ShowWarning(fmt.Sprintf("Could not copy to clipboard: %v", err))
			} else {
				ui.ShowSuccess("✅ SSH public key copied to clipboard!")
				ui.ShowInfo("You can now paste it wherever needed")
			}
		}
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

	ui.ShowInfo("📋 Press Enter to copy the key to clipboard")
	ui.ShowInfo("👉 Then add it to: https://github.com/settings/keys")
	
	fmt.Print("\nPress Enter to copy the key to clipboard...")
	fmt.Scanln()
	
	// Copy to clipboard
	err = ui.ShowLoadingAnimation("Copying to clipboard", func() error {
		return copyToClipboard(publicKey)
	})
	
	if err != nil {
		ui.ShowWarning(fmt.Sprintf("Could not copy to clipboard: %v", err))
		ui.ShowInfo("Please copy the key manually from above")
	} else {
		ui.ShowSuccess("✅ GPG public key copied to clipboard!")
	}

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

// cleanupSigning removes GPG/SSH signing configuration
func (m *Module) cleanupSigning(cfg *config.Config) error {
	ui.ShowWarning("⚠️  This will remove signing configuration and keys")
	fmt.Println()
	
	options := []string{
		"Remove GPG keys only",
		"Remove SSH signing keys only", 
		"Remove all signing configuration",
		"Cancel",
	}
	
	choice, err := ui.SelectFromList("What would you like to clean up?", options)
	if err != nil || choice == 3 {
		return nil
	}
	
	switch choice {
	case 0:
		return m.cleanupGPG(cfg)
	case 1:
		return m.cleanupSSH(cfg)
	case 2:
		// Clean up both
		if err := m.cleanupGPG(cfg); err != nil {
			ui.ShowError(fmt.Sprintf("Failed to cleanup GPG: %v", err))
		}
		if err := m.cleanupSSH(cfg); err != nil {
			ui.ShowError(fmt.Sprintf("Failed to cleanup SSH: %v", err))
		}
		return nil
	}
	
	return nil
}

// cleanupGPG removes GPG keys and configuration
func (m *Module) cleanupGPG(cfg *config.Config) error {
	ui.ShowWarning("This will remove GPG keys from your system and GitHub")
	if !ui.GetConfirmation("Are you sure you want to continue?") {
		return nil
	}
	
	// List GPG keys
	gpgKeys, err := listGPGKeys()
	if err != nil {
		ui.ShowWarning(fmt.Sprintf("Could not list GPG keys: %v", err))
	} else if len(gpgKeys) > 0 {
		fmt.Println("\n📋 Found GPG keys:")
		for _, key := range gpgKeys {
			fmt.Printf("  • %s - %s\n", key.ID, key.Email)
		}
		
		if ui.GetConfirmation("Remove these GPG keys from your system?") {
			err := ui.ShowLoadingAnimation("Removing GPG keys", func() error {
				for _, key := range gpgKeys {
					if err := removeGPGKey(key.ID); err != nil {
						return fmt.Errorf("failed to remove key %s: %w", key.ID, err)
					}
				}
				return nil
			})
			
			if err != nil {
				ui.ShowError(fmt.Sprintf("Failed to remove GPG keys: %v", err))
			} else {
				ui.ShowSuccess("GPG keys removed from system")
			}
		}
	} else {
		ui.ShowInfo("No GPG keys found on system")
	}
	
	// Remove from GitHub if configured
	if cfg.GitHub.Token != "" && ui.GetConfirmation("Remove GPG keys from GitHub?") {
		err := ui.ShowLoadingAnimation("Removing from GitHub", func() error {
			return removeGPGKeysFromGitHub(cfg)
		})
		
		if err != nil {
			ui.ShowWarning(fmt.Sprintf("Could not remove from GitHub: %v", err))
		} else {
			ui.ShowSuccess("GPG keys removed from GitHub")
		}
	}
	
	// Clear git configuration
	err = ui.ShowLoadingAnimation("Clearing git configuration", func() error {
		return clearGitGPGConfig()
	})
	
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to clear git config: %v", err))
	} else {
		ui.ShowSuccess("Git GPG configuration cleared")
	}
	
	// Clear config file
	cfg.GPG.KeyID = ""
	cfg.GPG.Email = ""
	
	return nil
}

// cleanupSSH removes SSH signing keys and configuration
func (m *Module) cleanupSSH(cfg *config.Config) error {
	ui.ShowWarning("This will remove SSH signing configuration")
	if !ui.GetConfirmation("Are you sure you want to continue?") {
		return nil
	}
	
	// Check if SSH signing key exists
	keyPath := cfg.SSH.SigningKeyPath
	keyExists := false
	if _, err := os.Stat(keyPath); err == nil {
		keyExists = true
		fmt.Printf("\n📋 Found SSH signing key: %s\n", keyPath)
		
		if ui.GetConfirmation("Remove this SSH signing key from your system?") {
			err := ui.ShowLoadingAnimation("Removing SSH key", func() error {
				// Remove private key
				if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove private key: %w", err)
				}
				// Remove public key
				if err := os.Remove(keyPath + ".pub"); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove public key: %w", err)
				}
				return nil
			})
			
			if err != nil {
				ui.ShowError(fmt.Sprintf("Failed to remove SSH key: %v", err))
			} else {
				ui.ShowSuccess("SSH signing key removed from system")
			}
		}
	} else {
		ui.ShowInfo("No SSH signing key found at configured path")
	}
	
	// Remove from GitHub if configured
	if cfg.GitHub.Token != "" && keyExists && ui.GetConfirmation("Remove SSH signing keys from GitHub?") {
		err := ui.ShowLoadingAnimation("Removing from GitHub", func() error {
			return removeSSHKeysFromGitHub(cfg)
		})
		
		if err != nil {
			ui.ShowWarning(fmt.Sprintf("Could not remove from GitHub: %v", err))
		} else {
			ui.ShowSuccess("SSH signing keys removed from GitHub")
		}
	}
	
	// Clear git configuration
	err := ui.ShowLoadingAnimation("Clearing git configuration", func() error {
		return clearGitSSHConfig()
	})
	
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to clear git config: %v", err))
	} else {
		ui.ShowSuccess("Git SSH configuration cleared")
	}
	
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