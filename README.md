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

### 📊 Cursor AI Usage Reporter

- Track API usage and costs
- Compare Free vs Pro vs Business plans
- View 30-day usage history with charts
- Get cost-saving recommendations
- Export usage data

### ⚙️ Configuration Manager

- Manage all settings from one place
- Secure credential storage
- Easy configuration updates

### 🎨 Beautiful UI

- Interactive menus with keyboard navigation
- Animated banners and loading indicators
- Color-coded messages and status updates
- Progress bars and spinners

## Installation

### Quick Install (macOS)

The easiest way to install DevTools on macOS is using our install script:

```bash
# System-wide installation (may require sudo)
curl -fsSL https://raw.githubusercontent.com/kkz6/devtools/main/install.sh | bash

# User-only installation (no sudo required)
curl -fsSL https://raw.githubusercontent.com/kkz6/devtools/main/install.sh | bash -s -- --user
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

settings:
  preferred_signing_method: "ssh"
```

## Usage

Simply run:

```bash
devtools
```

You'll be presented with an interactive menu where you can:

1. **Git Commit Signing Setup** - Configure SSH/GPG signing for secure commits
2. **Configuration Manager** - Manage your DevTools settings
3. **Cursor AI Usage Report** - Track your AI assistant usage and costs

### Quick Examples

#### Set up Git SSH Signing

```bash
devtools
# Select "Git Commit Signing Setup"
# Choose "SSH signing setup (recommended)"
# Follow the prompts
```

#### Check Cursor AI Usage

```bash
devtools
# Select "Cursor AI Usage Report"
# View your usage statistics and cost analysis
```

#### Clean Up Old Keys

```bash
devtools
# Select "Git Commit Signing Setup"
# Choose "Clean up GPG/SSH signing (Remove all)"
# Select what to remove
```

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

### Build Requirements

- Go 1.23 or later (only if building from source)

## License

MIT License
