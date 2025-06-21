package gitsigning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/ui"
	"golang.org/x/crypto/ssh"
)

// SSHSigner handles SSH-based commit signing
type SSHSigner struct {
	cfg *config.Config
}

// NewSSHSigner creates a new SSH signer
func NewSSHSigner(cfg *config.Config) *SSHSigner {
	return &SSHSigner{cfg: cfg}
}

// SetupKey sets up the SSH signing key
func (s *SSHSigner) SetupKey(force bool, importMode bool) error {
	keyPath := s.cfg.SSH.SigningKeyPath
	pubKeyPath := keyPath + ".pub"

	// Check if key exists
	if _, err := os.Stat(keyPath); err == nil && !force {
		if importMode {
			ui.ShowInfo(fmt.Sprintf("Using imported SSH signing key at %s", keyPath))
		} else {
			ui.ShowInfo(fmt.Sprintf("SSH signing key already exists at %s", keyPath))
		}
		return nil
	}

	if importMode {
		return fmt.Errorf("SSH key files not found at %s and %s", keyPath, pubKeyPath)
	}

	// Ensure .ssh directory exists
	sshDir := filepath.Dir(keyPath)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create SSH directory: %w", err)
	}

	// Generate new key
	ui.ShowInfo("Generating new SSH key...")
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-C", s.cfg.SSH.KeyComment, "-f", keyPath, "-N", "")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate SSH key: %w", err)
	}

	// Add to SSH agent
	ui.ShowInfo("Adding SSH key to agent...")
	
	// Start ssh-agent if needed
	agentCmd := exec.Command("ssh-agent", "-s")
	output, err := agentCmd.Output()
	if err == nil {
		// Parse and set SSH_AUTH_SOCK
		for _, line := range bytes.Split(output, []byte("\n")) {
			if bytes.Contains(line, []byte("SSH_AUTH_SOCK=")) {
				parts := bytes.Split(line, []byte("="))
				if len(parts) >= 2 {
					sock := bytes.TrimSuffix(parts[1], []byte(";"))
					os.Setenv("SSH_AUTH_SOCK", string(sock))
				}
			}
		}
	}

	// Add key to agent
	addCmd := exec.Command("ssh-add", keyPath)
	if err := addCmd.Run(); err != nil {
		// Non-fatal error
		ui.ShowWarning(fmt.Sprintf("Could not add key to SSH agent: %v", err))
	}

	ui.ShowSuccess(fmt.Sprintf("SSH key generated at %s", keyPath))
	return nil
}

// ConfigureGit configures Git to use SSH signing
func (s *SSHSigner) ConfigureGit() error {
	// Read public key
	pubKeyPath := s.cfg.SSH.SigningKeyPath + ".pub"
	pubKeyData, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	commands := [][]string{
		{"config", "--global", "gpg.format", "ssh"},
		{"config", "--global", "user.signingkey", string(pubKeyData)},
		{"config", "--global", "commit.gpgsign", "true"},
	}

	for _, args := range commands {
		cmd := exec.Command("git", args...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run git %v: %w", args, err)
		}
	}

	return nil
}

// GetPublicKey returns the SSH public key
func (s *SSHSigner) GetPublicKey() (string, error) {
	pubKeyPath := s.cfg.SSH.SigningKeyPath + ".pub"
	pubKeyData, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read public key: %w", err)
	}
	return string(pubKeyData), nil
}

// UploadToGitHub uploads the SSH signing key to GitHub
func (s *SSHSigner) UploadToGitHub() error {
	if s.cfg.GitHub.Token == "" || s.cfg.GitHub.Username == "" {
		return fmt.Errorf("GitHub credentials not configured")
	}

	// Read public key
	pubKeyPath := s.cfg.SSH.SigningKeyPath + ".pub"
	pubKeyData, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	// Parse to validate
	_, _, _, _, err = ssh.ParseAuthorizedKey(pubKeyData)
	if err != nil {
		return fmt.Errorf("invalid SSH public key: %w", err)
	}

	// Prepare request
	hostname, _ := os.Hostname()
	title := fmt.Sprintf("Git SSH Signing Key from %s on %s", hostname, time.Now().Format("2006-01-02_15-04-05"))

	payload := map[string]string{
		"key":   string(pubKeyData),
		"title": title,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Make request
	req, err := http.NewRequest("POST", "https://api.github.com/user/ssh_signing_keys", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+s.cfg.GitHub.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		return nil
	}

	// Read error response
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("failed to upload key (HTTP %d): %s", resp.StatusCode, string(body))
} 