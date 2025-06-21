package gitsigning

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kkz6/devtools/internal/config"
)

// GitHubGPGKey represents a GPG key on GitHub
type GitHubGPGKey struct {
	ID        int    `json:"id"`
	KeyID     string `json:"key_id"`
	PublicKey string `json:"public_key"`
	Emails    []struct {
		Email    string `json:"email"`
		Verified bool   `json:"verified"`
	} `json:"emails"`
}

// GitHubSSHKey represents an SSH signing key on GitHub
type GitHubSSHKey struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Title string `json:"title"`
}

// removeGPGKeysFromGitHub removes all GPG keys from GitHub
func removeGPGKeysFromGitHub(cfg *config.Config) error {
	// List GPG keys
	keys, err := listGitHubGPGKeys(cfg)
	if err != nil {
		return fmt.Errorf("failed to list GPG keys: %w", err)
	}

	if len(keys) == 0 {
		return nil // No keys to remove
	}

	// Remove each key
	for _, key := range keys {
		if err := deleteGitHubGPGKey(cfg, key.ID); err != nil {
			return fmt.Errorf("failed to delete GPG key %d: %w", key.ID, err)
		}
	}

	return nil
}

// removeSSHKeysFromGitHub removes all SSH signing keys from GitHub
func removeSSHKeysFromGitHub(cfg *config.Config) error {
	// List SSH signing keys
	keys, err := listGitHubSSHSigningKeys(cfg)
	if err != nil {
		return fmt.Errorf("failed to list SSH signing keys: %w", err)
	}

	if len(keys) == 0 {
		return nil // No keys to remove
	}

	// Remove each key
	for _, key := range keys {
		if err := deleteGitHubSSHSigningKey(cfg, key.ID); err != nil {
			return fmt.Errorf("failed to delete SSH signing key %d: %w", key.ID, err)
		}
	}

	return nil
}

// listGitHubGPGKeys lists all GPG keys on GitHub
func listGitHubGPGKeys(cfg *config.Config) ([]GitHubGPGKey, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/gpg_keys", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+cfg.GitHub.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var keys []GitHubGPGKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return keys, nil
}

// deleteGitHubGPGKey deletes a GPG key from GitHub
func deleteGitHubGPGKey(cfg *config.Config, keyID int) error {
	url := fmt.Sprintf("https://api.github.com/user/gpg_keys/%d", keyID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+cfg.GitHub.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// listGitHubSSHSigningKeys lists all SSH signing keys on GitHub
func listGitHubSSHSigningKeys(cfg *config.Config) ([]GitHubSSHKey, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/ssh_signing_keys", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+cfg.GitHub.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var keys []GitHubSSHKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return keys, nil
}

// deleteGitHubSSHSigningKey deletes an SSH signing key from GitHub
func deleteGitHubSSHSigningKey(cfg *config.Config, keyID int) error {
	url := fmt.Sprintf("https://api.github.com/user/ssh_signing_keys/%d", keyID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+cfg.GitHub.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
} 