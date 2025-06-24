package fluttermanager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/ui"
)

// AndroidBuilder handles Android build operations
type AndroidBuilder struct {
	cfg *config.Config
}

// NewAndroidBuilder creates a new Android builder
func NewAndroidBuilder(cfg *config.Config) *AndroidBuilder {
	return &AndroidBuilder{cfg: cfg}
}

// Build executes the Android build based on the build type
func (ab *AndroidBuilder) Build(buildType BuildType) error {
	// Check Flutter installation
	if err := ab.checkFlutter(); err != nil {
		return err
	}

	switch buildType {
	case BuildAPKDebug:
		return ab.buildAPKDebug()
	case BuildAPKRelease:
		return ab.buildAPKRelease()
	case BuildAppBundle:
		return ab.buildAppBundle()
	case BuildSplitAPKs:
		return ab.buildSplitAPKs()
	case BuildCustomFlavor:
		return ab.buildWithFlavor()
	default:
		return fmt.Errorf("unknown build type")
	}
}

// checkFlutter verifies Flutter is installed and available
func (ab *AndroidBuilder) checkFlutter() error {
	cmd := exec.Command("flutter", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Flutter not found. Please install Flutter first")
	}
	return nil
}

// buildAPKDebug builds a debug APK
func (ab *AndroidBuilder) buildAPKDebug() error {
	ui.ShowInfo("Building Debug APK...")

	err := ui.ShowLoadingAnimation("Building debug APK", func() error {
		cmd := exec.Command("flutter", "build", "apk", "--debug")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	if err != nil {
		return fmt.Errorf("failed to build debug APK: %w", err)
	}

	// Find the output APK
	apkPath := filepath.Join("build", "app", "outputs", "flutter-apk", "app-debug.apk")
	if _, err := os.Stat(apkPath); err == nil {
		ui.ShowSuccess(fmt.Sprintf("✅ Debug APK built successfully: %s", apkPath))

		// Get file size
		if info, err := os.Stat(apkPath); err == nil {
			sizeMB := float64(info.Size()) / (1024 * 1024)
			ui.ShowInfo(fmt.Sprintf("📦 APK size: %.2f MB", sizeMB))
		}

		if ui.GetConfirmation("Open build directory?") {
			ab.openBuildDirectory()
		}
	}

	return nil
}

// buildAPKRelease builds a release APK
func (ab *AndroidBuilder) buildAPKRelease() error {
	// Check if signing is configured
	signingMgr := NewSigningManager(ab.cfg)
	status := signingMgr.GetStatus()

	if !status.IsConfigured {
		ui.ShowWarning("⚠️  Release builds require signing configuration")
		if ui.GetConfirmation("Configure signing now?") {
			if err := signingMgr.CreateKeystore(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("signing configuration required for release builds")
		}
	}

	ui.ShowInfo("Building Release APK...")

	// Check for obfuscation preference
	obfuscate := ui.GetConfirmation("Enable code obfuscation?")

	args := []string{"build", "apk", "--release"}
	if obfuscate {
		args = append(args, "--obfuscate", "--split-debug-info=build/debug-info")
	}

	err := ui.ShowLoadingAnimation("Building release APK", func() error {
		cmd := exec.Command("flutter", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	if err != nil {
		return fmt.Errorf("failed to build release APK: %w", err)
	}

	// Find the output APK
	apkPath := filepath.Join("build", "app", "outputs", "flutter-apk", "app-release.apk")
	if _, err := os.Stat(apkPath); err == nil {
		ui.ShowSuccess(fmt.Sprintf("✅ Release APK built successfully: %s", apkPath))

		// Get file size
		if info, err := os.Stat(apkPath); err == nil {
			sizeMB := float64(info.Size()) / (1024 * 1024)
			ui.ShowInfo(fmt.Sprintf("📦 APK size: %.2f MB", sizeMB))
		}

		if ui.GetConfirmation("Open build directory?") {
			ab.openBuildDirectory()
		}
	}

	return nil
}

// buildAppBundle builds an Android App Bundle
func (ab *AndroidBuilder) buildAppBundle() error {
	// Check if signing is configured
	signingMgr := NewSigningManager(ab.cfg)
	status := signingMgr.GetStatus()

	if !status.IsConfigured {
		ui.ShowWarning("⚠️  App Bundles require signing configuration")
		if ui.GetConfirmation("Configure signing now?") {
			if err := signingMgr.CreateKeystore(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("signing configuration required for app bundles")
		}
	}

	ui.ShowInfo("Building App Bundle (AAB)...")
	ui.ShowInfo("App Bundles are recommended for Google Play Store uploads")

	// Check for obfuscation preference
	obfuscate := ui.GetConfirmation("Enable code obfuscation?")

	args := []string{"build", "appbundle", "--release"}
	if obfuscate {
		args = append(args, "--obfuscate", "--split-debug-info=build/debug-info")
	}

	err := ui.ShowLoadingAnimation("Building app bundle", func() error {
		cmd := exec.Command("flutter", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	if err != nil {
		return fmt.Errorf("failed to build app bundle: %w", err)
	}

	// Find the output AAB
	aabPath := filepath.Join("build", "app", "outputs", "bundle", "release", "app-release.aab")
	if _, err := os.Stat(aabPath); err == nil {
		ui.ShowSuccess(fmt.Sprintf("✅ App Bundle built successfully: %s", aabPath))

		// Get file size
		if info, err := os.Stat(aabPath); err == nil {
			sizeMB := float64(info.Size()) / (1024 * 1024)
			ui.ShowInfo(fmt.Sprintf("📦 Bundle size: %.2f MB", sizeMB))
		}

		ui.ShowInfo("💡 Use bundletool to test the app bundle locally")
		ui.ShowInfo("💡 Upload this .aab file to Google Play Console")

		if ui.GetConfirmation("Open build directory?") {
			ab.openBuildDirectory()
		}
	}

	return nil
}

// buildSplitAPKs builds split APKs by ABI
func (ab *AndroidBuilder) buildSplitAPKs() error {
	ui.ShowInfo("Building Split APKs...")
	ui.ShowInfo("This will create separate APKs for each CPU architecture")

	// Check if signing is configured for release builds
	release := ui.GetConfirmation("Build release APKs? (No = Debug)")

	if release {
		signingMgr := NewSigningManager(ab.cfg)
		status := signingMgr.GetStatus()

		if !status.IsConfigured {
			ui.ShowWarning("⚠️  Release builds require signing configuration")
			if ui.GetConfirmation("Configure signing now?") {
				if err := signingMgr.CreateKeystore(); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("signing configuration required for release builds")
			}
		}
	}

	args := []string{"build", "apk", "--split-per-abi"}
	if release {
		args = append(args, "--release")
	} else {
		args = append(args, "--debug")
	}

	err := ui.ShowLoadingAnimation("Building split APKs", func() error {
		cmd := exec.Command("flutter", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	if err != nil {
		return fmt.Errorf("failed to build split APKs: %w", err)
	}

	// List generated APKs
	buildMode := "debug"
	if release {
		buildMode = "release"
	}

	apkDir := filepath.Join("build", "app", "outputs", "flutter-apk")
	ui.ShowSuccess("✅ Split APKs built successfully!")

	// List all APKs
	abis := []string{"armeabi-v7a", "arm64-v8a", "x86_64"}
	fmt.Println("\n📦 Generated APKs:")
	for _, abi := range abis {
		apkName := fmt.Sprintf("app-%s-%s.apk", abi, buildMode)
		apkPath := filepath.Join(apkDir, apkName)
		if info, err := os.Stat(apkPath); err == nil {
			sizeMB := float64(info.Size()) / (1024 * 1024)
			fmt.Printf("  • %s (%.2f MB)\n", apkName, sizeMB)
		}
	}

	if ui.GetConfirmation("Open build directory?") {
		ab.openBuildDirectory()
	}

	return nil
}

// buildWithFlavor builds with a custom flavor
func (ab *AndroidBuilder) buildWithFlavor() error {
	// Get available flavors from build.gradle
	flavors := ab.getAvailableFlavors()

	if len(flavors) == 0 {
		ui.ShowWarning("No flavors found in build.gradle")
		if !ui.GetConfirmation("Continue with custom flavor name?") {
			return nil
		}
	} else {
		fmt.Println("\n📋 Available flavors:")
		for _, flavor := range flavors {
			fmt.Printf("  • %s\n", flavor)
		}
	}

	// Get flavor name
	flavorName, err := ui.GetInput(
		"Enter flavor name",
		"production",
		false,
		func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("flavor name cannot be empty")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	// Get build type
	buildTypes := []string{"Debug", "Release"}
	buildTypeIdx, err := ui.SelectFromList("Select build type:", buildTypes)
	if err != nil {
		return err
	}

	// Get output type
	outputTypes := []string{"APK", "App Bundle"}
	outputTypeIdx, err := ui.SelectFromList("Select output type:", outputTypes)
	if err != nil {
		return err
	}

	// Build command
	args := []string{"build"}
	if outputTypeIdx == 0 {
		args = append(args, "apk")
	} else {
		args = append(args, "appbundle")
	}

	args = append(args, "--flavor", flavorName)

	if buildTypeIdx == 1 {
		args = append(args, "--release")

		// Check signing for release builds
		signingMgr := NewSigningManager(ab.cfg)
		status := signingMgr.GetStatus()

		if !status.IsConfigured {
			ui.ShowWarning("⚠️  Release builds require signing configuration")
			if ui.GetConfirmation("Configure signing now?") {
				if err := signingMgr.CreateKeystore(); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("signing configuration required for release builds")
			}
		}
	} else {
		args = append(args, "--debug")
	}

	ui.ShowInfo(fmt.Sprintf("Building %s flavor (%s)...", flavorName, buildTypes[buildTypeIdx]))

	err = ui.ShowLoadingAnimation("Building with flavor", func() error {
		cmd := exec.Command("flutter", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	if err != nil {
		return fmt.Errorf("failed to build with flavor: %w", err)
	}

	ui.ShowSuccess(fmt.Sprintf("✅ %s flavor built successfully!", flavorName))

	if ui.GetConfirmation("Open build directory?") {
		ab.openBuildDirectory()
	}

	return nil
}

// getAvailableFlavors attempts to parse available flavors from build.gradle
func (ab *AndroidBuilder) getAvailableFlavors() []string {
	// This is a simplified implementation
	// In a real scenario, you'd parse the build.gradle file
	buildGradlePath := filepath.Join("android", "app", "build.gradle")

	content, err := os.ReadFile(buildGradlePath)
	if err != nil {
		return []string{}
	}

	// Simple flavor detection (this is a basic implementation)
	flavors := []string{}
	lines := strings.Split(string(content), "\n")
	inFlavorBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "productFlavors") {
			inFlavorBlock = true
			continue
		}
		if inFlavorBlock && strings.Contains(trimmed, "}") {
			break
		}
		if inFlavorBlock && !strings.HasPrefix(trimmed, "//") {
			// Extract flavor name (basic parsing)
			parts := strings.Fields(trimmed)
			if len(parts) > 0 && !strings.Contains(parts[0], "{") {
				flavors = append(flavors, parts[0])
			}
		}
	}

	return flavors
}

// openBuildDirectory opens the build output directory
func (ab *AndroidBuilder) openBuildDirectory() error {
	buildDir := filepath.Join("build", "app", "outputs")

	var cmd *exec.Cmd
	switch os := strings.ToLower(string(os.PathSeparator)); {
	case strings.Contains(os, "darwin"):
		cmd = exec.Command("open", buildDir)
	case strings.Contains(os, "linux"):
		cmd = exec.Command("xdg-open", buildDir)
	case strings.Contains(os, "windows"):
		cmd = exec.Command("explorer", buildDir)
	default:
		ui.ShowInfo(fmt.Sprintf("Build directory: %s", buildDir))
		return nil
	}

	return cmd.Run()
}
