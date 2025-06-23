package releasemanager

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/types"
	"github.com/kkz6/devtools/internal/ui"
)

// Module implements the release manager module
type Module struct{}

// New creates a new release manager module
func New() *Module {
	return &Module{}
}

// Info returns module information
func (m *Module) Info() types.ModuleInfo {
	return types.ModuleInfo{
		ID:          "release-manager",
		Name:        "Release Manager",
		Description: "Manage DevTools releases and development workflow",
	}
}

// Execute runs the release manager
func (m *Module) Execute(cfg *config.Config) error {
	ui.ShowBanner()

	title := ui.GetGradientTitle("🚀 Release Manager")
	fmt.Println(title)
	fmt.Println()

	// Show current version
	if version, err := m.getCurrentVersion(); err == nil {
		ui.ShowInfo(fmt.Sprintf("Current Version: %s", version))
		fmt.Println()
	}

	for {
		options := []string{
			"Create Release",
			"Tag Management",
			"Development Workflow",
			"Git Operations",
			"GitHub Integration",
			"Project Status",
			"Back to main menu",
		}

		choice, err := ui.SelectFromList("Select release management action:", options)
		if err != nil || choice == 6 {
			return types.ErrNavigateBack
		}

		switch choice {
		case 0:
			if err := m.handleReleaseMenu(); err != nil {
				ui.ShowError(fmt.Sprintf("Release error: %v", err))
			}
		case 1:
			if err := m.handleTagMenu(); err != nil {
				ui.ShowError(fmt.Sprintf("Tag management error: %v", err))
			}
		case 2:
			if err := m.handleDevelopmentMenu(); err != nil {
				ui.ShowError(fmt.Sprintf("Development error: %v", err))
			}
		case 3:
			if err := m.handleGitMenu(); err != nil {
				ui.ShowError(fmt.Sprintf("Git operation error: %v", err))
			}
		case 4:
			if err := m.handleGitHubMenu(); err != nil {
				ui.ShowError(fmt.Sprintf("GitHub integration error: %v", err))
			}
		case 5:
			if err := m.showProjectStatus(); err != nil {
				ui.ShowError(fmt.Sprintf("Status error: %v", err))
			}
		}

		time.Sleep(1 * time.Second)
	}
}

// handleReleaseMenu handles release creation
func (m *Module) handleReleaseMenu() error {
	fmt.Println()
	currentVersion, err := m.getCurrentVersion()
	if err != nil {
		currentVersion = "v0.0.0"
	}

	nextVersions := m.calculateNextVersions(currentVersion)

	ui.ShowInfo(fmt.Sprintf("Current Version: %s", currentVersion))
	fmt.Println()

	options := []string{
		fmt.Sprintf("Patch Release (%s) - Bug fixes", nextVersions.Patch),
		fmt.Sprintf("Minor Release (%s) - New features", nextVersions.Minor),
		fmt.Sprintf("Major Release (%s) - Breaking changes", nextVersions.Major),
		fmt.Sprintf("Alpha Release (%s-alpha.1)", nextVersions.Minor),
		fmt.Sprintf("Beta Release (%s-beta.1)", nextVersions.Minor),
		fmt.Sprintf("RC Release (%s-rc.1)", nextVersions.Minor),
		"Custom Version",
		"Back",
	}

	choice, err := ui.SelectFromList("Select release type:", options)
	if err != nil || choice == 7 {
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
		targetVersion = nextVersions.Minor + "-alpha.1"
	case 4:
		targetVersion = nextVersions.Minor + "-beta.1"
	case 5:
		targetVersion = nextVersions.Minor + "-rc.1"
	case 6:
		customVersion, err := ui.GetInput(
			"Enter custom version (e.g., v1.2.3)",
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
	}

	return m.createRelease(targetVersion)
}

// handleTagMenu handles tag management
func (m *Module) handleTagMenu() error {
	fmt.Println()
	options := []string{
		"List All Tags",
		"Delete Tag",
		"Back",
	}

	choice, err := ui.SelectFromList("Tag Management:", options)
	if err != nil || choice == 2 {
		return nil
	}

	switch choice {
	case 0:
		return m.listTags()
	case 1:
		return m.deleteTag()
	}

	return nil
}

// handleDevelopmentMenu handles development workflow
func (m *Module) handleDevelopmentMenu() error {
	fmt.Println()
	options := []string{
		"Build Locally",
		"Install Locally",
		"Run Tests",
		"Run Linter",
		"Clean Build Artifacts",
		"Update Changelog",
		"Open Changelog",
		"Back",
	}

	choice, err := ui.SelectFromList("Development Workflow:", options)
	if err != nil || choice == 7 {
		return nil
	}

	switch choice {
	case 0:
		return m.buildLocal()
	case 1:
		return m.installLocal()
	case 2:
		return m.runTests()
	case 3:
		return m.runLinter()
	case 4:
		return m.cleanArtifacts()
	case 5:
		return m.updateChangelog()
	case 6:
		return m.openChangelog()
	}

	return nil
}

// handleGitMenu handles git operations
func (m *Module) handleGitMenu() error {
	fmt.Println()
	options := []string{
		"Push Changes",
		"Pull Changes",
		"Sync (Pull + Push)",
		"Git Status",
		"Back",
	}

	choice, err := ui.SelectFromList("Git Operations:", options)
	if err != nil || choice == 4 {
		return nil
	}

	switch choice {
	case 0:
		return m.pushChanges()
	case 1:
		return m.pullChanges()
	case 2:
		return m.syncChanges()
	case 3:
		return m.showGitStatus()
	}

	return nil
}

// handleGitHubMenu handles GitHub integration
func (m *Module) handleGitHubMenu() error {
	fmt.Println()
	options := []string{
		"Open Issues",
		"Open Pull Requests",
		"Generate Release Notes",
		"Back",
	}

	choice, err := ui.SelectFromList("GitHub Integration:", options)
	if err != nil || choice == 3 {
		return nil
	}

	switch choice {
	case 0:
		return m.openGitHubIssues()
	case 1:
		return m.openGitHubPulls()
	case 2:
		return m.generateReleaseNotes()
	}

	return nil
}

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
}

// NextVersions holds calculated next versions
type NextVersions struct {
	Patch string
	Minor string
	Major string
}

// getCurrentVersion gets the current version from git tags
func (m *Module) getCurrentVersion() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return "v0.0.0", nil // Default if no tags exist
	}
	return strings.TrimSpace(string(output)), nil
}

// calculateNextVersions calculates next patch, minor, and major versions
func (m *Module) calculateNextVersions(currentVersion string) NextVersions {
	version := m.parseVersion(currentVersion)

	return NextVersions{
		Patch: fmt.Sprintf("v%d.%d.%d", version.Major, version.Minor, version.Patch+1),
		Minor: fmt.Sprintf("v%d.%d.0", version.Major, version.Minor+1),
		Major: fmt.Sprintf("v%d.0.0", version.Major+1),
	}
}

// parseVersion parses a version string into components
func (m *Module) parseVersion(versionStr string) Version {
	// Remove 'v' prefix and any pre-release suffixes
	versionStr = strings.TrimPrefix(versionStr, "v")
	if idx := strings.Index(versionStr, "-"); idx != -1 {
		versionStr = versionStr[:idx]
	}

	parts := strings.Split(versionStr, ".")
	version := Version{Major: 0, Minor: 0, Patch: 0}

	if len(parts) >= 1 {
		if major, err := strconv.Atoi(parts[0]); err == nil {
			version.Major = major
		}
	}
	if len(parts) >= 2 {
		if minor, err := strconv.Atoi(parts[1]); err == nil {
			version.Minor = minor
		}
	}
	if len(parts) >= 3 {
		if patch, err := strconv.Atoi(parts[2]); err == nil {
			version.Patch = patch
		}
	}

	return version
}

// createRelease creates a new release with the specified version
func (m *Module) createRelease(version string) error {
	fmt.Println()
	ui.ShowInfo(fmt.Sprintf("Creating release %s...", version))
	fmt.Println()

	// Pre-release checklist
	ui.ShowInfo("📋 Pre-release checklist:")
	fmt.Println("  ☐ All changes committed?")
	fmt.Println("  ☐ Tests passing?")
	fmt.Println("  ☐ CHANGELOG.md updated?")
	fmt.Println()

	proceed, err := ui.GetInput(
		"Continue with release? (y/N)",
		"N",
		false,
		nil,
	)
	if err != nil {
		return err
	}

	if strings.ToLower(proceed) != "y" {
		ui.ShowInfo("Release cancelled")
		return nil
	}

	// Get tag message
	message, err := ui.GetInput(
		"Tag message",
		fmt.Sprintf("Release %s", version),
		false,
		func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("tag message cannot be empty")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	return ui.ShowLoadingAnimation("Creating release", func() error {
		// Create git tag
		if err := m.runCommand("git", "tag", "-a", version, "-m", message); err != nil {
			return fmt.Errorf("failed to create tag: %v", err)
		}

		// Push changes and tags
		if err := m.runCommand("git", "push", "origin", "main"); err != nil {
			return fmt.Errorf("failed to push changes: %v", err)
		}

		if err := m.runCommand("git", "push", "origin", version); err != nil {
			return fmt.Errorf("failed to push tag: %v", err)
		}

		ui.ShowSuccess(fmt.Sprintf("✅ Release %s created!", version))
		fmt.Println()
		ui.ShowInfo("🚀 GitHub Actions will now:")
		fmt.Println("  • Build binaries for all platforms")
		fmt.Println("  • Create GitHub release")
		fmt.Println("  • Upload release assets")
		fmt.Println()

		if repo, err := m.getGitHubRepo(); err == nil {
			ui.ShowInfo("📊 Monitor progress at:")
			fmt.Printf("  https://github.com/%s/actions\n", repo)
			fmt.Println()
			ui.ShowInfo("📦 View release at:")
			fmt.Printf("  https://github.com/%s/releases/tag/%s\n", repo, version)
		}

		return nil
	})
}

// Helper method to run commands
func (m *Module) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
