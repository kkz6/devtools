# DevTools Manager

A modular Go application for managing various development tools and configurations. Currently supports Git commit signing setup with SSH and GPG keys.

## Features

- **Git Signing Configuration**: Set up SSH or GPG signing for Git commits

  - Generate new SSH/GPG keys
  - Import existing SSH keys
  - Configure Git to use signing keys
  - Upload SSH signing keys to GitHub
  - Export GPG public keys

- **Modular Architecture**: Easy to extend with new tools
- **Configuration Management**: All settings stored in YAML file
- **Interactive CLI**: User-friendly menu system

## Installation

```bash
# Clone the repository
git clone https://github.com/kkz6/devtools.git
cd devtools

# Build the application
go build -o devtools

# Run the application
./devtools
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

1. Run the application:

   ```bash
   ./devtools
   ```

2. Select "Git Commit Signing Setup" from the menu

3. Choose your preferred signing method:

   - SSH signing (recommended)
   - GPG signing
   - Export existing SSH key to GitHub
   - Import existing SSH signing key

4. Follow the interactive prompts

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

- Go 1.21 or later
- Git
- SSH (for SSH signing)
- GPG (for GPG signing, will be installed automatically on macOS)

## License

MIT License
