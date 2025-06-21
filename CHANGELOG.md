# Changelog

All notable changes to DevTools will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.2-alpha] - 2024-12-21

### Added

- Arrow-key navigation for all selection menus
- Clean navigation back to main menu without success messages
- Clipboard support for copying GPG and SSH public keys
- Cleanup functionality for removing GPG/SSH keys from system and GitHub
- Author information displayed prominently in banner
- Improved error handling with `ErrNavigateBack`

### Fixed

- "Back to main menu" now works correctly without exiting the application
- GitHub Actions release permissions issue resolved
- Windows build compatibility in GitHub Actions

### Changed

- All menus now use interactive arrow-key selection instead of number input
- Success messages only shown for completed tasks, not navigation

## [0.0.1] - 2024-12-20

### Added

- Initial release of DevTools
- Git Commit Signing Setup module (SSH and GPG)
- Configuration Manager module
- Cursor AI Usage Report module
- Beautiful TUI with Bubble Tea and Lipgloss
- Modular architecture for easy extension
- GitHub Actions workflow for automated builds and releases
- Installation script for macOS
- Comprehensive documentation and module creation guide

### Features

- SSH key generation and GitHub upload
- GPG key setup and configuration
- Import existing SSH keys
- View and toggle Git signing status
- Manage all configuration settings
- Track Cursor AI usage and costs
- Cost analysis and plan recommendations

---

Created by Karthick

- Email: karthick@gigcodes.com
- Website: https://devkarti.com
