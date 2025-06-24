package fluttermanager

import (
	"time"
)

// BuildType represents the type of Android build
type BuildType int

const (
	BuildAPKDebug BuildType = iota
	BuildAPKRelease
	BuildAppBundle
	BuildSplitAPKs
	BuildCustomFlavor
)

// VersionBumpType represents the type of version bump
type VersionBumpType int

const (
	BumpPatch VersionBumpType = iota
	BumpMinor
	BumpMajor
)

// SigningStatus represents the current signing configuration status
type SigningStatus struct {
	IsConfigured bool
	KeystorePath string
	KeyAlias     string
	ValidUntil   time.Time
}

// Device represents a connected device or emulator
type Device struct {
	ID       string
	Name     string
	Platform string
	IsActive bool
}

// BackupInfo represents information about a backup
type BackupInfo struct {
	ID          string
	Name        string
	CreatedAt   time.Time
	Size        int64
	Description string
	Contents    []string
}

// FlutterProject represents Flutter project information
type FlutterProject struct {
	Name          string
	Version       string
	BuildNumber   string
	Description   string
	AndroidMinSDK int
	IOSMinVersion string
}

// KeystoreInfo represents Android keystore information
type KeystoreInfo struct {
	Path         string
	Alias        string
	Password     string
	KeyPassword  string
	ValidFrom    time.Time
	ValidUntil   time.Time
	Fingerprints map[string]string // SHA1, SHA256, etc.
}

// BuildConfig represents build configuration
type BuildConfig struct {
	Type          BuildType
	Flavor        string
	Target        string
	ExtraArgs     []string
	SigningConfig *KeystoreInfo
}

// VersionInfo represents version information
type VersionInfo struct {
	Version     string
	BuildNumber string
	Timestamp   time.Time
	CommitHash  string
	Tag         string
}
