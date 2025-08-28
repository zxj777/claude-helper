#!/bin/bash

set -e

# Configuration
REPO_OWNER="zxj777"  # Replace with your GitHub username
REPO_NAME="claude-helper"
BINARY_NAME="claude-helper"
INSTALL_DIR="/usr/local/bin"

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case $OS in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    mingw*|msys*|cygwin*) 
        OS="windows"
        INSTALL_DIR="$HOME/bin"
        # Create install directory if it doesn't exist
        mkdir -p "$INSTALL_DIR"
        # Add to PATH if not already there
        if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
            echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> ~/.bashrc
            echo "Added $INSTALL_DIR to PATH in ~/.bashrc"
            echo "Please run: source ~/.bashrc or restart your terminal"
        fi
        ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Get latest release
echo "Fetching latest release info..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest")
TAG=$(echo "$LATEST_RELEASE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$TAG" ]; then
    echo "Failed to get latest release tag"
    exit 1
fi

echo "Latest version: $TAG"

# Construct download URL
BINARY_FILE="${BINARY_NAME}-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY_FILE="${BINARY_FILE}.exe"
fi

DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$TAG/$BINARY_FILE"

echo "Downloading $DOWNLOAD_URL..."

# Download binary
TEMP_FILE=$(mktemp)
if ! curl -L -o "$TEMP_FILE" "$DOWNLOAD_URL"; then
    echo "Failed to download binary"
    exit 1
fi

# Make executable
chmod +x "$TEMP_FILE"

# Install
if [ "$OS" = "windows" ]; then
    TARGET_FILE="$INSTALL_DIR/$BINARY_NAME.exe"
else
    TARGET_FILE="$INSTALL_DIR/$BINARY_NAME"
fi

echo "Installing to $TARGET_FILE..."
if [ "$OS" = "windows" ] || [ -w "$INSTALL_DIR" ]; then
    mv "$TEMP_FILE" "$TARGET_FILE"
else
    sudo mv "$TEMP_FILE" "$TARGET_FILE"
fi

echo "Installation complete!"
echo "You can now run: $BINARY_NAME"