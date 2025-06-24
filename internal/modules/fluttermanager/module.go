package fluttermanager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/types"
	"github.com/kkz6/devtools/internal/ui"
)

// Module implements the Flutter application management module
type Module struct{}

// New creates a new Flutter manager module
func New() *Module {
	return &Module{}
}

// Info returns module information
func (m *Module) Info() types.ModuleInfo {
	return types.ModuleInfo{
		ID:          "flutter-manager",
		Name:        "Flutter Application Manager",
		Description: "Manage Flutter builds, versions, signing, and backups",
	}
}

// Execute runs the Flutter management interface
func (m *Module) Execute(cfg *config.Config) error {
	ui.ShowBanner()

	title := ui.GetGradientTitle("📱 Flutter Application Manager")
	fmt.Println(title)
	fmt.Println()

	// Check if we're in a Flutter project
	if !isFlutterProject() {
		ui.ShowWarning("Not in a Flutter project directory")
		if !ui.GetConfirmation("Continue anyway?") {
			return types.ErrNavigateBack
		}
	}

	options := []string{
		"Build Android (APK/Bundle)",
		"Version Manager",
		"Signing Manager",
		"Backup Signing Configuration",
		"Project Setup & Dependencies",
		"Clean & Rebuild",
		"Device Management",
		"Back to main menu",
	}

	choice, err := ui.SelectFromList("Select an option:", options)
	if err != nil {
		return nil
	}

	switch choice {
	case 0:
		return m.buildAndroid(cfg)
	case 1:
		return m.manageVersion(cfg)
	case 2:
		return m.manageSigning(cfg)
	case 3:
		return m.backupSigning(cfg)
	case 4:
		return m.projectSetup(cfg)
	case 5:
		return m.cleanAndRebuild(cfg)
	case 6:
		return m.deviceManagement(cfg)
	case 7:
		return types.ErrNavigateBack
	default:
		ui.ShowError("Invalid choice")
		return m.Execute(cfg)
	}
}

// isFlutterProject checks if the current directory is a Flutter project
func isFlutterProject() bool {
	_, err := os.Stat("pubspec.yaml")
	return err == nil
}

// buildAndroid handles Android build operations
func (m *Module) buildAndroid(cfg *config.Config) error {
	builder := NewAndroidBuilder(cfg)

	fmt.Println()
	ui.ShowInfo("📱 Android Build Options")
	fmt.Println()

	options := []string{
		"Build APK (Debug)",
		"Build APK (Release)",
		"Build App Bundle (Release)",
		"Build Split APKs",
		"Build with custom flavor",
		"Back",
	}

	choice, err := ui.SelectFromList("Select build type:", options)
	if err != nil || choice == 5 {
		return m.Execute(cfg)
	}

	return builder.Build(BuildType(choice))
}

// manageVersion handles version management
func (m *Module) manageVersion(cfg *config.Config) error {
	versionMgr := NewVersionManager(cfg)

	fmt.Println()
	ui.ShowInfo("🔢 Version Management")
	fmt.Println()

	// Show current version
	currentVersion, buildNumber, err := versionMgr.GetCurrentVersion()
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to get current version: %v", err))
		return m.Execute(cfg)
	}

	ui.ShowInfo(fmt.Sprintf("Current version: %s+%s", currentVersion, buildNumber))
	fmt.Println()

	options := []string{
		"Bump patch version (0.0.x)",
		"Bump minor version (0.x.0)",
		"Bump major version (x.0.0)",
		"Set custom version",
		"Auto-increment build number",
		"Version history",
		"Back",
	}

	choice, err := ui.SelectFromList("Select version action:", options)
	if err != nil || choice == 6 {
		return m.Execute(cfg)
	}

	switch choice {
	case 0, 1, 2:
		return versionMgr.BumpVersion(VersionBumpType(choice))
	case 3:
		return versionMgr.SetCustomVersion()
	case 4:
		return versionMgr.IncrementBuildNumber()
	case 5:
		return versionMgr.ShowHistory()
	}

	return m.Execute(cfg)
}

// manageSigning handles signing configuration
func (m *Module) manageSigning(cfg *config.Config) error {
	signingMgr := NewSigningManager(cfg)

	fmt.Println()
	ui.ShowInfo("🔐 Signing Configuration")
	fmt.Println()

	// Check current signing status
	status := signingMgr.GetStatus()
	if status.IsConfigured {
		ui.ShowSuccess("✅ Signing is configured")
		if status.KeystorePath != "" {
			fmt.Printf("   Keystore: %s\n", filepath.Base(status.KeystorePath))
		}
	} else {
		ui.ShowWarning("⚠️  Signing is not configured")
	}
	fmt.Println()

	options := []string{
		"Create new keystore",
		"Import existing keystore",
		"View signing configuration",
		"Update keystore password",
		"Export signing configuration",
		"Verify keystore",
		"Back",
	}

	choice, err := ui.SelectFromList("Select signing action:", options)
	if err != nil || choice == 6 {
		return m.Execute(cfg)
	}

	switch choice {
	case 0:
		return signingMgr.CreateKeystore()
	case 1:
		return signingMgr.ImportKeystore()
	case 2:
		return signingMgr.ViewConfiguration()
	case 3:
		return signingMgr.UpdatePassword()
	case 4:
		return signingMgr.ExportConfiguration()
	case 5:
		return signingMgr.VerifyKeystore()
	}

	return m.Execute(cfg)
}

// backupSigning handles signing configuration backup
func (m *Module) backupSigning(cfg *config.Config) error {
	backupMgr := NewBackupManager(cfg)

	fmt.Println()
	ui.ShowInfo("💾 Backup Signing Configuration")
	fmt.Println()

	options := []string{
		"Create backup archive",
		"Restore from backup",
		"List existing backups",
		"Verify backup integrity",
		"Export to cloud storage",
		"Back",
	}

	choice, err := ui.SelectFromList("Select backup action:", options)
	if err != nil || choice == 5 {
		return m.Execute(cfg)
	}

	switch choice {
	case 0:
		return backupMgr.CreateBackup()
	case 1:
		return backupMgr.RestoreBackup()
	case 2:
		return backupMgr.ListBackups()
	case 3:
		return backupMgr.VerifyBackup()
	case 4:
		return backupMgr.ExportToCloud()
	}

	return m.Execute(cfg)
}

// projectSetup handles project setup and dependencies
func (m *Module) projectSetup(cfg *config.Config) error {
	setupMgr := NewSetupManager(cfg)

	fmt.Println()
	ui.ShowInfo("🛠️  Project Setup & Dependencies")
	fmt.Println()

	options := []string{
		"Check Flutter environment",
		"Update dependencies",
		"Configure Android SDK",
		"Setup Firebase",
		"Configure build flavors",
		"Generate launcher icons",
		"Back",
	}

	choice, err := ui.SelectFromList("Select setup action:", options)
	if err != nil || choice == 6 {
		return m.Execute(cfg)
	}

	switch choice {
	case 0:
		return setupMgr.CheckEnvironment()
	case 1:
		return setupMgr.UpdateDependencies()
	case 2:
		return setupMgr.ConfigureAndroidSDK()
	case 3:
		return setupMgr.SetupFirebase()
	case 4:
		return setupMgr.ConfigureFlavors()
	case 5:
		return setupMgr.GenerateIcons()
	}

	return m.Execute(cfg)
}

// cleanAndRebuild handles cleaning and rebuilding
func (m *Module) cleanAndRebuild(cfg *config.Config) error {
	fmt.Println()
	ui.ShowInfo("🧹 Clean & Rebuild")
	fmt.Println()

	options := []string{
		"Flutter clean",
		"Clean build cache",
		"Reset pods (iOS)",
		"Clean and get packages",
		"Full project reset",
		"Back",
	}

	choice, err := ui.SelectFromList("Select clean action:", options)
	if err != nil || choice == 5 {
		return m.Execute(cfg)
	}

	cleaner := NewCleaner(cfg)

	switch choice {
	case 0:
		return cleaner.FlutterClean()
	case 1:
		return cleaner.CleanBuildCache()
	case 2:
		return cleaner.ResetPods()
	case 3:
		return cleaner.CleanAndGet()
	case 4:
		return cleaner.FullReset()
	}

	return m.Execute(cfg)
}

// deviceManagement handles device and emulator management
func (m *Module) deviceManagement(cfg *config.Config) error {
	deviceMgr := NewDeviceManager(cfg)

	fmt.Println()
	ui.ShowInfo("📱 Device Management")
	fmt.Println()

	// Show connected devices
	devices, err := deviceMgr.ListDevices()
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to list devices: %v", err))
	} else if len(devices) > 0 {
		fmt.Println("Connected devices:")
		for _, device := range devices {
			fmt.Printf("  • %s (%s)\n", device.Name, device.ID)
		}
		fmt.Println()
	} else {
		ui.ShowWarning("No devices connected")
		fmt.Println()
	}

	options := []string{
		"Launch Android emulator",
		"Launch iOS simulator",
		"Install APK to device",
		"Stream device logs",
		"Take screenshot",
		"Back",
	}

	choice, err := ui.SelectFromList("Select device action:", options)
	if err != nil || choice == 5 {
		return m.Execute(cfg)
	}

	switch choice {
	case 0:
		return deviceMgr.LaunchAndroidEmulator()
	case 1:
		return deviceMgr.LaunchIOSSimulator()
	case 2:
		return deviceMgr.InstallAPK()
	case 3:
		return deviceMgr.StreamLogs()
	case 4:
		return deviceMgr.TakeScreenshot()
	}

	return m.Execute(cfg)
}
