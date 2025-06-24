package fluttermanager

import (
	"archive/tar"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kkz6/devtools/internal/config"
	"github.com/kkz6/devtools/internal/ui"
)

// BackupManager handles backup and restore of signing configurations
type BackupManager struct {
	cfg *config.Config
}

// NewBackupManager creates a new backup manager
func NewBackupManager(cfg *config.Config) *BackupManager {
	return &BackupManager{cfg: cfg}
}

// CreateBackup creates a backup of signing configuration
func (bm *BackupManager) CreateBackup() error {
	ui.ShowInfo("💾 Creating backup of signing configuration")
	fmt.Println()

	// Check if signing is configured
	signingMgr := NewSigningManager(bm.cfg)
	status := signingMgr.GetStatus()

	if !status.IsConfigured {
		ui.ShowWarning("No signing configuration found to backup")
		return nil
	}

	// Get backup name
	backupName, err := ui.GetInput(
		"Backup name",
		fmt.Sprintf("flutter-signing-backup-%s", time.Now().Format("20060102")),
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Create backup directory
	backupDir := filepath.Join(".flutter", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create temporary directory for files
	tempDir := filepath.Join(backupDir, "temp-"+time.Now().Format("20060102150405"))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Collect files to backup
	filesToBackup := []string{}

	// Key properties
	keyPropertiesPath := filepath.Join("android", "key.properties")
	if _, err := os.Stat(keyPropertiesPath); err == nil {
		destPath := filepath.Join(tempDir, "key.properties")
		if err := copyFile(keyPropertiesPath, destPath); err == nil {
			filesToBackup = append(filesToBackup, "key.properties")
		}
	}

	// Keystore file
	if status.KeystorePath != "" {
		if _, err := os.Stat(status.KeystorePath); err == nil {
			destPath := filepath.Join(tempDir, filepath.Base(status.KeystorePath))
			if err := copyFile(status.KeystorePath, destPath); err == nil {
				filesToBackup = append(filesToBackup, filepath.Base(status.KeystorePath))
			}
		}
	}

	// Credentials file if exists
	credentialsPath := filepath.Join(".flutter", "signing-credentials.txt")
	if _, err := os.Stat(credentialsPath); err == nil {
		destPath := filepath.Join(tempDir, "signing-credentials.txt")
		if err := copyFile(credentialsPath, destPath); err == nil {
			filesToBackup = append(filesToBackup, "signing-credentials.txt")
		}
	}

	// Create metadata
	metadata := fmt.Sprintf(`Flutter Signing Backup
=====================
Created: %s
Files: %s
Version: 1.0
`, time.Now().Format("2006-01-02 15:04:05"), strings.Join(filesToBackup, ", "))

	metadataPath := filepath.Join(tempDir, "backup-metadata.txt")
	if err := os.WriteFile(metadataPath, []byte(metadata), 0644); err == nil {
		filesToBackup = append(filesToBackup, "backup-metadata.txt")
	}

	// Ask for encryption
	var password string
	if ui.GetConfirmation("Encrypt backup with password?") {
		password, err = ui.GetInput(
			"Enter encryption password",
			"",
			true,
			func(s string) error {
				if len(s) < 8 {
					return fmt.Errorf("password must be at least 8 characters")
				}
				return nil
			},
		)
		if err != nil {
			return err
		}

		// Confirm password
		confirm, err := ui.GetInput("Confirm password", "", true, nil)
		if err != nil {
			return err
		}

		if password != confirm {
			return fmt.Errorf("passwords do not match")
		}
	}

	// Create archive
	archivePath := filepath.Join(backupDir, backupName+".tar.gz")

	err = ui.ShowLoadingAnimation("Creating backup archive", func() error {
		if err := createTarGz(tempDir, archivePath); err != nil {
			return err
		}

		// Encrypt if password provided
		if password != "" {
			encryptedPath := archivePath + ".enc"
			if err := encryptFile(archivePath, encryptedPath, password); err != nil {
				return err
			}

			// Remove unencrypted archive
			os.Remove(archivePath)
			archivePath = encryptedPath
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Get file size
	if info, err := os.Stat(archivePath); err == nil {
		sizeMB := float64(info.Size()) / (1024 * 1024)
		ui.ShowSuccess(fmt.Sprintf("✅ Backup created: %s (%.2f MB)", filepath.Base(archivePath), sizeMB))
	}

	// Create password hint file if encrypted
	if password != "" {
		hintPath := archivePath + ".hint"
		hint := fmt.Sprintf("Backup encrypted on %s\nRemember your password!\n",
			time.Now().Format("2006-01-02"))
		os.WriteFile(hintPath, []byte(hint), 0644)

		ui.ShowWarning("⚠️  Remember your password! It cannot be recovered.")
	}

	return nil
}

// RestoreBackup restores a backup
func (bm *BackupManager) RestoreBackup() error {
	ui.ShowInfo("📥 Restore signing configuration from backup")
	fmt.Println()

	// List available backups
	backups, err := bm.listBackupFiles()
	if err != nil || len(backups) == 0 {
		ui.ShowWarning("No backups found")
		return nil
	}

	// Select backup
	backupNames := make([]string, len(backups))
	for i, backup := range backups {
		backupNames[i] = fmt.Sprintf("%s (%s)", backup.Name, backup.CreatedAt.Format("2006-01-02 15:04"))
	}

	choice, err := ui.SelectFromList("Select backup to restore:", backupNames)
	if err != nil {
		return nil
	}

	selectedBackup := backups[choice]

	// Check if encrypted
	var password string
	if strings.HasSuffix(selectedBackup.Name, ".enc") {
		password, err = ui.GetInput("Enter decryption password", "", true, nil)
		if err != nil {
			return err
		}
	}

	// Create temp directory for extraction
	tempDir := filepath.Join(".flutter", "restore-temp-"+time.Now().Format("20060102150405"))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract backup
	err = ui.ShowLoadingAnimation("Extracting backup", func() error {
		backupPath := filepath.Join(".flutter", "backups", selectedBackup.Name)

		// Decrypt if needed
		if password != "" {
			decryptedPath := filepath.Join(tempDir, "decrypted.tar.gz")
			if err := decryptFile(backupPath, decryptedPath, password); err != nil {
				return fmt.Errorf("decryption failed: %w", err)
			}
			backupPath = decryptedPath
		}

		// Extract tar.gz
		return extractTarGz(backupPath, tempDir)
	})

	if err != nil {
		return err
	}

	// Show what will be restored
	fmt.Println("\n📋 Files to restore:")
	files, _ := filepath.Glob(filepath.Join(tempDir, "*"))
	for _, file := range files {
		fmt.Printf("  • %s\n", filepath.Base(file))
	}
	fmt.Println()

	if !ui.GetConfirmation("Proceed with restore?") {
		return nil
	}

	// Backup current configuration if exists
	signingMgr := NewSigningManager(bm.cfg)
	if signingMgr.GetStatus().IsConfigured {
		if ui.GetConfirmation("Backup current configuration before restoring?") {
			bm.CreateBackup()
		}
	}

	// Restore files
	err = ui.ShowLoadingAnimation("Restoring files", func() error {
		// Restore key.properties
		keyPropSrc := filepath.Join(tempDir, "key.properties")
		if _, err := os.Stat(keyPropSrc); err == nil {
			keyPropDest := filepath.Join("android", "key.properties")
			if err := copyFile(keyPropSrc, keyPropDest); err != nil {
				return fmt.Errorf("failed to restore key.properties: %w", err)
			}
		}

		// Restore keystore files
		keystoreFiles, _ := filepath.Glob(filepath.Join(tempDir, "*.jks"))
		keystoreFiles2, _ := filepath.Glob(filepath.Join(tempDir, "*.keystore"))
		keystoreFiles = append(keystoreFiles, keystoreFiles2...)

		for _, ksFile := range keystoreFiles {
			destPath := filepath.Join("android", filepath.Base(ksFile))
			if err := copyFile(ksFile, destPath); err != nil {
				return fmt.Errorf("failed to restore keystore: %w", err)
			}
		}

		// Restore credentials if exists
		credSrc := filepath.Join(tempDir, "signing-credentials.txt")
		if _, err := os.Stat(credSrc); err == nil {
			credDest := filepath.Join(".flutter", "signing-credentials.txt")
			os.MkdirAll(filepath.Dir(credDest), 0755)
			copyFile(credSrc, credDest)
		}

		return nil
	})

	if err != nil {
		return err
	}

	ui.ShowSuccess("✅ Backup restored successfully!")
	ui.ShowInfo("Please verify your signing configuration")

	return nil
}

// ListBackups lists all available backups
func (bm *BackupManager) ListBackups() error {
	backups, err := bm.listBackupFiles()
	if err != nil {
		return err
	}

	if len(backups) == 0 {
		ui.ShowWarning("No backups found")
		return nil
	}

	fmt.Println("\n📦 Available Backups:")
	fmt.Println("====================")

	for i, backup := range backups {
		fmt.Printf("\n%d. %s\n", i+1, backup.Name)
		fmt.Printf("   Created: %s\n", backup.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Size: %.2f MB\n", float64(backup.Size)/(1024*1024))

		if strings.HasSuffix(backup.Name, ".enc") {
			fmt.Printf("   Status: 🔒 Encrypted\n")
		} else {
			fmt.Printf("   Status: 🔓 Unencrypted\n")
		}

		if len(backup.Contents) > 0 {
			fmt.Printf("   Contents: %s\n", strings.Join(backup.Contents, ", "))
		}
	}

	fmt.Print("\nPress Enter to continue...")
	fmt.Scanln()

	return nil
}

// VerifyBackup verifies backup integrity
func (bm *BackupManager) VerifyBackup() error {
	backups, err := bm.listBackupFiles()
	if err != nil || len(backups) == 0 {
		ui.ShowWarning("No backups found")
		return nil
	}

	// Select backup
	backupNames := make([]string, len(backups))
	for i, backup := range backups {
		backupNames[i] = backup.Name
	}

	choice, err := ui.SelectFromList("Select backup to verify:", backupNames)
	if err != nil {
		return nil
	}

	selectedBackup := backups[choice]
	backupPath := filepath.Join(".flutter", "backups", selectedBackup.Name)

	// Check if encrypted
	var password string
	if strings.HasSuffix(selectedBackup.Name, ".enc") {
		password, err = ui.GetInput("Enter decryption password", "", true, nil)
		if err != nil {
			return err
		}
	}

	// Create temp directory
	tempDir := filepath.Join(".flutter", "verify-temp-"+time.Now().Format("20060102150405"))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Verify backup
	err = ui.ShowLoadingAnimation("Verifying backup", func() error {
		// Decrypt if needed
		verifyPath := backupPath
		if password != "" {
			decryptedPath := filepath.Join(tempDir, "decrypted.tar.gz")
			if err := decryptFile(backupPath, decryptedPath, password); err != nil {
				return fmt.Errorf("decryption failed - invalid password or corrupted file")
			}
			verifyPath = decryptedPath
		}

		// Try to extract
		if err := extractTarGz(verifyPath, tempDir); err != nil {
			return fmt.Errorf("extraction failed - backup may be corrupted")
		}

		// Check for expected files
		expectedFiles := []string{"key.properties"}
		for _, expected := range expectedFiles {
			if _, err := os.Stat(filepath.Join(tempDir, expected)); os.IsNotExist(err) {
				return fmt.Errorf("missing expected file: %s", expected)
			}
		}

		return nil
	})

	if err != nil {
		ui.ShowError(fmt.Sprintf("Backup verification failed: %v", err))
		return nil
	}

	// List contents
	fmt.Println("\n✅ Backup is valid!")
	fmt.Println("\n📋 Backup contents:")
	files, _ := filepath.Glob(filepath.Join(tempDir, "*"))
	for _, file := range files {
		info, _ := os.Stat(file)
		fmt.Printf("  • %s (%d bytes)\n", filepath.Base(file), info.Size())
	}

	return nil
}

// ExportToCloud exports backup to cloud storage
func (bm *BackupManager) ExportToCloud() error {
	ui.ShowInfo("☁️  Export to Cloud Storage")
	fmt.Println()

	options := []string{
		"Export to Google Drive",
		"Export to Dropbox",
		"Export to OneDrive",
		"Generate shareable link",
		"Back",
	}

	choice, err := ui.SelectFromList("Select export destination:", options)
	if err != nil || choice == 4 {
		return nil
	}

	// This is a placeholder implementation
	ui.ShowInfo("Cloud export requires additional setup")
	ui.ShowInfo("For now, you can manually upload the backup files from:")
	ui.ShowInfo(filepath.Join(".flutter", "backups"))

	return nil
}

// listBackupFiles lists all backup files
func (bm *BackupManager) listBackupFiles() ([]BackupInfo, error) {
	backupDir := filepath.Join(".flutter", "backups")
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return []BackupInfo{}, nil
	}

	files, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, err
	}

	backups := []BackupInfo{}
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".tar.gz") ||
			strings.HasSuffix(file.Name(), ".tar.gz.enc")) {
			info, err := file.Info()
			if err != nil {
				continue
			}

			backup := BackupInfo{
				Name:      file.Name(),
				CreatedAt: info.ModTime(),
				Size:      info.Size(),
			}

			backups = append(backups, backup)
		}
	}

	return backups, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// createTarGz creates a tar.gz archive from a directory
func createTarGz(sourceDir, targetFile string) error {
	file, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the source directory itself
		if path == sourceDir {
			return nil
		}

		// Create header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Update header name to be relative
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write file content if not a directory
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})
}

// extractTarGz extracts a tar.gz archive to a directory
func extractTarGz(sourceFile, targetDir string) error {
	file, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create directory if needed
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}

			// Create file
			file, err := os.Create(targetPath)
			if err != nil {
				return err
			}

			// Copy content
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return err
			}

			file.Close()
		}
	}

	return nil
}

// encryptFile encrypts a file using AES
func encryptFile(inputFile, outputFile, password string) error {
	// Read input file
	plaintext, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	// Derive key from password
	key := sha256.Sum256([]byte(password))

	// Create cipher
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Write to output file
	return os.WriteFile(outputFile, ciphertext, 0600)
}

// decryptFile decrypts a file using AES
func decryptFile(inputFile, outputFile, password string) error {
	// Read encrypted file
	ciphertext, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	// Derive key from password
	key := sha256.Sum256([]byte(password))

	// Create cipher
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	// Write to output file
	return os.WriteFile(outputFile, plaintext, 0600)
}
