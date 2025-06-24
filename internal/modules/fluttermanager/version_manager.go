package fluttermanager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/ui"
)

// VersionManager handles version management for Flutter projects
type VersionManager struct {
	cfg *config.Config
}

// NewVersionManager creates a new version manager
func NewVersionManager(cfg *config.Config) *VersionManager {
	return &VersionManager{cfg: cfg}
}

// GetCurrentVersion returns the current version and build number
func (vm *VersionManager) GetCurrentVersion() (string, string, error) {
	pubspecPath := "pubspec.yaml"

	content, err := os.ReadFile(pubspecPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read pubspec.yaml: %w", err)
	}

	// Parse version line (format: version: 1.0.0+1)
	versionRegex := regexp.MustCompile(`version:\s*(\d+\.\d+\.\d+)\+(\d+)`)
	matches := versionRegex.FindStringSubmatch(string(content))

	if len(matches) < 3 {
		// Try without build number
		versionRegex = regexp.MustCompile(`version:\s*(\d+\.\d+\.\d+)`)
		matches = versionRegex.FindStringSubmatch(string(content))
		if len(matches) < 2 {
			return "", "", fmt.Errorf("version not found in pubspec.yaml")
		}
		return matches[1], "1", nil
	}

	return matches[1], matches[2], nil
}

// BumpVersion increments the version based on the bump type
func (vm *VersionManager) BumpVersion(bumpType VersionBumpType) error {
	currentVersion, buildNumber, err := vm.GetCurrentVersion()
	if err != nil {
		return err
	}

	// Parse version parts
	parts := strings.Split(currentVersion, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid version format: %s", currentVersion)
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	// Bump version based on type
	switch bumpType {
	case BumpPatch:
		patch++
	case BumpMinor:
		minor++
		patch = 0
	case BumpMajor:
		major++
		minor = 0
		patch = 0
	}

	newVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)

	// Increment build number
	buildNum, _ := strconv.Atoi(buildNumber)
	buildNum++
	newBuildNumber := strconv.Itoa(buildNum)

	// Show changes
	ui.ShowInfo(fmt.Sprintf("Current version: %s+%s", currentVersion, buildNumber))
	ui.ShowInfo(fmt.Sprintf("New version: %s+%s", newVersion, newBuildNumber))

	if !ui.GetConfirmation("Apply version bump?") {
		return nil
	}

	// Update pubspec.yaml
	if err := vm.updatePubspecVersion(newVersion, newBuildNumber); err != nil {
		return err
	}

	// Create git tag if desired
	if ui.GetConfirmation("Create git tag?") {
		tagName := fmt.Sprintf("v%s", newVersion)
		if err := vm.createGitTag(tagName, fmt.Sprintf("Version %s", newVersion)); err != nil {
			ui.ShowWarning(fmt.Sprintf("Failed to create git tag: %v", err))
		} else {
			ui.ShowSuccess(fmt.Sprintf("Created git tag: %s", tagName))
		}
	}

	// Save to version history
	vm.saveVersionHistory(newVersion, newBuildNumber)

	ui.ShowSuccess(fmt.Sprintf("✅ Version bumped to %s+%s", newVersion, newBuildNumber))
	return nil
}

// SetCustomVersion allows setting a custom version
func (vm *VersionManager) SetCustomVersion() error {
	currentVersion, buildNumber, err := vm.GetCurrentVersion()
	if err != nil {
		ui.ShowWarning("Could not get current version")
		currentVersion = "1.0.0"
		buildNumber = "1"
	}

	ui.ShowInfo(fmt.Sprintf("Current version: %s+%s", currentVersion, buildNumber))

	// Get new version
	newVersion, err := ui.GetInput(
		"Enter new version (x.y.z)",
		currentVersion,
		false,
		func(s string) error {
			// Validate version format
			versionRegex := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
			if !versionRegex.MatchString(s) {
				return fmt.Errorf("invalid version format. Use x.y.z (e.g., 1.2.3)")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	// Get new build number
	newBuildNumber, err := ui.GetInput(
		"Enter build number",
		buildNumber,
		false,
		func(s string) error {
			if _, err := strconv.Atoi(s); err != nil {
				return fmt.Errorf("build number must be a valid integer")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	// Update pubspec.yaml
	if err := vm.updatePubspecVersion(newVersion, newBuildNumber); err != nil {
		return err
	}

	// Save to version history
	vm.saveVersionHistory(newVersion, newBuildNumber)

	ui.ShowSuccess(fmt.Sprintf("✅ Version set to %s+%s", newVersion, newBuildNumber))
	return nil
}

// IncrementBuildNumber increments only the build number
func (vm *VersionManager) IncrementBuildNumber() error {
	currentVersion, buildNumber, err := vm.GetCurrentVersion()
	if err != nil {
		return err
	}

	buildNum, _ := strconv.Atoi(buildNumber)
	buildNum++
	newBuildNumber := strconv.Itoa(buildNum)

	ui.ShowInfo(fmt.Sprintf("Current: %s+%s", currentVersion, buildNumber))
	ui.ShowInfo(fmt.Sprintf("New: %s+%s", currentVersion, newBuildNumber))

	if !ui.GetConfirmation("Increment build number?") {
		return nil
	}

	// Update pubspec.yaml
	if err := vm.updatePubspecVersion(currentVersion, newBuildNumber); err != nil {
		return err
	}

	ui.ShowSuccess(fmt.Sprintf("✅ Build number incremented to %s", newBuildNumber))
	return nil
}

// ShowHistory displays version history
func (vm *VersionManager) ShowHistory() error {
	historyFile := filepath.Join(".flutter", "version_history.txt")

	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		ui.ShowWarning("No version history found")
		return nil
	}

	content, err := os.ReadFile(historyFile)
	if err != nil {
		return fmt.Errorf("failed to read version history: %w", err)
	}

	fmt.Println("\n📜 Version History:")
	fmt.Println(string(content))

	fmt.Print("\nPress Enter to continue...")
	fmt.Scanln()

	return nil
}

// updatePubspecVersion updates the version in pubspec.yaml
func (vm *VersionManager) updatePubspecVersion(version, buildNumber string) error {
	pubspecPath := "pubspec.yaml"

	// Read file
	content, err := os.ReadFile(pubspecPath)
	if err != nil {
		return fmt.Errorf("failed to read pubspec.yaml: %w", err)
	}

	// Update version line
	versionLine := fmt.Sprintf("version: %s+%s", version, buildNumber)
	versionRegex := regexp.MustCompile(`version:\s*\S+`)
	updatedContent := versionRegex.ReplaceAllString(string(content), versionLine)

	// Write back
	if err := os.WriteFile(pubspecPath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write pubspec.yaml: %w", err)
	}

	return nil
}

// createGitTag creates a git tag for the version
func (vm *VersionManager) createGitTag(tagName, message string) error {
	// Check if git is available
	if err := exec.Command("git", "status").Run(); err != nil {
		return fmt.Errorf("git not available or not in a git repository")
	}

	// Create annotated tag
	cmd := exec.Command("git", "tag", "-a", tagName, "-m", message)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	// Ask to push tag
	if ui.GetConfirmation("Push tag to remote?") {
		cmd = exec.Command("git", "push", "origin", tagName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to push tag: %w", err)
		}
	}

	return nil
}

// saveVersionHistory saves version change to history file
func (vm *VersionManager) saveVersionHistory(version, buildNumber string) {
	// Create .flutter directory if it doesn't exist
	flutterDir := ".flutter"
	if err := os.MkdirAll(flutterDir, 0755); err != nil {
		return
	}

	historyFile := filepath.Join(flutterDir, "version_history.txt")

	// Open file in append mode
	file, err := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	// Get git commit hash if available
	commitHash := "unknown"
	if output, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output(); err == nil {
		commitHash = strings.TrimSpace(string(output))
	}

	// Write entry
	entry := fmt.Sprintf("%s | Version: %s+%s | Commit: %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		version,
		buildNumber,
		commitHash,
	)

	file.WriteString(entry)
}

// GetVersionFromGitTag gets the latest version from git tags
func (vm *VersionManager) GetVersionFromGitTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("no git tags found")
	}

	tag := strings.TrimSpace(string(output))
	// Remove 'v' prefix if present
	if strings.HasPrefix(tag, "v") {
		tag = tag[1:]
	}

	return tag, nil
}

// AutoBumpForCI automatically bumps version for CI/CD pipelines
func (vm *VersionManager) AutoBumpForCI(bumpType string) error {
	// Map string to VersionBumpType
	var versionBumpType VersionBumpType
	switch strings.ToLower(bumpType) {
	case "patch":
		versionBumpType = BumpPatch
	case "minor":
		versionBumpType = BumpMinor
	case "major":
		versionBumpType = BumpMajor
	default:
		return fmt.Errorf("invalid bump type: %s (use patch, minor, or major)", bumpType)
	}

	// Get current version
	currentVersion, buildNumber, err := vm.GetCurrentVersion()
	if err != nil {
		return err
	}

	// Parse and bump version
	parts := strings.Split(currentVersion, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid version format: %s", currentVersion)
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	switch versionBumpType {
	case BumpPatch:
		patch++
	case BumpMinor:
		minor++
		patch = 0
	case BumpMajor:
		major++
		minor = 0
		patch = 0
	}

	newVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)

	// Use build number from CI if available
	ciBuildNumber := os.Getenv("BUILD_NUMBER")
	if ciBuildNumber == "" {
		ciBuildNumber = os.Getenv("GITHUB_RUN_NUMBER")
	}
	if ciBuildNumber == "" {
		buildNum, _ := strconv.Atoi(buildNumber)
		buildNum++
		ciBuildNumber = strconv.Itoa(buildNum)
	}

	// Update version
	if err := vm.updatePubspecVersion(newVersion, ciBuildNumber); err != nil {
		return err
	}

	fmt.Printf("Version bumped from %s+%s to %s+%s\n",
		currentVersion, buildNumber, newVersion, ciBuildNumber)

	return nil
}

// ValidateVersion validates the current version format
func (vm *VersionManager) ValidateVersion() error {
	version, buildNumber, err := vm.GetCurrentVersion()
	if err != nil {
		return err
	}

	// Validate semantic version
	versionRegex := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	if !versionRegex.MatchString(version) {
		return fmt.Errorf("invalid version format: %s (expected x.y.z)", version)
	}

	// Validate build number
	if _, err := strconv.Atoi(buildNumber); err != nil {
		return fmt.Errorf("invalid build number: %s (expected integer)", buildNumber)
	}

	ui.ShowSuccess(fmt.Sprintf("✅ Version %s+%s is valid", version, buildNumber))
	return nil
}

// CompareVersions compares two semantic versions
func (vm *VersionManager) CompareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < 3; i++ {
		num1, _ := strconv.Atoi(parts1[i])
		num2, _ := strconv.Atoi(parts2[i])

		if num1 > num2 {
			return 1
		} else if num1 < num2 {
			return -1
		}
	}

	return 0
}

// ExportVersionInfo exports version information to a file
func (vm *VersionManager) ExportVersionInfo() error {
	version, buildNumber, err := vm.GetCurrentVersion()
	if err != nil {
		return err
	}

	// Create version info
	info := fmt.Sprintf(`Flutter App Version Information
==============================
Version: %s
Build Number: %s
Date: %s
`, version, buildNumber, time.Now().Format("2006-01-02 15:04:05"))

	// Add git info if available
	if commitHash, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
		info += fmt.Sprintf("Git Commit: %s", strings.TrimSpace(string(commitHash)))
	}

	if branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		info += fmt.Sprintf("Git Branch: %s", strings.TrimSpace(string(branch)))
	}

	// Write to file
	filename := fmt.Sprintf("version-info-%s.txt", version)
	if err := os.WriteFile(filename, []byte(info), 0644); err != nil {
		return fmt.Errorf("failed to write version info: %w", err)
	}

	ui.ShowSuccess(fmt.Sprintf("Version info exported to %s", filename))
	return nil
}
