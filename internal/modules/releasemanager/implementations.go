package releasemanager

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kkz6/devtools/internal/ui"
)

// showProjectStatus displays current project status
func (m *Module) showProjectStatus() error {
	fmt.Println()
	ui.ShowInfo("📊 DevTools Project Status")
	fmt.Println()

	// Show current version
	if version, err := m.getCurrentVersion(); err == nil {
		ui.ShowInfo(fmt.Sprintf("Current Version: %s", version))
	} else {
		ui.ShowInfo("Current Version: No tags found")
	}

	// Show git status
	fmt.Println()
	ui.ShowInfo("📝 Git Status:")
	if err := m.runCommand("git", "status", "--short"); err != nil {
		ui.ShowError("Failed to get git status")
	}

	// Show recent commits
	fmt.Println()
	ui.ShowInfo("📅 Recent Commits:")
	cmd := exec.Command("git", "log", "--oneline", "-5")
	if output, err := cmd.Output(); err == nil {
		fmt.Println(string(output))
	}

	fmt.Println()
	ui.ShowInfo("Press Enter to continue...")
	fmt.Scanln()
	return nil
}

// listTags lists all git tags
func (m *Module) listTags() error {
	fmt.Println()
	ui.ShowInfo("📋 All Git Tags:")
	fmt.Println()

	cmd := exec.Command("git", "tag", "-l")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list tags: %v", err)
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(tags) == 1 && tags[0] == "" {
		ui.ShowInfo("No tags found")
		return nil
	}

	// Sort tags using semantic versioning
	sort.Slice(tags, func(i, j int) bool {
		return m.compareVersions(tags[i], tags[j]) < 0
	})

	for i, tag := range tags {
		fmt.Printf("  %d. %s\n", i+1, tag)
	}

	fmt.Println()
	ui.ShowInfo("Press Enter to continue...")
	fmt.Scanln()
	return nil
}

// deleteTag deletes a git tag
func (m *Module) deleteTag() error {
	fmt.Println()

	// Get list of tags
	cmd := exec.Command("git", "tag", "-l")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list tags: %v", err)
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(tags) == 1 && tags[0] == "" {
		ui.ShowInfo("No tags found to delete")
		return nil
	}

	// Sort tags
	sort.Slice(tags, func(i, j int) bool {
		return m.compareVersions(tags[i], tags[j]) < 0
	})

	ui.ShowInfo("Select tag to delete:")
	fmt.Println()
	for i, tag := range tags {
		fmt.Printf("  %d. %s\n", i+1, tag)
	}

	fmt.Print("\nEnter tag number (or 0 to cancel): ")
	var input string
	fmt.Scanln(&input)

	if input == "0" || input == "" {
		ui.ShowInfo("Operation cancelled")
		return nil
	}

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(tags) {
		return fmt.Errorf("invalid selection")
	}

	tagToDelete := tags[choice-1]

	// Confirm deletion
	confirm, err := ui.GetInput(
		fmt.Sprintf("Delete tag '%s' locally and remotely? (y/N)", tagToDelete),
		"N",
		false,
		nil,
	)
	if err != nil {
		return err
	}

	if strings.ToLower(confirm) != "y" {
		ui.ShowInfo("Deletion cancelled")
		return nil
	}

	return ui.ShowLoadingAnimation("Deleting tag", func() error {
		// Delete local tag
		if err := m.runCommand("git", "tag", "-d", tagToDelete); err != nil {
			return fmt.Errorf("failed to delete local tag: %v", err)
		}

		// Delete remote tag
		if err := m.runCommandSilent("git", "push", "--delete", "origin", tagToDelete); err != nil {
			ui.ShowInfo("Note: Remote tag not found or already deleted")
		}

		ui.ShowSuccess(fmt.Sprintf("✅ Tag '%s' deleted", tagToDelete))
		return nil
	})
}

// buildLocal builds the project locally
func (m *Module) buildLocal() error {
	fmt.Println()
	ui.ShowInfo("🔨 Building DevTools locally...")
	fmt.Println()

	return ui.ShowLoadingAnimation("Building", func() error {
		if err := m.runCommand("go", "build", "-v", "-o", "devtools", "."); err != nil {
			return fmt.Errorf("build failed: %v", err)
		}

		ui.ShowSuccess("✅ Build complete: ./devtools")
		return nil
	})
}

// installLocal installs the built binary locally
func (m *Module) installLocal() error {
	fmt.Println()

	// First build if binary doesn't exist
	if _, err := os.Stat("devtools"); os.IsNotExist(err) {
		ui.ShowInfo("Binary not found, building first...")
		if err := m.buildLocal(); err != nil {
			return err
		}
	}

	ui.ShowInfo("📦 Installing DevTools to /usr/local/bin...")
	fmt.Println()

	return ui.ShowLoadingAnimation("Installing", func() error {
		if err := m.runCommand("sudo", "cp", "devtools", "/usr/local/bin/"); err != nil {
			return fmt.Errorf("installation failed: %v", err)
		}

		ui.ShowSuccess("✅ Installed successfully to /usr/local/bin/devtools")
		return nil
	})
}

// runTests runs the test suite
func (m *Module) runTests() error {
	fmt.Println()
	ui.ShowInfo("🧪 Running tests...")
	fmt.Println()

	return ui.ShowLoadingAnimation("Testing", func() error {
		if err := m.runCommand("go", "test", "-v", "./..."); err != nil {
			return fmt.Errorf("tests failed: %v", err)
		}

		ui.ShowSuccess("✅ All tests passed")
		return nil
	})
}

// runLinter runs the linter
func (m *Module) runLinter() error {
	fmt.Println()
	ui.ShowInfo("🔍 Running linter...")
	fmt.Println()

	// Check if golangci-lint is available
	if err := m.runCommandSilent("which", "golangci-lint"); err != nil {
		ui.ShowError("golangci-lint not found")
		ui.ShowInfo("Install with: brew install golangci-lint")
		return nil
	}

	return ui.ShowLoadingAnimation("Linting", func() error {
		if err := m.runCommand("golangci-lint", "run"); err != nil {
			return fmt.Errorf("linter found issues: %v", err)
		}

		ui.ShowSuccess("✅ Linter passed")
		return nil
	})
}

// cleanArtifacts cleans build artifacts
func (m *Module) cleanArtifacts() error {
	fmt.Println()
	ui.ShowInfo("🧹 Cleaning build artifacts...")
	fmt.Println()

	return ui.ShowLoadingAnimation("Cleaning", func() error {
		// Remove binary
		if err := os.Remove("devtools"); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove binary: %v", err)
		}

		// Remove dist directory if it exists
		if err := os.RemoveAll("dist/"); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove dist directory: %v", err)
		}

		ui.ShowSuccess("✅ Clean complete")
		return nil
	})
}

// openChangelog opens the changelog for editing
func (m *Module) openChangelog() error {
	fmt.Println()
	ui.ShowInfo("📝 Opening CHANGELOG.md...")

	// Check if file exists
	changelogPath := "CHANGELOG.md"
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		ui.ShowInfo("CHANGELOG.md not found, creating...")
		if err := m.createDefaultChangelog(); err != nil {
			return fmt.Errorf("failed to create changelog: %v", err)
		}
	}

	// Get editor from environment or use default
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	ui.ShowInfo(fmt.Sprintf("Opening with %s...", editor))
	time.Sleep(1 * time.Second)

	return m.runCommand(editor, changelogPath)
}

// pushChanges pushes changes to remote
func (m *Module) pushChanges() error {
	fmt.Println()
	ui.ShowInfo("📤 Pushing changes...")
	fmt.Println()

	return ui.ShowLoadingAnimation("Pushing", func() error {
		// Push main branch
		if err := m.runCommand("git", "push", "origin", "main"); err != nil {
			return fmt.Errorf("failed to push main: %v", err)
		}

		// Push tags
		if err := m.runCommand("git", "push", "origin", "--tags"); err != nil {
			return fmt.Errorf("failed to push tags: %v", err)
		}

		ui.ShowSuccess("✅ Changes pushed successfully")
		return nil
	})
}

// pullChanges pulls changes from remote
func (m *Module) pullChanges() error {
	fmt.Println()
	ui.ShowInfo("📥 Pulling changes...")
	fmt.Println()

	return ui.ShowLoadingAnimation("Pulling", func() error {
		// Pull main branch
		if err := m.runCommand("git", "pull", "origin", "main"); err != nil {
			return fmt.Errorf("failed to pull main: %v", err)
		}

		// Pull tags
		if err := m.runCommand("git", "pull", "origin", "--tags"); err != nil {
			return fmt.Errorf("failed to pull tags: %v", err)
		}

		ui.ShowSuccess("✅ Changes pulled successfully")
		return nil
	})
}

// syncChanges syncs changes (pull then push)
func (m *Module) syncChanges() error {
	fmt.Println()
	ui.ShowInfo("🔄 Syncing changes...")
	fmt.Println()

	if err := m.pullChanges(); err != nil {
		return err
	}

	return m.pushChanges()
}

// showGitStatus shows git status
func (m *Module) showGitStatus() error {
	fmt.Println()
	ui.ShowInfo("📊 Git Status:")
	fmt.Println()

	if err := m.runCommand("git", "status"); err != nil {
		return fmt.Errorf("failed to get git status: %v", err)
	}

	fmt.Println()
	ui.ShowInfo("Press Enter to continue...")
	fmt.Scanln()
	return nil
}

// openGitHubIssues opens GitHub issues in browser
func (m *Module) openGitHubIssues() error {
	repo, err := m.getGitHubRepo()
	if err != nil {
		return fmt.Errorf("failed to get repository info: %v", err)
	}

	url := fmt.Sprintf("https://github.com/%s/issues", repo)
	ui.ShowInfo(fmt.Sprintf("Opening: %s", url))

	return m.runCommand("open", url)
}

// openGitHubPulls opens GitHub pull requests in browser
func (m *Module) openGitHubPulls() error {
	repo, err := m.getGitHubRepo()
	if err != nil {
		return fmt.Errorf("failed to get repository info: %v", err)
	}

	url := fmt.Sprintf("https://github.com/%s/pulls", repo)
	ui.ShowInfo(fmt.Sprintf("Opening: %s", url))

	return m.runCommand("open", url)
}

// generateReleaseNotes generates release notes from git log
func (m *Module) generateReleaseNotes() error {
	fmt.Println()

	currentVersion, err := m.getCurrentVersion()
	if err != nil {
		ui.ShowInfo("No previous version found, showing all commits")
		currentVersion = ""
	}

	ui.ShowInfo(fmt.Sprintf("📝 Release notes since %s:", currentVersion))
	fmt.Println()

	var cmd *exec.Cmd
	if currentVersion != "" {
		cmd = exec.Command("git", "log", fmt.Sprintf("%s..HEAD", currentVersion), "--pretty=format:- %s", "--reverse")
	} else {
		cmd = exec.Command("git", "log", "--pretty=format:- %s", "--reverse")
	}

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to generate release notes: %v", err)
	}

	notes := string(output)

	// Filter out merge commits
	lines := strings.Split(notes, "\n")
	var filteredLines []string
	for _, line := range lines {
		if !strings.Contains(strings.ToLower(line), "merge") {
			filteredLines = append(filteredLines, line)
		}
	}

	if len(filteredLines) == 0 {
		ui.ShowInfo("No new commits found")
	} else {
		for _, line := range filteredLines {
			if strings.TrimSpace(line) != "" {
				fmt.Println(line)
			}
		}
	}

	fmt.Println()
	ui.ShowInfo("Press Enter to continue...")
	fmt.Scanln()
	return nil
}

// updateChangelog helps update the CHANGELOG.md file
func (m *Module) updateChangelog() error {
	fmt.Println()
	ui.ShowInfo("📝 Update Changelog")
	fmt.Println()

	// Get current version
	currentVersion, err := m.getCurrentVersion()
	if err != nil {
		currentVersion = "v0.0.0"
	}

	// Calculate next versions
	nextVersions := m.calculateNextVersions(currentVersion)

	// Select version for changelog entry
	options := []string{
		fmt.Sprintf("Next Patch (%s)", nextVersions.Patch),
		fmt.Sprintf("Next Minor (%s)", nextVersions.Minor),
		fmt.Sprintf("Next Major (%s)", nextVersions.Major),
		"Custom Version",
		"Update Unreleased Section",
		"Back",
	}

	choice, err := ui.SelectFromList("Select version for changelog entry:", options)
	if err != nil || choice == 5 {
		return nil
	}

	var targetVersion string
	switch choice {
	case 0:
		targetVersion = nextVersions.Patch
	case 1:
		targetVersion = nextVersions.Minor
	case 2:
		targetVersion = nextVersions.Major
	case 3:
		customVersion, err := ui.GetInput(
			"Enter version (e.g., v1.2.3)",
			"",
			false,
			func(s string) error {
				if !strings.HasPrefix(s, "v") {
					return fmt.Errorf("version must start with 'v'")
				}
				return nil
			},
		)
		if err != nil {
			return err
		}
		targetVersion = customVersion
	case 4:
		targetVersion = "Unreleased"
	}

	// Select change type
	changeTypes := []string{
		"Added - for new features",
		"Changed - for changes in existing functionality",
		"Deprecated - for soon-to-be removed features",
		"Removed - for now removed features",
		"Fixed - for any bug fixes",
		"Security - in case of vulnerabilities",
		"Back",
	}

	changeChoice, err := ui.SelectFromList("Select change type:", changeTypes)
	if err != nil || changeChoice == 6 {
		return nil
	}

	changeType := []string{"Added", "Changed", "Deprecated", "Removed", "Fixed", "Security"}[changeChoice]

	// Get change description
	fmt.Println()
	ui.ShowInfo("Enter change description (press Enter twice to finish):")
	fmt.Println()

	var lines []string
	for {
		line, err := ui.GetInput("", "", false, nil)
		if err != nil {
			if err.Error() == "cancelled" {
				return nil
			}
			// Empty line
			if len(lines) > 0 && line == "" {
				break
			}
		}
		if line == "" && len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1] // Remove last empty line
			break
		}
		lines = append(lines, line)
	}

	if len(lines) == 0 {
		ui.ShowWarning("No changes entered")
		return nil
	}

	// Read current changelog
	content, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		return fmt.Errorf("failed to read CHANGELOG.md: %v", err)
	}

	// Update changelog
	updatedContent := m.insertChangelogEntry(string(content), targetVersion, changeType, lines)

	// Write updated changelog
	if err := os.WriteFile("CHANGELOG.md", []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write CHANGELOG.md: %v", err)
	}

	ui.ShowSuccess("✅ Changelog updated successfully!")

	// Show the changes
	fmt.Println()
	ui.ShowInfo("Changes added:")
	fmt.Printf("\n## [%s]\n\n### %s\n\n", targetVersion, changeType)
	for _, line := range lines {
		fmt.Printf("- %s\n", line)
	}
	fmt.Println()

	// Ask if user wants to open the changelog
	if ui.GetConfirmation("Open changelog to review?") {
		return m.openChangelog()
	}

	return nil
}

// insertChangelogEntry inserts a new entry into the changelog content
func (m *Module) insertChangelogEntry(content, version, changeType string, changes []string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inserted := false
	currentDate := time.Now().Format("2006-01-02")

	for i, line := range lines {
		// Look for the right place to insert
		if !inserted {
			if version == "Unreleased" && strings.HasPrefix(line, "## [Unreleased]") {
				// Add to existing Unreleased section
				result = append(result, line)

				// Find or create the change type section
				changeTypeFound := false
				j := i + 1
				for j < len(lines) && !strings.HasPrefix(lines[j], "##") {
					if strings.HasPrefix(lines[j], "### "+changeType) {
						changeTypeFound = true
						// Add changes under this section
						result = append(result, lines[j])
						j++
						// Skip empty lines
						for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
							result = append(result, lines[j])
							j++
						}
						// Add new changes
						for _, change := range changes {
							result = append(result, fmt.Sprintf("- %s", change))
						}
						// Continue with the rest
						for k := j; k < len(lines); k++ {
							result = append(result, lines[k])
						}
						return strings.Join(result, "\n")
					}
					result = append(result, lines[j])
					j++
				}

				if !changeTypeFound {
					// Add new change type section
					result = append(result, "")
					result = append(result, "### "+changeType)
					result = append(result, "")
					for _, change := range changes {
						result = append(result, fmt.Sprintf("- %s", change))
					}
					// Continue with the rest
					for k := j; k < len(lines); k++ {
						result = append(result, lines[k])
					}
					return strings.Join(result, "\n")
				}
			} else if strings.HasPrefix(line, "## [Unreleased]") && version != "Unreleased" {
				// Insert new version section after Unreleased
				result = append(result, line)
				result = append(result, "")
				result = append(result, fmt.Sprintf("## [%s] - %s", version, currentDate))
				result = append(result, "")
				result = append(result, "### "+changeType)
				result = append(result, "")
				for _, change := range changes {
					result = append(result, fmt.Sprintf("- %s", change))
				}
				inserted = true
			} else {
				result = append(result, line)
			}
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// Helper methods

// runCommandSilent runs a command without showing output
func (m *Module) runCommandSilent(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// getGitHubRepo gets the GitHub repository in owner/repo format
func (m *Module) getGitHubRepo() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	url := strings.TrimSpace(string(output))

	// Convert various GitHub URL formats to owner/repo format
	if strings.Contains(url, "github.com") {
		// Handle SSH URLs
		if strings.HasPrefix(url, "git@github.com:") {
			repo := strings.TrimPrefix(url, "git@github.com:")
			repo = strings.TrimSuffix(repo, ".git")
			return repo, nil
		}
		// Handle HTTPS URLs
		if strings.Contains(url, "github.com/") {
			parts := strings.Split(url, "github.com/")
			if len(parts) >= 2 {
				repo := parts[1]
				repo = strings.TrimSuffix(repo, ".git")
				return repo, nil
			}
		}
	}

	return "", fmt.Errorf("not a GitHub repository")
}

// compareVersions compares two version strings
func (m *Module) compareVersions(v1, v2 string) int {
	version1 := m.parseVersion(v1)
	version2 := m.parseVersion(v2)

	if version1.Major != version2.Major {
		return version1.Major - version2.Major
	}
	if version1.Minor != version2.Minor {
		return version1.Minor - version2.Minor
	}
	return version1.Patch - version2.Patch
}

// createDefaultChangelog creates a default changelog if it doesn't exist
func (m *Module) createDefaultChangelog() error {
	content := `# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release

### Changed

### Deprecated

### Removed

### Fixed

### Security
`

	return os.WriteFile("CHANGELOG.md", []byte(content), 0644)
}
