package gitsigning

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// checkGitSigningStatus checks if git signing is enabled and what method is being used
func checkGitSigningStatus() (bool, string, error) {
	// Check if commit.gpgsign is true
	cmd := exec.Command("git", "config", "--global", "commit.gpgsign")
	output, err := cmd.Output()
	if err != nil {
		// Config not set, signing is disabled
		return false, "", nil
	}

	isEnabled := strings.TrimSpace(string(output)) == "true"
	if !isEnabled {
		return false, "", nil
	}

	// Check signing format (ssh or openpgp)
	cmd = exec.Command("git", "config", "--global", "gpg.format")
	output, err = cmd.Output()
	if err != nil {
		// Default is openpgp if not set
		return true, "GPG", nil
	}

	format := strings.TrimSpace(string(output))
	if format == "ssh" {
		return true, "SSH", nil
	}
	return true, "GPG", nil
}

// getGitSigningConfig gets the current signing configuration
func getGitSigningConfig() (string, string, error) {
	// Get signing format
	cmd := exec.Command("git", "config", "--global", "gpg.format")
	output, err := cmd.Output()
	format := "openpgp" // default
	if err == nil {
		format = strings.TrimSpace(string(output))
	}

	// Get signing key
	cmd = exec.Command("git", "config", "--global", "user.signingkey")
	output, err = cmd.Output()
	if err != nil {
		return format, "", nil
	}

	key := strings.TrimSpace(string(output))
	return format, key, nil
}

// enableGitSigning enables git commit signing
func enableGitSigning() error {
	cmd := exec.Command("git", "config", "--global", "commit.gpgsign", "true")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable git signing: %s", stderr.String())
	}
	
	// Also enable tag signing
	cmd = exec.Command("git", "config", "--global", "tag.gpgsign", "true")
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// Non-fatal, just log
		fmt.Printf("Note: Could not enable tag signing: %s\n", stderr.String())
	}
	
	return nil
}

// disableGitSigning disables git commit signing
func disableGitSigning() error {
	cmd := exec.Command("git", "config", "--global", "commit.gpgsign", "false")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable git signing: %s", stderr.String())
	}
	
	// Also disable tag signing
	cmd = exec.Command("git", "config", "--global", "tag.gpgsign", "false")
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// Non-fatal, just log
		fmt.Printf("Note: Could not disable tag signing: %s\n", stderr.String())
	}
	
	return nil
}

// showGitSigningStatus displays detailed git signing configuration
func showGitSigningStatus() error {
	fmt.Println("\n📋 Git Signing Configuration:")
	fmt.Println("════════════════════════════")
	
	configs := []struct {
		key   string
		label string
	}{
		{"commit.gpgsign", "Commit signing"},
		{"tag.gpgsign", "Tag signing"},
		{"gpg.format", "Signing format"},
		{"user.signingkey", "Signing key"},
		{"gpg.program", "GPG program"},
	}
	
	for _, config := range configs {
		cmd := exec.Command("git", "config", "--global", config.key)
		output, err := cmd.Output()
		
		value := "not set"
		if err == nil && len(output) > 0 {
			value = strings.TrimSpace(string(output))
			// Truncate long keys for display
			if config.key == "user.signingkey" && len(value) > 50 {
				value = value[:47] + "..."
			}
		}
		
		fmt.Printf("  %-20s: %s\n", config.label, value)
	}
	
	fmt.Println()
	return nil
}

// copyToClipboard copies text to the system clipboard
func copyToClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		// macOS
		cmd = exec.Command("pbcopy")
	case "linux":
		// Linux - try xclip first, then xsel
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard utility found (install xclip or xsel)")
		}
	case "windows":
		// Windows
		cmd = exec.Command("cmd", "/c", "clip")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// Set the text as stdin
	cmd.Stdin = strings.NewReader(text)
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("clipboard error: %s", stderr.String())
		}
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}

// GPGKey represents a GPG key
type GPGKey struct {
	ID    string
	Email string
}

// listGPGKeys lists all GPG keys on the system
func listGPGKeys() ([]GPGKey, error) {
	cmd := exec.Command("gpg", "--list-secret-keys", "--keyid-format=long")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var keys []GPGKey
	lines := strings.Split(string(output), "\n")
	
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "sec") {
			// Extract key ID
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				parts := strings.Split(fields[1], "/")
				if len(parts) == 2 {
					keyID := parts[1]
					
					// Look for email in uid line
					email := ""
					for j := i + 1; j < len(lines) && j < i + 5; j++ {
						if strings.HasPrefix(lines[j], "uid") {
							// Extract email from uid line
							if start := strings.Index(lines[j], "<"); start != -1 {
								if end := strings.Index(lines[j], ">"); end != -1 && end > start {
									email = lines[j][start+1 : end]
								}
							}
							break
						}
					}
					
					keys = append(keys, GPGKey{ID: keyID, Email: email})
				}
			}
		}
	}
	
	return keys, nil
}

// removeGPGKey removes a GPG key from the system
func removeGPGKey(keyID string) error {
	// Delete secret key
	cmd := exec.Command("gpg", "--batch", "--yes", "--delete-secret-keys", keyID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete secret key: %w", err)
	}
	
	// Delete public key
	cmd = exec.Command("gpg", "--batch", "--yes", "--delete-keys", keyID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete public key: %w", err)
	}
	
	return nil
}

// clearGitGPGConfig clears GPG-related git configuration
func clearGitGPGConfig() error {
	configs := []string{
		"user.signingkey",
		"commit.gpgsign",
		"tag.gpgsign",
	}
	
	// Only clear GPG-specific settings if using GPG
	cmd := exec.Command("git", "config", "--global", "gpg.format")
	output, _ := cmd.Output()
	format := strings.TrimSpace(string(output))
	
	if format != "ssh" {
		// Also clear format if it's GPG
		configs = append(configs, "gpg.format")
	}
	
	for _, config := range configs {
		cmd := exec.Command("git", "config", "--global", "--unset", config)
		cmd.Run() // Ignore errors for configs that might not exist
	}
	
	return nil
}

// clearGitSSHConfig clears SSH-related git configuration
func clearGitSSHConfig() error {
	// Check if currently using SSH signing
	cmd := exec.Command("git", "config", "--global", "gpg.format")
	output, _ := cmd.Output()
	format := strings.TrimSpace(string(output))
	
	if format == "ssh" {
		configs := []string{
			"user.signingkey",
			"commit.gpgsign",
			"tag.gpgsign",
			"gpg.format",
		}
		
		for _, config := range configs {
			cmd := exec.Command("git", "config", "--global", "--unset", config)
			cmd.Run() // Ignore errors for configs that might not exist
		}
	}
	
	return nil
} 