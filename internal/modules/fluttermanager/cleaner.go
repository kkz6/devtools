package fluttermanager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/ui"
)

// Cleaner handles cleaning and rebuilding operations
type Cleaner struct {
	cfg *config.Config
}

// NewCleaner creates a new cleaner
func NewCleaner(cfg *config.Config) *Cleaner {
	return &Cleaner{cfg: cfg}
}

// FlutterClean runs flutter clean
func (c *Cleaner) FlutterClean() error {
	ui.ShowInfo("🧹 Running Flutter clean...")

	err := ui.ShowLoadingAnimation("Cleaning Flutter project", func() error {
		cmd := exec.Command("flutter", "clean")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	if err != nil {
		return fmt.Errorf("flutter clean failed: %w", err)
	}

	ui.ShowSuccess("✅ Flutter project cleaned")

	// Ask if user wants to get packages
	if ui.GetConfirmation("Run flutter pub get?") {
		err = ui.ShowLoadingAnimation("Getting packages", func() error {
			cmd := exec.Command("flutter", "pub", "get")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		})

		if err != nil {
			ui.ShowError("Failed to get packages")
		} else {
			ui.ShowSuccess("✅ Packages restored")
		}
	}

	return nil
}

// CleanBuildCache cleans build cache
func (c *Cleaner) CleanBuildCache() error {
	ui.ShowInfo("🧹 Cleaning build cache...")

	// Directories to clean
	dirsToClean := []string{
		"build",
		".dart_tool",
		".flutter-plugins",
		".flutter-plugins-dependencies",
		"android/.gradle",
		"android/app/build",
		"ios/.symlinks",
		"ios/Pods",
		"ios/Flutter/Flutter.framework",
		"ios/Flutter/Flutter.podspec",
	}

	// Calculate total size before cleaning
	var totalSize int64
	for _, dir := range dirsToClean {
		size, _ := c.getDirSize(dir)
		totalSize += size
	}

	if totalSize > 0 {
		sizeMB := float64(totalSize) / (1024 * 1024)
		ui.ShowInfo(fmt.Sprintf("📊 Total cache size: %.2f MB", sizeMB))

		if !ui.GetConfirmation("Clean build cache?") {
			return nil
		}
	}

	// Clean directories
	cleanedSize := int64(0)
	for _, dir := range dirsToClean {
		if _, err := os.Stat(dir); err == nil {
			size, _ := c.getDirSize(dir)

			ui.ShowInfo(fmt.Sprintf("Cleaning %s...", dir))
			if err := os.RemoveAll(dir); err != nil {
				ui.ShowWarning(fmt.Sprintf("Failed to clean %s: %v", dir, err))
			} else {
				cleanedSize += size
			}
		}
	}

	// Clean individual files
	filesToClean := []string{
		".packages",
		"pubspec.lock",
	}

	for _, file := range filesToClean {
		if _, err := os.Stat(file); err == nil {
			os.Remove(file)
		}
	}

	if cleanedSize > 0 {
		cleanedMB := float64(cleanedSize) / (1024 * 1024)
		ui.ShowSuccess(fmt.Sprintf("✅ Cleaned %.2f MB of cache", cleanedMB))
	} else {
		ui.ShowInfo("No cache to clean")
	}

	return nil
}

// ResetPods resets iOS CocoaPods
func (c *Cleaner) ResetPods() error {
	ui.ShowInfo("🍎 Resetting iOS Pods...")

	// Check if iOS directory exists
	if _, err := os.Stat("ios"); os.IsNotExist(err) {
		ui.ShowWarning("iOS directory not found")
		return nil
	}

	// Check if pod is installed
	if err := exec.Command("pod", "--version").Run(); err != nil {
		ui.ShowError("CocoaPods not installed")
		ui.ShowInfo("Install with: sudo gem install cocoapods")
		return nil
	}

	steps := []struct {
		name string
		fn   func() error
	}{
		{
			name: "Cleaning Pods directory",
			fn: func() error {
				return os.RemoveAll("ios/Pods")
			},
		},
		{
			name: "Removing Podfile.lock",
			fn: func() error {
				return os.Remove("ios/Podfile.lock")
			},
		},
		{
			name: "Deintegrating pods",
			fn: func() error {
				cmd := exec.Command("pod", "deintegrate")
				cmd.Dir = "ios"
				return cmd.Run()
			},
		},
		{
			name: "Cleaning pod cache",
			fn: func() error {
				cmd := exec.Command("pod", "cache", "clean", "--all")
				return cmd.Run()
			},
		},
		{
			name: "Installing pods",
			fn: func() error {
				cmd := exec.Command("pod", "install", "--repo-update")
				cmd.Dir = "ios"
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				return cmd.Run()
			},
		},
	}

	for _, step := range steps {
		err := ui.ShowLoadingAnimation(step.name, step.fn)
		if err != nil {
			ui.ShowWarning(fmt.Sprintf("Failed: %s - %v", step.name, err))
			// Continue with other steps
		}
	}

	ui.ShowSuccess("✅ iOS Pods reset complete")
	return nil
}

// CleanAndGet cleans and gets packages
func (c *Cleaner) CleanAndGet() error {
	ui.ShowInfo("🧹 Clean and get packages...")

	// Run flutter clean
	err := ui.ShowLoadingAnimation("Running flutter clean", func() error {
		cmd := exec.Command("flutter", "clean")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	if err != nil {
		ui.ShowError("Flutter clean failed")
		return err
	}

	// Delete pubspec.lock
	os.Remove("pubspec.lock")

	// Run flutter pub get
	err = ui.ShowLoadingAnimation("Getting packages", func() error {
		cmd := exec.Command("flutter", "pub", "get")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	if err != nil {
		ui.ShowError("Failed to get packages")
		return err
	}

	ui.ShowSuccess("✅ Clean and get packages complete")

	// Offer to run on iOS if directory exists
	if _, err := os.Stat("ios"); err == nil {
		if ui.GetConfirmation("Update iOS pods?") {
			cmd := exec.Command("pod", "install")
			cmd.Dir = "ios"
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				ui.ShowWarning("Failed to update iOS pods")
			}
		}
	}

	return nil
}

// FullReset performs a full project reset
func (c *Cleaner) FullReset() error {
	ui.ShowWarning("⚠️  Full project reset will delete all build artifacts and caches")
	ui.ShowWarning("This includes: build files, dependencies, IDE files, etc.")

	if !ui.GetConfirmation("Are you sure you want to perform a full reset?") {
		return nil
	}

	// Comprehensive list of items to clean
	itemsToClean := []string{
		// Flutter/Dart
		"build",
		".dart_tool",
		".packages",
		"pubspec.lock",
		".flutter-plugins",
		".flutter-plugins-dependencies",

		// Android
		"android/.gradle",
		"android/gradle",
		"android/gradlew",
		"android/gradlew.bat",
		"android/local.properties",
		"android/app/build",
		"android/.idea",
		"android/*.iml",
		"android/app/*.iml",

		// iOS
		"ios/Pods",
		"ios/Podfile.lock",
		"ios/.symlinks",
		"ios/Flutter/Flutter.framework",
		"ios/Flutter/Flutter.podspec",
		"ios/Flutter/Generated.xcconfig",
		"ios/Flutter/app.flx",
		"ios/Flutter/app.zip",
		"ios/Flutter/flutter_assets",
		"ios/ServiceDefinitions.json",
		"ios/Runner/GeneratedPluginRegistrant.*",
		"ios/*.xcworkspace",

		// IDE
		".idea",
		".vscode/launch.json",
		"*.iml",
		"*.ipr",
		"*.iws",
		".metadata",

		// Misc
		"*.log",
		".DS_Store",
		"Thumbs.db",
	}

	ui.ShowInfo("📋 Items to clean:")
	for _, item := range itemsToClean {
		if _, err := os.Stat(item); err == nil {
			fmt.Printf("  • %s\n", item)
		}
	}
	fmt.Println()

	if !ui.GetConfirmation("Proceed with full reset?") {
		return nil
	}

	// Perform cleaning
	err := ui.ShowLoadingAnimation("Performing full reset", func() error {
		for _, item := range itemsToClean {
			if info, err := os.Stat(item); err == nil {
				if info.IsDir() {
					os.RemoveAll(item)
				} else {
					os.Remove(item)
				}
			}
		}
		return nil
	})

	if err != nil {
		ui.ShowError("Full reset failed")
		return err
	}

	ui.ShowSuccess("✅ Full reset complete")

	// Restore essential files
	ui.ShowInfo("Restoring essential files...")

	// Run flutter create to restore Android/iOS files
	if ui.GetConfirmation("Restore Android/iOS project files?") {
		projectName := filepath.Base(c.getCurrentDirectory())

		err = ui.ShowLoadingAnimation("Restoring project files", func() error {
			cmd := exec.Command("flutter", "create", "--project-name", projectName, ".")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		})

		if err != nil {
			ui.ShowError("Failed to restore project files")
		}
	}

	// Get packages
	if ui.GetConfirmation("Run flutter pub get?") {
		cmd := exec.Command("flutter", "pub", "get")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	ui.ShowInfo("💡 You may need to reconfigure:")
	ui.ShowInfo("  • Signing configuration")
	ui.ShowInfo("  • Firebase configuration")
	ui.ShowInfo("  • Custom Android/iOS settings")

	return nil
}

// Helper methods

// getDirSize calculates directory size
func (c *Cleaner) getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// getCurrentDirectory gets the current working directory name
func (c *Cleaner) getCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		return "app"
	}
	return filepath.Base(dir)
}
