package fluttermanager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/ui"
)

// DeviceManager handles device and emulator management
type DeviceManager struct {
	cfg *config.Config
}

// NewDeviceManager creates a new device manager
func NewDeviceManager(cfg *config.Config) *DeviceManager {
	return &DeviceManager{cfg: cfg}
}

// ListDevices lists all connected devices
func (dm *DeviceManager) ListDevices() ([]Device, error) {
	cmd := exec.Command("flutter", "devices")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	devices := []Device{}
	lines := strings.Split(string(output), "\n")

	// Parse device list (skip header lines)
	deviceRegex := regexp.MustCompile(`^(.+?)\s+•\s+(.+?)\s+•\s+(.+?)\s+•\s+(.+?)$`)

	for _, line := range lines {
		matches := deviceRegex.FindStringSubmatch(strings.TrimSpace(line))
		if len(matches) == 5 {
			device := Device{
				Name:     strings.TrimSpace(matches[1]),
				ID:       strings.TrimSpace(matches[2]),
				Platform: strings.TrimSpace(matches[3]),
				IsActive: true,
			}
			devices = append(devices, device)
		}
	}

	return devices, nil
}

// LaunchAndroidEmulator launches an Android emulator
func (dm *DeviceManager) LaunchAndroidEmulator() error {
	ui.ShowInfo("🤖 Android Emulator Management")
	fmt.Println()

	// Check if emulator is available
	if err := exec.Command("emulator", "-version").Run(); err != nil {
		ui.ShowError("Android emulator not found")
		ui.ShowInfo("Please install Android Studio or command line tools")
		return nil
	}

	// List available AVDs
	cmd := exec.Command("emulator", "-list-avds")
	output, err := cmd.Output()
	if err != nil {
		ui.ShowError("Failed to list AVDs")
		return nil
	}

	avds := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(avds) == 0 || (len(avds) == 1 && avds[0] == "") {
		ui.ShowWarning("No Android Virtual Devices (AVDs) found")
		ui.ShowInfo("Create an AVD using Android Studio or avdmanager")

		if ui.GetConfirmation("Open AVD Manager instructions?") {
			fmt.Println("\nTo create an AVD:")
			fmt.Println("1. Open Android Studio")
			fmt.Println("2. Go to Tools > AVD Manager")
			fmt.Println("3. Click 'Create Virtual Device'")
			fmt.Println("4. Select a device and system image")
			fmt.Println("\nOr use command line:")
			fmt.Println("avdmanager create avd -n MyDevice -k 'system-images;android-33;google_apis;x86_64'")
		}
		return nil
	}

	// Select AVD
	fmt.Println("Available AVDs:")
	for i, avd := range avds {
		fmt.Printf("%d. %s\n", i+1, avd)
	}

	choice, err := ui.SelectFromList("Select AVD to launch:", avds)
	if err != nil {
		return nil
	}

	selectedAVD := avds[choice]

	// Launch emulator
	ui.ShowInfo(fmt.Sprintf("Launching %s...", selectedAVD))

	cmd = exec.Command("emulator", "-avd", selectedAVD)
	if err := cmd.Start(); err != nil {
		ui.ShowError(fmt.Sprintf("Failed to launch emulator: %v", err))
		return nil
	}

	ui.ShowSuccess(fmt.Sprintf("✅ Emulator %s launched", selectedAVD))

	// Wait for device to be ready
	if ui.GetConfirmation("Wait for emulator to be ready?") {
		dm.waitForDevice()
	}

	return nil
}

// LaunchIOSSimulator launches an iOS simulator
func (dm *DeviceManager) LaunchIOSSimulator() error {
	ui.ShowInfo("🍎 iOS Simulator Management")
	fmt.Println()

	// Check if on macOS
	if err := exec.Command("xcrun", "--version").Run(); err != nil {
		ui.ShowError("iOS Simulator is only available on macOS with Xcode installed")
		return nil
	}

	// List available simulators
	cmd := exec.Command("xcrun", "simctl", "list", "devices", "available")
	output, err := cmd.Output()
	if err != nil {
		ui.ShowError("Failed to list simulators")
		return nil
	}

	// Parse simulator list
	simulators := dm.parseSimulatorList(string(output))
	if len(simulators) == 0 {
		ui.ShowWarning("No iOS simulators found")
		ui.ShowInfo("Install simulators through Xcode > Preferences > Components")
		return nil
	}

	// Display simulators by category
	fmt.Println("Available iOS Simulators:")
	simNames := []string{}
	simMap := make(map[string]string)

	for category, sims := range simulators {
		fmt.Printf("\n%s:\n", category)
		for _, sim := range sims {
			displayName := fmt.Sprintf("%s (%s)", sim.name, category)
			simNames = append(simNames, displayName)
			simMap[displayName] = sim.id
			fmt.Printf("  • %s\n", sim.name)
		}
	}

	fmt.Println()
	choice, err := ui.SelectFromList("Select simulator to launch:", simNames)
	if err != nil {
		return nil
	}

	selectedID := simMap[simNames[choice]]

	// Boot simulator
	ui.ShowInfo("Booting simulator...")
	cmd = exec.Command("xcrun", "simctl", "boot", selectedID)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Check if already booted
		if !strings.Contains(string(output), "already booted") {
			ui.ShowError(fmt.Sprintf("Failed to boot simulator: %s", string(output)))
			return nil
		}
	}

	// Open Simulator app
	cmd = exec.Command("open", "-a", "Simulator")
	if err := cmd.Run(); err != nil {
		ui.ShowError("Failed to open Simulator app")
		return nil
	}

	ui.ShowSuccess("✅ iOS Simulator launched")

	// Wait for device to be ready
	if ui.GetConfirmation("Wait for simulator to be ready?") {
		dm.waitForDevice()
	}

	return nil
}

// InstallAPK installs an APK to a device
func (dm *DeviceManager) InstallAPK() error {
	ui.ShowInfo("📲 Install APK to Device")
	fmt.Println()

	// List connected devices
	devices, err := dm.ListDevices()
	if err != nil || len(devices) == 0 {
		ui.ShowWarning("No devices connected")
		ui.ShowInfo("Connect a device or launch an emulator first")
		return nil
	}

	// Select device if multiple
	var selectedDevice Device
	if len(devices) == 1 {
		selectedDevice = devices[0]
		ui.ShowInfo(fmt.Sprintf("Installing to: %s", selectedDevice.Name))
	} else {
		deviceNames := make([]string, len(devices))
		for i, device := range devices {
			deviceNames[i] = fmt.Sprintf("%s (%s)", device.Name, device.Platform)
		}

		choice, err := ui.SelectFromList("Select target device:", deviceNames)
		if err != nil {
			return nil
		}
		selectedDevice = devices[choice]
	}

	// Find APK files
	apkFiles := dm.findAPKFiles()
	if len(apkFiles) == 0 {
		ui.ShowWarning("No APK files found")
		ui.ShowInfo("Build an APK first using 'Build Android (APK/Bundle)'")
		return nil
	}

	// Select APK
	var selectedAPK string
	if len(apkFiles) == 1 {
		selectedAPK = apkFiles[0]
	} else {
		choice, err := ui.SelectFromList("Select APK to install:", apkFiles)
		if err != nil {
			return nil
		}
		selectedAPK = apkFiles[choice]
	}

	// Install APK
	ui.ShowInfo(fmt.Sprintf("Installing %s...", filepath.Base(selectedAPK)))

	err = ui.ShowLoadingAnimation("Installing APK", func() error {
		cmd := exec.Command("flutter", "install", "-d", selectedDevice.ID, "--apk", selectedAPK)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	if err != nil {
		ui.ShowError("Failed to install APK")
		return err
	}

	ui.ShowSuccess("✅ APK installed successfully")

	// Launch app
	if ui.GetConfirmation("Launch the app?") {
		cmd := exec.Command("flutter", "run", "-d", selectedDevice.ID)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	return nil
}

// StreamLogs streams device logs
func (dm *DeviceManager) StreamLogs() error {
	ui.ShowInfo("📋 Stream Device Logs")
	fmt.Println()

	// List connected devices
	devices, err := dm.ListDevices()
	if err != nil || len(devices) == 0 {
		ui.ShowWarning("No devices connected")
		return nil
	}

	// Select device
	var selectedDevice Device
	if len(devices) == 1 {
		selectedDevice = devices[0]
	} else {
		deviceNames := make([]string, len(devices))
		for i, device := range devices {
			deviceNames[i] = fmt.Sprintf("%s (%s)", device.Name, device.Platform)
		}

		choice, err := ui.SelectFromList("Select device:", deviceNames)
		if err != nil {
			return nil
		}
		selectedDevice = devices[choice]
	}

	ui.ShowInfo(fmt.Sprintf("Streaming logs from: %s", selectedDevice.Name))
	ui.ShowInfo("Press Ctrl+C to stop")
	fmt.Println()

	// Stream logs based on platform
	var cmd *exec.Cmd
	if strings.Contains(strings.ToLower(selectedDevice.Platform), "android") {
		// Use adb logcat for Android
		cmd = exec.Command("adb", "-s", selectedDevice.ID, "logcat")

		// Add filter options
		if ui.GetConfirmation("Filter logs by app package?") {
			// Try to get package name from pubspec
			packageName := dm.getAndroidPackageName()
			if packageName != "" {
				cmd = exec.Command("adb", "-s", selectedDevice.ID, "logcat", "--pid=$(pidof "+packageName+")")
			}
		}
	} else if strings.Contains(strings.ToLower(selectedDevice.Platform), "ios") {
		// Use idevicesyslog for iOS (requires libimobiledevice)
		cmd = exec.Command("idevicesyslog", "-u", selectedDevice.ID)

		// Check if idevicesyslog is available
		if err := exec.Command("which", "idevicesyslog").Run(); err != nil {
			ui.ShowWarning("idevicesyslog not found")
			ui.ShowInfo("Install with: brew install libimobiledevice")

			// Fallback to flutter logs
			ui.ShowInfo("Using flutter logs instead...")
			cmd = exec.Command("flutter", "logs", "-d", selectedDevice.ID)
		}
	} else {
		// Generic flutter logs
		cmd = exec.Command("flutter", "logs", "-d", selectedDevice.ID)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start streaming
	if err := cmd.Start(); err != nil {
		ui.ShowError("Failed to start log streaming")
		return err
	}

	// Wait for interrupt
	cmd.Wait()

	return nil
}

// TakeScreenshot takes a screenshot from device
func (dm *DeviceManager) TakeScreenshot() error {
	ui.ShowInfo("📸 Take Device Screenshot")
	fmt.Println()

	// List connected devices
	devices, err := dm.ListDevices()
	if err != nil || len(devices) == 0 {
		ui.ShowWarning("No devices connected")
		return nil
	}

	// Select device
	var selectedDevice Device
	if len(devices) == 1 {
		selectedDevice = devices[0]
	} else {
		deviceNames := make([]string, len(devices))
		for i, device := range devices {
			deviceNames[i] = fmt.Sprintf("%s (%s)", device.Name, device.Platform)
		}

		choice, err := ui.SelectFromList("Select device:", deviceNames)
		if err != nil {
			return nil
		}
		selectedDevice = devices[choice]
	}

	// Create screenshots directory
	screenshotsDir := "screenshots"
	if err := os.MkdirAll(screenshotsDir, 0755); err != nil {
		ui.ShowError("Failed to create screenshots directory")
		return err
	}

	// Generate filename
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("screenshot-%s-%s.png",
		strings.ReplaceAll(selectedDevice.Name, " ", "-"),
		timestamp)
	filepath := filepath.Join(screenshotsDir, filename)

	ui.ShowInfo(fmt.Sprintf("Taking screenshot from: %s", selectedDevice.Name))

	err = ui.ShowLoadingAnimation("Capturing screenshot", func() error {
		if strings.Contains(strings.ToLower(selectedDevice.Platform), "android") {
			// Use adb for Android
			cmd := exec.Command("adb", "-s", selectedDevice.ID, "exec-out", "screencap", "-p")
			output, err := cmd.Output()
			if err != nil {
				return err
			}
			return os.WriteFile(filepath, output, 0644)
		} else if strings.Contains(strings.ToLower(selectedDevice.Platform), "ios") {
			// Use idevicescreenshot for iOS
			cmd := exec.Command("idevicescreenshot", "-u", selectedDevice.ID, filepath)
			if err := cmd.Run(); err != nil {
				// Fallback to xcrun for simulators
				cmd = exec.Command("xcrun", "simctl", "io", selectedDevice.ID, "screenshot", filepath)
				return cmd.Run()
			}
			return nil
		} else {
			// Use flutter screenshot
			cmd := exec.Command("flutter", "screenshot", "-d", selectedDevice.ID, "-o", filepath)
			return cmd.Run()
		}
	})

	if err != nil {
		ui.ShowError("Failed to take screenshot")
		return err
	}

	ui.ShowSuccess(fmt.Sprintf("✅ Screenshot saved: %s", filepath))

	// Open screenshot
	if ui.GetConfirmation("Open screenshot?") {
		dm.openFile(filepath)
	}

	return nil
}

// Helper methods

// waitForDevice waits for a device to be ready
func (dm *DeviceManager) waitForDevice() {
	ui.ShowInfo("Waiting for device to be ready...")

	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		devices, err := dm.ListDevices()
		if err == nil && len(devices) > 0 {
			ui.ShowSuccess("✅ Device is ready")
			return
		}

		time.Sleep(2 * time.Second)
		fmt.Print(".")
	}

	ui.ShowWarning("Device took too long to be ready")
}

// parseSimulatorList parses iOS simulator list output
func (dm *DeviceManager) parseSimulatorList(output string) map[string][]struct{ name, id string } {
	simulators := make(map[string][]struct{ name, id string })

	lines := strings.Split(output, "\n")
	currentRuntime := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for runtime header
		if strings.HasPrefix(line, "-- ") && strings.HasSuffix(line, " --") {
			currentRuntime = strings.TrimSuffix(strings.TrimPrefix(line, "-- "), " --")
			continue
		}

		// Parse device line
		if currentRuntime != "" && strings.Contains(line, "(") && strings.Contains(line, ")") {
			// Extract device name and ID
			nameEnd := strings.LastIndex(line, "(")
			if nameEnd > 0 {
				name := strings.TrimSpace(line[:nameEnd])

				idStart := strings.LastIndex(line, "(")
				idEnd := strings.LastIndex(line, ")")
				if idStart >= 0 && idEnd > idStart {
					id := strings.TrimSpace(line[idStart+1 : idEnd])

					simulators[currentRuntime] = append(simulators[currentRuntime],
						struct{ name, id string }{name, id})
				}
			}
		}
	}

	return simulators
}

// findAPKFiles finds APK files in the build directory
func (dm *DeviceManager) findAPKFiles() []string {
	apkFiles := []string{}

	// Common APK locations
	locations := []string{
		"build/app/outputs/flutter-apk/*.apk",
		"build/app/outputs/apk/**/*.apk",
		"*.apk",
	}

	for _, pattern := range locations {
		matches, _ := filepath.Glob(pattern)
		apkFiles = append(apkFiles, matches...)
	}

	// Remove duplicates
	seen := make(map[string]bool)
	unique := []string{}
	for _, apk := range apkFiles {
		if !seen[apk] {
			seen[apk] = true
			unique = append(unique, apk)
		}
	}

	return unique
}

// getAndroidPackageName tries to get the Android package name
func (dm *DeviceManager) getAndroidPackageName() string {
	// Try to read from AndroidManifest.xml
	manifestPath := filepath.Join("android", "app", "src", "main", "AndroidManifest.xml")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return ""
	}

	// Simple regex to find package name
	packageRegex := regexp.MustCompile(`package="([^"]+)"`)
	matches := packageRegex.FindStringSubmatch(string(content))
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// openFile opens a file with the default application
func (dm *DeviceManager) openFile(filepath string) error {
	var cmd *exec.Cmd

	osOutput, _ := exec.Command("uname", "-s").Output()
	switch {
	case strings.Contains(string(osOutput), "Darwin"):
		cmd = exec.Command("open", filepath)
	case strings.Contains(string(osOutput), "Linux"):
		cmd = exec.Command("xdg-open", filepath)
	default:
		cmd = exec.Command("cmd", "/c", "start", filepath)
	}

	return cmd.Run()
}
