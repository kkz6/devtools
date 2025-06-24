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

// SetupManager handles project setup and configuration
type SetupManager struct {
	cfg *config.Config
}

// NewSetupManager creates a new setup manager
func NewSetupManager(cfg *config.Config) *SetupManager {
	return &SetupManager{cfg: cfg}
}

// CheckEnvironment checks Flutter environment setup
func (sm *SetupManager) CheckEnvironment() error {
	ui.ShowInfo("🔍 Checking Flutter Environment")
	fmt.Println()

	// Run flutter doctor
	err := ui.ShowLoadingAnimation("Running flutter doctor", func() error {
		cmd := exec.Command("flutter", "doctor", "-v")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	if err != nil {
		ui.ShowError("Failed to run flutter doctor")
		return err
	}

	fmt.Println()

	// Additional checks
	checks := []struct {
		name    string
		command []string
	}{
		{"Flutter Version", []string{"flutter", "--version"}},
		{"Dart Version", []string{"dart", "--version"}},
		{"Java Version", []string{"java", "-version"}},
		{"Android SDK", []string{"sdkmanager", "--version"}},
	}

	fmt.Println("\n📋 Additional Information:")
	for _, check := range checks {
		cmd := exec.Command(check.command[0], check.command[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("❌ %s: Not found or error\n", check.name)
		} else {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 && lines[0] != "" {
				fmt.Printf("✅ %s: %s\n", check.name, strings.TrimSpace(lines[0]))
			}
		}
	}

	fmt.Print("\nPress Enter to continue...")
	fmt.Scanln()

	return nil
}

// UpdateDependencies updates Flutter dependencies
func (sm *SetupManager) UpdateDependencies() error {
	ui.ShowInfo("📦 Updating Dependencies")
	fmt.Println()

	options := []string{
		"flutter pub get",
		"flutter pub upgrade",
		"flutter pub upgrade --major-versions",
		"flutter pub outdated",
		"Add new dependency",
		"Remove dependency",
		"Back",
	}

	choice, err := ui.SelectFromList("Select action:", options)
	if err != nil || choice == 6 {
		return nil
	}

	switch choice {
	case 0:
		return sm.runPubGet()
	case 1:
		return sm.runPubUpgrade(false)
	case 2:
		return sm.runPubUpgrade(true)
	case 3:
		return sm.checkOutdated()
	case 4:
		return sm.addDependency()
	case 5:
		return sm.removeDependency()
	}

	return nil
}

// ConfigureAndroidSDK configures Android SDK settings
func (sm *SetupManager) ConfigureAndroidSDK() error {
	ui.ShowInfo("🤖 Android SDK Configuration")
	fmt.Println()

	// Check ANDROID_HOME
	androidHome := os.Getenv("ANDROID_HOME")
	if androidHome == "" {
		androidHome = os.Getenv("ANDROID_SDK_ROOT")
	}

	if androidHome == "" {
		ui.ShowWarning("ANDROID_HOME environment variable not set")

		androidHome, err := ui.GetInput(
			"Enter Android SDK path",
			"",
			false,
			func(s string) error {
				if _, err := os.Stat(s); err != nil {
					return fmt.Errorf("directory does not exist")
				}
				return nil
			},
		)
		if err != nil {
			return err
		}

		ui.ShowInfo(fmt.Sprintf("Add this to your shell profile:"))
		ui.ShowInfo(fmt.Sprintf("export ANDROID_HOME=%s", androidHome))
		ui.ShowInfo(fmt.Sprintf("export PATH=$PATH:$ANDROID_HOME/tools:$ANDROID_HOME/platform-tools"))
	} else {
		ui.ShowSuccess(fmt.Sprintf("✅ Android SDK found: %s", androidHome))
	}

	// Check installed packages
	fmt.Println("\n📋 Checking Android SDK packages...")

	cmd := exec.Command("sdkmanager", "--list_installed")
	output, err := cmd.Output()
	if err != nil {
		ui.ShowError("Failed to list SDK packages")
		return nil
	}

	fmt.Println(string(output))

	// Offer to install common packages
	if ui.GetConfirmation("Install/Update common SDK packages?") {
		packages := []string{
			"platform-tools",
			"build-tools;33.0.0",
			"platforms;android-33",
			"cmdline-tools;latest",
		}

		for _, pkg := range packages {
			ui.ShowInfo(fmt.Sprintf("Installing %s...", pkg))
			cmd := exec.Command("sdkmanager", pkg)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
	}

	// Accept licenses
	if ui.GetConfirmation("Accept Android SDK licenses?") {
		cmd := exec.Command("flutter", "doctor", "--android-licenses")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	return nil
}

// SetupFirebase sets up Firebase for the project
func (sm *SetupManager) SetupFirebase() error {
	ui.ShowInfo("🔥 Firebase Setup")
	fmt.Println()

	// Check if Firebase CLI is installed
	cmd := exec.Command("firebase", "--version")
	if err := cmd.Run(); err != nil {
		ui.ShowWarning("Firebase CLI not found")
		ui.ShowInfo("Install with: npm install -g firebase-tools")

		if ui.GetConfirmation("Install Firebase CLI now?") {
			cmd := exec.Command("npm", "install", "-g", "firebase-tools")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				ui.ShowError("Failed to install Firebase CLI")
				return nil
			}
		} else {
			return nil
		}
	}

	options := []string{
		"Initialize Firebase project",
		"Add Firebase to Flutter app",
		"Configure Firebase services",
		"Download google-services.json",
		"Back",
	}

	choice, err := ui.SelectFromList("Select Firebase action:", options)
	if err != nil || choice == 4 {
		return nil
	}

	switch choice {
	case 0:
		cmd := exec.Command("firebase", "init")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()

	case 1:
		// Check if FlutterFire CLI is installed
		cmd := exec.Command("flutterfire", "--version")
		if err := cmd.Run(); err != nil {
			ui.ShowInfo("Installing FlutterFire CLI...")
			cmd := exec.Command("dart", "pub", "global", "activate", "flutterfire_cli")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				ui.ShowError("Failed to install FlutterFire CLI")
				return nil
			}
		}

		ui.ShowInfo("Configuring Firebase for Flutter...")
		cmd = exec.Command("flutterfire", "configure")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()

	case 2:
		ui.ShowInfo("Add Firebase packages to pubspec.yaml:")
		ui.ShowInfo("  firebase_core: ^2.24.0")
		ui.ShowInfo("  firebase_auth: ^4.15.0")
		ui.ShowInfo("  cloud_firestore: ^4.13.0")
		ui.ShowInfo("  firebase_storage: ^11.5.0")
		ui.ShowInfo("  firebase_messaging: ^14.7.0")

		if ui.GetConfirmation("Add Firebase core package?") {
			cmd := exec.Command("flutter", "pub", "add", "firebase_core")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}

	case 3:
		ui.ShowInfo("Download google-services.json from Firebase Console:")
		ui.ShowInfo("1. Go to Firebase Console")
		ui.ShowInfo("2. Select your project")
		ui.ShowInfo("3. Go to Project Settings")
		ui.ShowInfo("4. Download google-services.json")
		ui.ShowInfo("5. Place it in android/app/")
	}

	return nil
}

// ConfigureFlavors configures build flavors
func (sm *SetupManager) ConfigureFlavors() error {
	ui.ShowInfo("🎨 Configure Build Flavors")
	fmt.Println()

	ui.ShowInfo("Build flavors allow you to create different versions of your app")
	ui.ShowInfo("(e.g., development, staging, production)")
	fmt.Println()

	if !ui.GetConfirmation("Set up build flavors?") {
		return nil
	}

	// Get flavor names
	flavors := []string{}
	for {
		flavor, err := ui.GetInput(
			fmt.Sprintf("Enter flavor name #%d (or press Enter to finish)", len(flavors)+1),
			"",
			false,
			nil,
		)
		if err != nil || flavor == "" {
			break
		}
		flavors = append(flavors, strings.ToLower(flavor))
	}

	if len(flavors) == 0 {
		ui.ShowWarning("No flavors added")
		return nil
	}

	// Generate flavor configuration
	ui.ShowInfo("\nAdd this to your android/app/build.gradle:")
	fmt.Println("\n```gradle")
	fmt.Println("flavorDimensions \"env\"")
	fmt.Println("productFlavors {")

	for _, flavor := range flavors {
		fmt.Printf("    %s {\n", flavor)
		fmt.Printf("        dimension \"env\"\n")
		fmt.Printf("        applicationIdSuffix \".%s\"\n", flavor)
		fmt.Printf("        versionNameSuffix \"-%s\"\n", flavor)
		fmt.Printf("    }\n")
	}

	fmt.Println("}")
	fmt.Println("```")

	// Create flavor directories
	if ui.GetConfirmation("Create flavor directories?") {
		for _, flavor := range flavors {
			flavorDir := filepath.Join("lib", "flavors")
			if err := os.MkdirAll(flavorDir, 0755); err == nil {
				// Create flavor config file
				configContent := fmt.Sprintf(`class FlavorConfig {
  final String name;
  final String apiUrl;
  
  FlavorConfig({
    required this.name,
    required this.apiUrl,
  });
}

final flavorConfig%s = FlavorConfig(
  name: '%s',
  apiUrl: 'https://api-%s.example.com',
);
`, strings.Title(flavor), flavor, flavor)

				configFile := filepath.Join(flavorDir, fmt.Sprintf("flavor_%s.dart", flavor))
				os.WriteFile(configFile, []byte(configContent), 0644)
			}
		}

		ui.ShowSuccess("✅ Flavor directories created")
	}

	// Show how to run with flavors
	fmt.Println("\n📱 Run with flavors:")
	for _, flavor := range flavors {
		fmt.Printf("  flutter run --flavor %s\n", flavor)
		fmt.Printf("  flutter build apk --flavor %s\n", flavor)
	}

	return nil
}

// GenerateIcons generates app launcher icons
func (sm *SetupManager) GenerateIcons() error {
	ui.ShowInfo("🎨 Generate Launcher Icons")
	fmt.Println()

	// Check if flutter_launcher_icons is in pubspec
	pubspecContent, err := os.ReadFile("pubspec.yaml")
	if err != nil {
		ui.ShowError("Failed to read pubspec.yaml")
		return nil
	}

	if !strings.Contains(string(pubspecContent), "flutter_launcher_icons") {
		ui.ShowInfo("Adding flutter_launcher_icons package...")

		// Add to dev dependencies
		cmd := exec.Command("flutter", "pub", "add", "--dev", "flutter_launcher_icons")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			ui.ShowError("Failed to add flutter_launcher_icons")
			return nil
		}
	}

	// Get icon path
	iconPath, err := ui.GetInput(
		"Path to icon image (1024x1024 PNG recommended)",
		"assets/icon/icon.png",
		false,
		func(s string) error {
			if !strings.HasSuffix(strings.ToLower(s), ".png") {
				return fmt.Errorf("icon must be a PNG file")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	// Check if icon exists
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		ui.ShowWarning(fmt.Sprintf("Icon file not found: %s", iconPath))
		ui.ShowInfo("Please ensure the icon file exists before generating")
		return nil
	}

	// Create flutter_launcher_icons configuration
	config := fmt.Sprintf(`flutter_launcher_icons:
  android: true
  ios: true
  image_path: "%s"
  adaptive_icon_background: "#ffffff"
  adaptive_icon_foreground: "%s"
`, iconPath, iconPath)

	// Append to pubspec.yaml
	ui.ShowInfo("Add this to your pubspec.yaml:")
	fmt.Println(config)

	if ui.GetConfirmation("Generate icons now?") {
		err := ui.ShowLoadingAnimation("Generating icons", func() error {
			cmd := exec.Command("flutter", "pub", "run", "flutter_launcher_icons")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		})

		if err != nil {
			ui.ShowError("Failed to generate icons")
			return nil
		}

		ui.ShowSuccess("✅ Launcher icons generated successfully!")
		ui.ShowInfo("Icons have been generated for both Android and iOS")
	}

	return nil
}

// Helper methods

func (sm *SetupManager) runPubGet() error {
	return ui.ShowLoadingAnimation("Running flutter pub get", func() error {
		cmd := exec.Command("flutter", "pub", "get")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
}

func (sm *SetupManager) runPubUpgrade(major bool) error {
	args := []string{"pub", "upgrade"}
	if major {
		args = append(args, "--major-versions")
	}

	return ui.ShowLoadingAnimation("Upgrading dependencies", func() error {
		cmd := exec.Command("flutter", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
}

func (sm *SetupManager) checkOutdated() error {
	cmd := exec.Command("flutter", "pub", "outdated")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (sm *SetupManager) addDependency() error {
	packageName, err := ui.GetInput("Package name", "", false, nil)
	if err != nil || packageName == "" {
		return nil
	}

	isDev := ui.GetConfirmation("Add as dev dependency?")

	args := []string{"pub", "add"}
	if isDev {
		args = append(args, "--dev")
	}
	args = append(args, packageName)

	cmd := exec.Command("flutter", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (sm *SetupManager) removeDependency() error {
	// Read pubspec.yaml to list dependencies
	pubspecContent, err := os.ReadFile("pubspec.yaml")
	if err != nil {
		ui.ShowError("Failed to read pubspec.yaml")
		return nil
	}

	// Simple parsing (in production, use a YAML parser)
	lines := strings.Split(string(pubspecContent), "\n")
	dependencies := []string{}
	inDeps := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "dependencies:" || trimmed == "dev_dependencies:" {
			inDeps = true
			continue
		}
		if inDeps && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			inDeps = false
		}
		if inDeps && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			parts := strings.Split(trimmed, ":")
			if len(parts) > 0 {
				dependencies = append(dependencies, strings.TrimSpace(parts[0]))
			}
		}
	}

	if len(dependencies) == 0 {
		ui.ShowWarning("No dependencies found")
		return nil
	}

	choice, err := ui.SelectFromList("Select dependency to remove:", dependencies)
	if err != nil {
		return nil
	}

	packageName := dependencies[choice]

	cmd := exec.Command("flutter", "pub", "remove", packageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
