#!/bin/bash

set -e

VERSION=${1:-"v1.0.0"}
REPO="zxj777/claude-helper"

if [ ! -d "dist" ]; then
    echo "Error: dist directory not found. Run ./build.sh first."
    exit 1
fi

echo "Creating release $VERSION..."

# Check if gh is installed
if ! command -v gh &> /dev/null; then
    echo "GitHub CLI (gh) not found. Please install it first:"
    echo "  brew install gh"
    echo ""
    echo "Or create the release manually at:"
    echo "  https://github.com/$REPO/releases/new"
    echo ""
    echo "Upload these files:"
    ls -la dist/
    exit 1
fi

# Create release with all binaries
gh release create "$VERSION" dist/* \
    --title "$VERSION" \
    --notes "Release $VERSION of claude-helper CLI tool

## Installation

### One-line install script:
\`\`\`bash
curl -sSL https://raw.githubusercontent.com/$REPO/main/install.sh | bash
\`\`\`

### Manual installation:
1. Download the binary for your platform
2. Move it to a directory in your PATH
3. Make it executable: \`chmod +x claude-helper\`

## Platform Support
- Linux (amd64, arm64)  
- macOS (amd64, arm64)
- Windows (amd64)"

echo "Release created successfully!"
echo "View at: https://github.com/$REPO/releases/tag/$VERSION"