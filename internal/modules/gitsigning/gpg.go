package gitsigning

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kkz6/devtools/internal/config"
)

// GPGSigner handles GPG-based commit signing
type GPGSigner struct {
	cfg *config.Config
}

// NewGPGSigner creates a new GPG signer
func NewGPGSigner(cfg *config.Config) *GPGSigner {
	return &GPGSigner{cfg: cfg}
}

// SetupKey sets up the GPG signing key
func (g *GPGSigner) SetupKey() error {
	// Check if GPG is installed
	if err := g.checkGPGInstalled(); err != nil {
		return err
	}

	// Check if key already exists
	if g.cfg.GPG.KeyID != "" {
		keyID, err := g.findExistingKey(g.cfg.GPG.Email)
		if err == nil && keyID != "" {
			fmt.Printf("✅ Found existing GPG key: %s\n", keyID)
			g.cfg.GPG.KeyID = keyID
			return nil
		}
	}

	// Generate new key
	fmt.Printf("🔐 Generating GPG key for email: %s\n", g.cfg.GPG.Email)
	
	// Create batch file for key generation
	batchFile := g.createBatchFile()
	defer os.Remove(batchFile)

	// Generate key
	cmd := exec.Command("gpg", "--batch", "--generate-key", batchFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate GPG key: %w", err)
	}

	// Find the generated key
	keyID, err := g.findExistingKey(g.cfg.GPG.Email)
	if err != nil {
		return fmt.Errorf("failed to find generated key: %w", err)
	}

	g.cfg.GPG.KeyID = keyID
	fmt.Printf("✅ Generated GPG key ID: %s\n", keyID)
	return nil
}

// ConfigureGit configures Git to use GPG signing
func (g *GPGSigner) ConfigureGit() error {
	fmt.Println("🛠  Configuring Git to use GPG signing...")

	// Find GPG program
	gpgPath, err := exec.LookPath("gpg")
	if err != nil {
		return fmt.Errorf("gpg not found in PATH: %w", err)
	}

	commands := [][]string{
		{"config", "--global", "user.email", g.cfg.GPG.Email},
		{"config", "--global", "user.signingkey", g.cfg.GPG.KeyID},
		{"config", "--global", "commit.gpgsign", "true"},
		{"config", "--global", "gpg.program", gpgPath},
	}

	for _, args := range commands {
		cmd := exec.Command("git", args...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run git %v: %w", args, err)
		}
	}

	fmt.Println("✅ Git configured for GPG signing")
	return nil
}

// ExportPublicKey exports the GPG public key
func (g *GPGSigner) ExportPublicKey() (string, error) {
	homeDir, _ := os.UserHomeDir()
	outputFile := filepath.Join(homeDir, fmt.Sprintf("%s_gpgkey.asc", g.cfg.GPG.Email))

	// Export to file
	cmd := exec.Command("gpg", "--armor", "--export", g.cfg.GPG.KeyID)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to export GPG key: %w", err)
	}

	// Save to file
	if err := os.WriteFile(outputFile, output, 0644); err != nil {
		return "", fmt.Errorf("failed to write GPG key file: %w", err)
	}

	fmt.Printf("📝 GPG public key exported to: %s\n", outputFile)
	return string(output), nil
}

// checkGPGInstalled checks if GPG is installed and installs it if needed
func (g *GPGSigner) checkGPGInstalled() error {
	_, err := exec.LookPath("gpg")
	if err == nil {
		return nil
	}

	fmt.Println("🔧 GPG not found. Installing GPG tools...")

	switch runtime.GOOS {
	case "darwin":
		// Check if Homebrew is installed
		if _, err := exec.LookPath("brew"); err != nil {
			return fmt.Errorf("Homebrew is required to install GPG. Please install Homebrew first")
		}

		cmd := exec.Command("brew", "install", "gpg2", "gnupg", "pinentry-mac")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install GPG: %w", err)
		}
	case "linux":
		fmt.Println("Please install GPG using your package manager:")
		fmt.Println("  Ubuntu/Debian: sudo apt-get install gnupg")
		fmt.Println("  Fedora/RHEL: sudo dnf install gnupg")
		fmt.Println("  Arch: sudo pacman -S gnupg")
		return fmt.Errorf("GPG not installed")
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return nil
}

// findExistingKey finds an existing GPG key for the given email
func (g *GPGSigner) findExistingKey(email string) (string, error) {
	cmd := exec.Command("gpg", "--list-secret-keys", "--keyid-format=long", email)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse output to find key ID
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "sec") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				// Format: sec   rsa4096/KEYID
				parts := strings.Split(fields[1], "/")
				if len(parts) == 2 {
					return parts[1], nil
				}
			}
		}
	}

	return "", fmt.Errorf("no key found for email: %s", email)
}

// createBatchFile creates a batch file for GPG key generation
func (g *GPGSigner) createBatchFile() string {
	content := fmt.Sprintf(`%%echo Generating GPG key
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: %s
Name-Email: %s
Expire-Date: 0
%%no-protection
%%commit
%%echo done`, g.cfg.GPG.Email, g.cfg.GPG.Email)

	tmpFile, _ := os.CreateTemp("", "gpg-batch-*.txt")
	tmpFile.WriteString(content)
	tmpFile.Close()

	return tmpFile.Name()
} 