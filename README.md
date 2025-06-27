# DevTools

A powerful and modular command-line toolkit for managing development tools, Git signing configurations, and more.

![Build Status](https://github.com/kkz6/devtools/actions/workflows/build.yml/badge.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.23-blue.svg)

## Features

### 🔐 Git Signing Configuration

- **SSH Signing** (Recommended)
  - Generate new SSH signing keys
  - Import existing SSH keys
  - Upload keys to GitHub automatically
  - Configure Git for commit/tag signing
- **GPG Signing**
  - Generate new GPG keys
  - Configure Git for GPG signing
  - Export public keys with clipboard support
- **Cleanup Tools**
  - Remove GPG/SSH keys from system and GitHub
  - Clear Git signing configuration
  - Start fresh with new keys

### 🐛 Issue Manager (Sentry/Linear)

- **Multi-Instance Support**
  - Manage multiple Linear accounts (work, personal, etc.)
  - Manage multiple Sentry instances (cloud, self-hosted)
  - Test connectivity for each instance
- **Manual Issue Creation**
  - Create issues in any Linear instance
  - Full metadata support (type, priority, labels, state)
  - Works independently without Sentry
- **Smart Bug Sync**
  - Connect any Sentry instance to any Linear instance
  - Map Sentry projects to Linear teams/projects
  - Batch sync with automatic priority assignment
  - Option to resolve issues in Sentry after sync

### 📊 Cursor AI Usage Reporter

- Track API usage and costs
- Compare Free vs Pro vs Business plans
- View 30-day usage history with charts
- Get cost-saving recommendations
- Export usage data

### 🔧 Configuration Manager

- Manage GitHub credentials and tokens
- Configure SSH and GPG signing keys
- Set up Cursor AI API integration
- Configure Sentry API settings with custom base URLs (for different regions or self-hosted instances)
- Configure Linear API credentials
- Store and manage global settings

### 📱 Flutter Application Manager

- Build and manage Android apps (APK/Bundle)
- Version and build number management
- Android signing configuration and backup
- Device and emulator management
- Project setup and dependency management

### 🎨 Beautiful UI

- Interactive menus with keyboard navigation
- Animated banners and loading indicators
- Color-coded messages and status updates
- Progress bars and spinners

### 🚀 Release Manager

- Create semantic version releases
- Generate changelogs from commit history
- Create and push Git tags
- Manage GitHub releases with auto-generated notes
- Support for major, minor, and patch releases

### 📱 Flutter Application Manager (Detailed)

- **Android Build Management**
  - Build APK files (Debug/Release)
  - Build App Bundles for Play Store
  - Split APKs by architecture
  - Custom flavor builds
  - Automatic signing configuration
- **Version Management**
  - Semantic versioning support
  - Auto-increment version and build numbers
  - Git tag creation and management
  - Version history tracking
  - CI/CD friendly version bumping
- **Signing Configuration**
  - Create new Android keystores
  - Import existing keystores
  - Secure password generation
  - Keystore verification
  - Export signing configurations
- **Backup & Restore**
  - Encrypted backup archives
  - Password-protected backups
  - Easy restoration process
  - Cloud export support
  - Backup integrity verification
- **Project Setup & Dependencies**
  - Flutter environment checks
  - Dependency management (pub get/upgrade)
  - Android SDK configuration
  - Firebase integration setup
  - Build flavor configuration
  - App icon generation
- **Clean & Rebuild**
  - Flutter clean operations
  - Build cache management
  - iOS Pod reset
  - Full project reset with restoration
  - Selective cleaning options
- **Device Management**
  - List connected devices
  - Launch Android emulators
  - Launch iOS simulators
  - Install APKs to devices
  - Stream device logs
  - Take device screenshots

### Issue Manager (Sentry/Linear)

The Issue Manager now supports multiple instances of both Linear and Sentry, allowing you to:

1. **Manage Multiple Instances**:

   - Add multiple Linear API configurations (e.g., work, personal)
   - Add multiple Sentry API configurations with different base URLs
   - Test connectivity for each instance

2. **Create Manual Issues**:

   - Select any configured Linear instance
   - Create issues with full metadata (type, priority, labels, state)
   - No Sentry configuration required for manual issue creation

3. **Sync Bugs Between Instances**:

   - Create connections between specific Linear and Sentry instances
   - Map Sentry projects to Linear teams/projects
   - Configure default labels for synced issues
   - Batch sync multiple issues with option to resolve in Sentry

4. **Configuration Example**:

```yaml
# Multiple Linear instances
linear:
  instances:
    work:
      name: "Work Linear"
      api_key: lin_api_work_key
    personal:
      name: "Personal Linear"
      api_key: lin_api_personal_key

# Multiple Sentry instances
sentry:
  instances:
    work:
      name: "Work Sentry"
      api_key: work_sentry_key
      base_url: https://sentry.io/api/0
    self_hosted:
      name: "Self-Hosted Sentry"
      api_key: self_hosted_key
      base_url: https://sentry.mycompany.com/api/0

# Connections between instances
bug_manager:
  connections:
    - name: "Work Projects"
      sentry_instance: work
      linear_instance: work
      project_mappings:
        - sentry_organization: my-org
          sentry_project: backend
          linear_team_id: team-uuid
          linear_project_id: project-uuid
          default_labels: [bug, sentry, backend]
```

## Installation

### Quick Install (macOS)

The easiest way to install DevTools on macOS is using our install script:

```bash
# System-wide installation (may require sudo)
bash -c "$(curl -fsSL https://raw.githubusercontent.com/kkz6/devtools/main/install.sh)"

# User-only installation (no sudo required)
bash -c "$(curl -fsSL https://raw.githubusercontent.com/kkz6/devtools/main/install.sh)" -- --user

# Alternative method (download and run)
curl -fsSL https://raw.githubusercontent.com/kkz6/devtools/main/install.sh -o install.sh
chmod +x install.sh
./install.sh
```

The install script will:

- Download the latest release
- Ask for sudo permission only if needed
- Install to `/usr/local/bin` (system) or `~/.local/bin` (user)
- Create the configuration directory at `~/.devtools`
- Provide PATH setup instructions if needed

### Manual Installation

#### From Release (Recommended)

1. Download the latest release for your platform from [GitHub Releases](https://github.com/kkz6/devtools/releases)
2. Extract and move to your PATH:

   ```bash
   # macOS/Linux
   chmod +x devtools-darwin-amd64
   sudo mv devtools-darwin-amd64 /usr/local/bin/devtools

   # Verify installation
   devtools --version
   ```

#### Build from Source

```bash
# Clone the repository
git clone https://github.com/kkz6/devtools.git
cd devtools

# Build the application
go build -o devtools

# Install to system (optional)
sudo mv devtools /usr/local/bin/

# Run the application
devtools
```

### Homebrew (Coming Soon)

```bash
brew tap kkz6/devtools
brew install devtools
```

## Configuration

The application stores its configuration in `~/.devtools/config.yaml`. A default configuration file is created on first run.

### Sample Configuration

```yaml
github:
  username: "your-github-username"
  token: "your-github-personal-access-token"
  email: "your-email@example.com"

ssh:
  signing_key_path: "/Users/username/.ssh/git-ssh-signing-key"
  key_comment: "git-ssh-signing-key"

gpg:
  key_id: ""
  email: "your-email@example.com"

flutter:
  android_sdk_path: "" # Optional, auto-detected if ANDROID_HOME is set
  keystore_dir: "~/.devtools/flutter/keystores"
  backup_dir: "~/.devtools/flutter/backups"
  default_build_mode: "release"
  projects: {} # Project-specific configurations

settings:
  preferred_signing_method: "ssh"
```

## Usage

Run the tool with:

```bash
devtools
```

This will present an interactive menu where you can select the desired functionality.

### Configuration Manager

Manage all your development tool configurations in one place:

- GitHub credentials for API access
- SSH keys for Git signing
- GPG keys for commit verification
- Cursor AI API settings

### Git Signing

Set up commit signing for enhanced security:

1. Choose between SSH or GPG signing
2. Generate or import signing keys
3. Configure Git to use the signing method
4. Upload public keys to GitHub

### Cursor AI Reporter

Monitor your AI assistant usage:

1. Configure your Cursor API key
2. Select a date range for the report
3. View usage statistics and costs
4. Export data for further analysis

### Release Manager

Create and manage releases:

1. Select release type (major/minor/patch)
2. Review and edit changelog
3. Create Git tag and GitHub release
4. Push changes to remote repository

### Bug Manager

Sync bugs from Sentry to Linear:

1. **Initial Setup**:

   - Configure Sentry API key (get from Sentry Settings → API Keys)
   - Configure Linear API key (get from Linear Settings → API → Personal API keys)

2. **Project Configuration**:

   - Map Sentry projects to Linear projects
   - Set default labels for imported bugs
   - Configure team and project associations

3. **Bug Syncing**:
   - Select a configured project mapping
   - View up to 5 recent unresolved bugs from Sentry
   - Review bug details including:
     - Error level and platform
     - Number of occurrences and affected users
     - Error message and stack trace information
   - Confirm and create the issue in Linear with:
     - Automatic priority based on severity
     - Relevant labels (bug, sentry, level, platform, impact)
     - Comprehensive description with Sentry link

### Flutter Application Manager

Manage Flutter app development workflow:

1. **Building Android Apps**:

   - Navigate to your Flutter project directory
   - Select "Build Android (APK/Bundle)"
   - Choose build type (Debug APK, Release APK, App Bundle, Split APKs)
   - Configure signing for release builds
   - Monitor build progress and output location

2. **Version Management**:

   - View current version and build number
   - Bump version (major/minor/patch)
   - Set custom version numbers
   - Auto-increment build numbers
   - Create Git tags for releases

3. **Signing Configuration**:

   - Create new keystore with secure passwords
   - Import existing keystores
   - Export signing configuration for backup
   - Verify keystore validity

4. **Backup & Restore**:

   - Create encrypted backups of signing configuration
   - Restore from previous backups
   - Password-protect sensitive data
   - Export to cloud storage

5. **Device Management**:
   - Launch Android emulators or iOS simulators
   - Install APKs to connected devices
   - Stream device logs for debugging
   - Take screenshots from devices

## GitHub Personal Access Token

To upload SSH keys to GitHub, you'll need a personal access token with the `write:ssh_signing_key` scope:

1. Go to https://github.com/settings/tokens
2. Click "Generate new token"
3. Select the `write:ssh_signing_key` scope
4. Generate and copy the token
5. Add it to your config file or enter it when prompted

## Adding New Modules

To add a new tool module:

1. Create a new package in `internal/modules/yourmodule/`
2. Implement the `Module` interface:
   ```go
   type Module interface {
       Execute(cfg *config.Config) error
       Info() ModuleInfo
   }
   ```
3. Register your module in `internal/modules/register.go`

## Project Structure

```
devtools/
├── main.go                           # Entry point
├── go.mod                           # Go module file
├── internal/
│   ├── config/                      # Configuration management
│   │   └── config.go
│   ├── menu/                        # Interactive menu system
│   │   └── menu.go
│   └── modules/                     # Tool modules
│       ├── registry.go              # Module registry
│       ├── register.go              # Module registration
│       └── gitsigning/              # Git signing module
│           ├── module.go            # Main module logic
│           ├── ssh.go               # SSH signing implementation
│           └── gpg.go               # GPG signing implementation
└── README.md
```

## Requirements

- **macOS**: 10.15 (Catalina) or later
- **Linux**: Ubuntu 20.04+ or equivalent
- **Windows**: Windows 10+ (WSL recommended)

### Runtime Dependencies

- Git 2.34+ (for SSH signing support)
- OpenSSH (included in macOS/Linux)
- GPG (optional, auto-installed on macOS if needed)

#### For Flutter Application Manager:

- Flutter SDK 3.0+ (with Dart)
- Android SDK (for Android builds)
- Java 11+ (for Android builds)
- Xcode (for iOS features on macOS)
- CocoaPods (for iOS dependencies)

### Build Requirements

- Go 1.23 or later (only if building from source)

## Author

**Karthick**  
Email: [karthick@gigcodes.com](mailto:karthick@gigcodes.com)  
Website: [devkarti.com](https://devkarti.com)

## License

MIT License

Copyright (c) 2024 Karthick

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
