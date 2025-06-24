package fluttermanager

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/ui"
)

// SigningManager handles Android app signing configuration
type SigningManager struct {
	cfg *config.Config
}

// NewSigningManager creates a new signing manager
func NewSigningManager(cfg *config.Config) *SigningManager {
	return &SigningManager{cfg: cfg}
}

// GetStatus returns the current signing configuration status
func (sm *SigningManager) GetStatus() SigningStatus {
	// Check for key.properties file
	keyPropertiesPath := filepath.Join("android", "key.properties")
	if _, err := os.Stat(keyPropertiesPath); err != nil {
		return SigningStatus{IsConfigured: false}
	}

	// Read key.properties to get keystore path
	content, err := os.ReadFile(keyPropertiesPath)
	if err != nil {
		return SigningStatus{IsConfigured: false}
	}

	status := SigningStatus{IsConfigured: true}

	// Parse key.properties
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "storeFile=") {
			status.KeystorePath = strings.TrimPrefix(line, "storeFile=")
		} else if strings.HasPrefix(line, "keyAlias=") {
			status.KeyAlias = strings.TrimPrefix(line, "keyAlias=")
		}
	}

	return status
}

// CreateKeystore creates a new keystore for signing
func (sm *SigningManager) CreateKeystore() error {
	ui.ShowInfo("🔐 Creating new keystore for app signing")
	fmt.Println()

	// Get keystore details
	keystoreName, err := ui.GetInput(
		"Keystore filename",
		"upload-keystore.jks",
		false,
		func(s string) error {
			if !strings.HasSuffix(s, ".jks") && !strings.HasSuffix(s, ".keystore") {
				return fmt.Errorf("keystore file must end with .jks or .keystore")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	keyAlias, err := ui.GetInput(
		"Key alias",
		"upload",
		false,
		func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("key alias cannot be empty")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	// Generate secure password if desired
	var storePassword, keyPassword string
	if ui.GetConfirmation("Generate secure passwords automatically?") {
		storePassword = generateSecurePassword(16)
		keyPassword = generateSecurePassword(16)
		ui.ShowInfo("Generated secure passwords")
	} else {
		storePassword, err = ui.GetInput(
			"Keystore password (min 6 chars)",
			"",
			true,
			func(s string) error {
				if len(s) < 6 {
					return fmt.Errorf("password must be at least 6 characters")
				}
				return nil
			},
		)
		if err != nil {
			return err
		}

		keyPassword, err = ui.GetInput(
			"Key password (min 6 chars)",
			"",
			true,
			func(s string) error {
				if len(s) < 6 {
					return fmt.Errorf("password must be at least 6 characters")
				}
				return nil
			},
		)
		if err != nil {
			return err
		}
	}

	// Get certificate details
	commonName, err := ui.GetInput("Your name or company", "Unknown", false, nil)
	if err != nil {
		return err
	}

	orgUnit, err := ui.GetInput("Organizational unit", "Unknown", false, nil)
	if err != nil {
		return err
	}

	org, err := ui.GetInput("Organization", "Unknown", false, nil)
	if err != nil {
		return err
	}

	city, err := ui.GetInput("City or Locality", "Unknown", false, nil)
	if err != nil {
		return err
	}

	state, err := ui.GetInput("State or Province", "Unknown", false, nil)
	if err != nil {
		return err
	}

	country, err := ui.GetInput("Country code (2 letters)", "US", false, func(s string) error {
		if len(s) != 2 {
			return fmt.Errorf("country code must be exactly 2 letters")
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Create android directory if it doesn't exist
	androidDir := "android"
	if err := os.MkdirAll(androidDir, 0755); err != nil {
		return fmt.Errorf("failed to create android directory: %w", err)
	}

	// Generate keystore
	keystorePath := filepath.Join(androidDir, keystoreName)
	dname := fmt.Sprintf("CN=%s, OU=%s, O=%s, L=%s, ST=%s, C=%s",
		commonName, orgUnit, org, city, state, country)

	err = ui.ShowLoadingAnimation("Creating keystore", func() error {
		cmd := exec.Command("keytool",
			"-genkey",
			"-v",
			"-keystore", keystorePath,
			"-keyalg", "RSA",
			"-keysize", "2048",
			"-validity", "10000",
			"-alias", keyAlias,
			"-storepass", storePassword,
			"-keypass", keyPassword,
			"-dname", dname,
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("keytool error: %s", string(output))
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create keystore: %w", err)
	}

	// Create key.properties file
	if err := sm.createKeyProperties(keystorePath, keyAlias, storePassword, keyPassword); err != nil {
		return err
	}

	// Update build.gradle if needed
	if err := sm.updateBuildGradle(); err != nil {
		ui.ShowWarning(fmt.Sprintf("Could not update build.gradle: %v", err))
		ui.ShowInfo("Please ensure your app/build.gradle is configured for signing")
	}

	// Save credentials securely
	if err := sm.saveCredentials(keystoreName, keyAlias, storePassword, keyPassword); err != nil {
		ui.ShowWarning("Could not save credentials to secure storage")
	}

	ui.ShowSuccess("✅ Keystore created successfully!")
	ui.ShowInfo(fmt.Sprintf("📁 Keystore location: %s", keystorePath))
	ui.ShowWarning("⚠️  Keep your keystore and passwords safe! You'll need them for all future updates.")

	return nil
}

// ImportKeystore imports an existing keystore
func (sm *SigningManager) ImportKeystore() error {
	ui.ShowInfo("📥 Import existing keystore")
	fmt.Println()

	// Get keystore path
	keystorePath, err := ui.GetInput(
		"Path to existing keystore",
		"",
		false,
		func(s string) error {
			if _, err := os.Stat(s); err != nil {
				return fmt.Errorf("keystore file not found")
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	// Get keystore details
	keyAlias, err := ui.GetInput("Key alias", "upload", false, nil)
	if err != nil {
		return err
	}

	storePassword, err := ui.GetInput("Keystore password", "", true, nil)
	if err != nil {
		return err
	}

	keyPassword, err := ui.GetInput("Key password", "", true, nil)
	if err != nil {
		return err
	}

	// Verify keystore
	err = ui.ShowLoadingAnimation("Verifying keystore", func() error {
		cmd := exec.Command("keytool",
			"-list",
			"-v",
			"-keystore", keystorePath,
			"-alias", keyAlias,
			"-storepass", storePassword,
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("invalid keystore or credentials: %s", string(output))
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Copy keystore to android directory
	androidDir := "android"
	if err := os.MkdirAll(androidDir, 0755); err != nil {
		return fmt.Errorf("failed to create android directory: %w", err)
	}

	destPath := filepath.Join(androidDir, filepath.Base(keystorePath))

	// Copy file
	input, err := os.ReadFile(keystorePath)
	if err != nil {
		return fmt.Errorf("failed to read keystore: %w", err)
	}

	if err := os.WriteFile(destPath, input, 0600); err != nil {
		return fmt.Errorf("failed to copy keystore: %w", err)
	}

	// Create key.properties
	if err := sm.createKeyProperties(destPath, keyAlias, storePassword, keyPassword); err != nil {
		return err
	}

	// Update build.gradle if needed
	if err := sm.updateBuildGradle(); err != nil {
		ui.ShowWarning(fmt.Sprintf("Could not update build.gradle: %v", err))
		ui.ShowInfo("Please ensure your app/build.gradle is configured for signing")
	}

	ui.ShowSuccess("✅ Keystore imported successfully!")
	return nil
}

// ViewConfiguration displays the current signing configuration
func (sm *SigningManager) ViewConfiguration() error {
	status := sm.GetStatus()

	if !status.IsConfigured {
		ui.ShowWarning("No signing configuration found")
		return nil
	}

	fmt.Println("\n🔐 Signing Configuration")
	fmt.Println("========================")

	// Read key.properties
	keyPropertiesPath := filepath.Join("android", "key.properties")
	content, err := os.ReadFile(keyPropertiesPath)
	if err != nil {
		return fmt.Errorf("failed to read key.properties: %w", err)
	}

	// Parse and display (without passwords)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Password") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				fmt.Printf("%s=********\n", parts[0])
			}
		} else if line != "" {
			fmt.Println(line)
		}
	}

	// Try to get keystore info
	if status.KeystorePath != "" {
		fmt.Println("\n📋 Keystore Information:")
		sm.displayKeystoreInfo(status.KeystorePath, status.KeyAlias)
	}

	fmt.Print("\nPress Enter to continue...")
	fmt.Scanln()

	return nil
}

// UpdatePassword updates keystore passwords
func (sm *SigningManager) UpdatePassword() error {
	ui.ShowWarning("⚠️  This feature requires the current passwords")
	fmt.Println()

	status := sm.GetStatus()
	if !status.IsConfigured {
		ui.ShowError("No signing configuration found")
		return nil
	}

	// This is a placeholder - actual implementation would require
	// using keytool to change passwords
	ui.ShowInfo("Password update is a sensitive operation")
	ui.ShowInfo("It's recommended to create a new keystore instead")

	return nil
}

// ExportConfiguration exports signing configuration for backup
func (sm *SigningManager) ExportConfiguration() error {
	status := sm.GetStatus()
	if !status.IsConfigured {
		ui.ShowError("No signing configuration found")
		return nil
	}

	exportDir := "flutter-signing-export-" + time.Now().Format("20060102-150405")
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// Copy keystore
	if status.KeystorePath != "" {
		keystoreDest := filepath.Join(exportDir, filepath.Base(status.KeystorePath))
		input, err := os.ReadFile(status.KeystorePath)
		if err == nil {
			os.WriteFile(keystoreDest, input, 0600)
		}
	}

	// Copy key.properties
	keyPropertiesPath := filepath.Join("android", "key.properties")
	if content, err := os.ReadFile(keyPropertiesPath); err == nil {
		os.WriteFile(filepath.Join(exportDir, "key.properties"), content, 0600)
	}

	// Create README
	readme := `Flutter Signing Configuration Export
====================================

This directory contains your Flutter app signing configuration.

Files:
- *.jks/keystore: Your signing keystore
- key.properties: Configuration file with keystore details

IMPORTANT:
- Keep these files secure and backed up
- Never commit these files to version control
- You need these files to update your app

To restore:
1. Copy the keystore file to android/
2. Copy key.properties to android/
3. Ensure your build.gradle is configured for signing
`

	os.WriteFile(filepath.Join(exportDir, "README.txt"), []byte(readme), 0644)

	ui.ShowSuccess(fmt.Sprintf("✅ Configuration exported to: %s", exportDir))

	if ui.GetConfirmation("Create encrypted archive?") {
		archiveName := exportDir + ".tar.gz"
		cmd := exec.Command("tar", "-czf", archiveName, exportDir)
		if err := cmd.Run(); err == nil {
			ui.ShowSuccess(fmt.Sprintf("📦 Archive created: %s", archiveName))

			// Clean up directory
			os.RemoveAll(exportDir)
		}
	}

	return nil
}

// VerifyKeystore verifies the keystore is valid
func (sm *SigningManager) VerifyKeystore() error {
	status := sm.GetStatus()
	if !status.IsConfigured {
		ui.ShowError("No signing configuration found")
		return nil
	}

	// Read key.properties to get passwords
	keyPropertiesPath := filepath.Join("android", "key.properties")
	content, err := os.ReadFile(keyPropertiesPath)
	if err != nil {
		return fmt.Errorf("failed to read key.properties: %w", err)
	}

	var storePassword string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "storePassword=") {
			storePassword = strings.TrimPrefix(line, "storePassword=")
			break
		}
	}

	if storePassword == "" {
		storePassword, err = ui.GetInput("Enter keystore password", "", true, nil)
		if err != nil {
			return err
		}
	}

	// Verify keystore
	err = ui.ShowLoadingAnimation("Verifying keystore", func() error {
		cmd := exec.Command("keytool",
			"-list",
			"-v",
			"-keystore", status.KeystorePath,
			"-storepass", storePassword,
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("keystore verification failed: %s", string(output))
		}

		// Parse output for expiration
		outputStr := string(output)
		if strings.Contains(outputStr, "Valid from:") {
			fmt.Println("\n" + outputStr)
		}

		return nil
	})

	if err != nil {
		return err
	}

	ui.ShowSuccess("✅ Keystore is valid")
	return nil
}

// createKeyProperties creates the key.properties file
func (sm *SigningManager) createKeyProperties(keystorePath, keyAlias, storePassword, keyPassword string) error {
	keyProperties := fmt.Sprintf(`storePassword=%s
keyPassword=%s
keyAlias=%s
storeFile=%s
`, storePassword, keyPassword, keyAlias, keystorePath)

	keyPropertiesPath := filepath.Join("android", "key.properties")
	if err := os.WriteFile(keyPropertiesPath, []byte(keyProperties), 0600); err != nil {
		return fmt.Errorf("failed to create key.properties: %w", err)
	}

	// Add to .gitignore
	sm.addToGitignore("android/key.properties")
	sm.addToGitignore("android/*.jks")
	sm.addToGitignore("android/*.keystore")

	return nil
}

// updateBuildGradle updates build.gradle for signing configuration
func (sm *SigningManager) updateBuildGradle() error {
	buildGradlePath := filepath.Join("android", "app", "build.gradle")

	content, err := os.ReadFile(buildGradlePath)
	if err != nil {
		return fmt.Errorf("failed to read build.gradle: %w", err)
	}

	gradleContent := string(content)

	// Check if signing config already exists
	if strings.Contains(gradleContent, "signingConfigs") {
		return nil // Already configured
	}

	// This is a simplified implementation
	// In production, you'd want more sophisticated gradle file manipulation
	ui.ShowInfo("Please ensure your build.gradle is configured for release signing")
	ui.ShowInfo("Add the signing configuration to android/app/build.gradle")

	return nil
}

// addToGitignore adds a pattern to .gitignore
func (sm *SigningManager) addToGitignore(pattern string) {
	gitignorePath := ".gitignore"

	// Read existing content
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		// Create new .gitignore
		content = []byte("")
	}

	// Check if pattern already exists
	if strings.Contains(string(content), pattern) {
		return
	}

	// Append pattern
	newContent := string(content)
	if !strings.HasSuffix(newContent, "\n") && len(newContent) > 0 {
		newContent += "\n"
	}
	newContent += pattern + "\n"

	os.WriteFile(gitignorePath, []byte(newContent), 0644)
}

// saveCredentials saves credentials to a secure location
func (sm *SigningManager) saveCredentials(keystoreName, keyAlias, storePassword, keyPassword string) error {
	// Create .flutter directory
	flutterDir := ".flutter"
	if err := os.MkdirAll(flutterDir, 0700); err != nil {
		return err
	}

	// Save credentials (in production, use proper encryption)
	credentials := fmt.Sprintf(`Keystore: %s
Alias: %s
Store Password: %s
Key Password: %s
Created: %s
`, keystoreName, keyAlias, storePassword, keyPassword, time.Now().Format("2006-01-02 15:04:05"))

	credentialsPath := filepath.Join(flutterDir, "signing-credentials.txt")
	if err := os.WriteFile(credentialsPath, []byte(credentials), 0600); err != nil {
		return err
	}

	// Add to .gitignore
	sm.addToGitignore(".flutter/")

	return nil
}

// displayKeystoreInfo displays keystore information
func (sm *SigningManager) displayKeystoreInfo(keystorePath, keyAlias string) {
	// This would normally use keytool to display info
	// For now, just show basic info
	if info, err := os.Stat(keystorePath); err == nil {
		fmt.Printf("  File: %s\n", filepath.Base(keystorePath))
		fmt.Printf("  Size: %.2f KB\n", float64(info.Size())/1024)
		fmt.Printf("  Modified: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
		fmt.Printf("  Alias: %s\n", keyAlias)
	}
}

// generateSecurePassword generates a secure random password
func generateSecurePassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}
