#!/bin/bash

# DevTools Installation Script for macOS
# This script downloads and installs the devtools binary to /usr/local/bin
#
# Usage:
#   bash -c "$(curl -fsSL https://raw.githubusercontent.com/kkz6/devtools/main/install.sh)"
#   bash -c "$(curl -fsSL https://raw.githubusercontent.com/kkz6/devtools/main/install.sh)" -- --user

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="kkz6/devtools"
BINARY_NAME="devtools"

# Default install directory
DEFAULT_INSTALL_DIR="/usr/local/bin"
USER_INSTALL_DIR="$HOME/.local/bin"

# Check for --user flag
USER_INSTALL=false
for arg in "$@"; do
    if [ "$arg" = "--user" ]; then
        USER_INSTALL=true
    fi
done

# Set install directory based on flag
if [ "$USER_INSTALL" = true ]; then
    INSTALL_DIR="$USER_INSTALL_DIR"
else
    INSTALL_DIR="$DEFAULT_INSTALL_DIR"
fi

# Functions
print_error() {
    echo -e "${RED}Error: $1${NC}" >&2
}

print_success() {
    echo -e "${GREEN}$1${NC}"
}

print_info() {
    echo -e "${YELLOW}$1${NC}"
}

# Check for help flag
for arg in "$@"; do
    if [ "$arg" = "--help" ] || [ "$arg" = "-h" ]; then
        echo "DevTools Installation Script"
        echo ""
        echo "Usage: ./install.sh [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --user    Install to user directory (~/.local/bin) instead of system-wide"
        echo "  --help    Show this help message"
        echo ""
        echo "Examples:"
        echo "  ./install.sh           # Install system-wide (may require sudo)"
        echo "  ./install.sh --user    # Install for current user only"
        echo ""
        exit 0
    fi
done

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    print_error "This installation script is for macOS only."
    exit 1
fi

# Check if curl is installed
if ! command -v curl &> /dev/null; then
    print_error "curl is required but not installed. Please install curl first."
    exit 1
fi

# Get the latest release version
print_info "Fetching latest release information..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    print_error "Could not fetch latest release information."
    print_info "You can manually download from: https://github.com/$REPO/releases"
    exit 1
fi

print_info "Latest version: $LATEST_RELEASE"

# Determine architecture
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
else
    print_error "Unsupported architecture: $ARCH"
    exit 1
fi

# Download URL
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/devtools-darwin-$ARCH"

# Create temp directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Download binary
print_info "Downloading devtools..."
if ! curl -L -o "$TEMP_DIR/$BINARY_NAME" "$DOWNLOAD_URL"; then
    print_error "Failed to download devtools."
    print_info "URL attempted: $DOWNLOAD_URL"
    exit 1
fi

# Make binary executable
chmod +x "$TEMP_DIR/$BINARY_NAME"

# Check if /usr/local/bin exists, create if it doesn't
if [ ! -d "$INSTALL_DIR" ]; then
    print_info "Creating $INSTALL_DIR directory..."
    print_info "This requires administrator privileges."
    if ! sudo mkdir -p "$INSTALL_DIR"; then
        print_error "Failed to create $INSTALL_DIR"
        exit 1
    fi
fi

# Check if we need sudo for installation
NEED_SUDO=false
if [ -w "$INSTALL_DIR" ]; then
    print_info "Installing devtools to $INSTALL_DIR..."
    if mv "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"; then
        print_success "✅ devtools has been successfully installed!"
    else
        print_error "Failed to install devtools to $INSTALL_DIR"
        exit 1
    fi
else
    print_info "Installing devtools to $INSTALL_DIR..."
    print_info "This requires administrator privileges."
    print_info ""
    read -p "Do you want to install with sudo? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if sudo mv "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"; then
            print_success "✅ devtools has been successfully installed!"
        else
            print_error "Failed to install devtools to $INSTALL_DIR"
            exit 1
        fi
    else
        print_info ""
        print_info "Installation cancelled. You can manually install by running:"
        print_info "  sudo cp $TEMP_DIR/$BINARY_NAME $INSTALL_DIR/$BINARY_NAME"
        print_info ""
        print_info "Or install to a different location:"
        print_info "  cp $TEMP_DIR/$BINARY_NAME ~/bin/devtools"
        exit 0
    fi
fi

# Verify installation
if command -v devtools &> /dev/null; then
    print_success "Installation verified! You can now run 'devtools' from anywhere."
    print_info ""
    print_info "To get started, run:"
    print_info "  devtools"
else
    print_info ""
    print_info "⚠️  devtools was installed but may not be in your PATH."
    
    # Provide appropriate PATH instructions based on install location
    if [ "$USER_INSTALL" = true ]; then
        print_info "Add $INSTALL_DIR to your PATH by adding this line to your shell profile:"
        print_info ""
        if [ -f "$HOME/.zshrc" ]; then
            print_info "For zsh (~/.zshrc):"
            print_info "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
            print_info "  source ~/.zshrc"
        elif [ -f "$HOME/.bashrc" ]; then
            print_info "For bash (~/.bashrc):"
            print_info "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
            print_info "  source ~/.bashrc"
        else
            print_info "  export PATH=\"$INSTALL_DIR:\$PATH\""
        fi
    else
        print_info "Add $INSTALL_DIR to your PATH by adding this line to your shell profile:"
        print_info "  export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
fi

# Create config directory
CONFIG_DIR="$HOME/.devtools"
if [ ! -d "$CONFIG_DIR" ]; then
    print_info ""
    print_info "Creating configuration directory at $CONFIG_DIR..."
    mkdir -p "$CONFIG_DIR"
    print_success "Configuration directory created."
fi

print_info ""
print_success "🎉 Installation complete!" 