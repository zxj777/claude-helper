#!/bin/bash

# Build script for cross-platform binaries
set -e

VERSION=${1:-"latest"}
BINARY_NAME="claude-helper"

echo "Building version: $VERSION"

# Clean previous builds
rm -rf dist
mkdir -p dist

# Build for different platforms
echo "Building for Linux amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o dist/${BINARY_NAME}-linux-amd64 ./cmd/claude-helper

echo "Building for Linux arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=$VERSION" -o dist/${BINARY_NAME}-linux-arm64 ./cmd/claude-helper

echo "Building for macOS amd64..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o dist/${BINARY_NAME}-darwin-amd64 ./cmd/claude-helper

echo "Building for macOS arm64..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=$VERSION" -o dist/${BINARY_NAME}-darwin-arm64 ./cmd/claude-helper

echo "Building for Windows amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o dist/${BINARY_NAME}-windows-amd64.exe ./cmd/claude-helper

echo "Build complete! Files in dist/ directory:"
ls -la dist/