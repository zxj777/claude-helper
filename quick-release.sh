#!/bin/bash

# Quick Release Script for claude-helper
# This script builds and releases in one command
# Usage: 
#   ./quick-release.sh [version]              # Specify exact version
#   ./quick-release.sh patch                  # Auto-increment patch (bug fix)
#   ./quick-release.sh minor                  # Auto-increment minor (new feature)
#   ./quick-release.sh major                  # Auto-increment major (breaking change)
#   ./quick-release.sh                        # Default to patch

set -e

# Configuration
REPO="zxj777/claude-helper"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')] $1${NC}"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Get version from argument or auto-increment
VERSION_ARG=${1:-"patch"}

# Function to get latest version
get_latest_version() {
    git tag -l "v*.*.*" | sort -V | tail -n1
}

# Function to increment version based on type
increment_version() {
    local current_version="$1"
    local increment_type="$2"
    
    if [[ -z "$current_version" ]]; then
        echo "v0.1.0"
        return
    fi
    
    if [[ "$current_version" =~ v([0-9]+)\.([0-9]+)\.([0-9]+) ]]; then
        local major="${BASH_REMATCH[1]}"
        local minor="${BASH_REMATCH[2]}"
        local patch="${BASH_REMATCH[3]}"
        
        case "$increment_type" in
            "major")
                echo "v$((major + 1)).0.0"
                ;;
            "minor")
                echo "v${major}.$((minor + 1)).0"
                ;;
            "patch"|*)
                echo "v${major}.${minor}.$((patch + 1))"
                ;;
        esac
    else
        echo "v0.1.0"
    fi
}

# Show help if requested
if [[ "$VERSION_ARG" == "--help" || "$VERSION_ARG" == "-h" ]]; then
    echo "Quick Release Script for claude-helper"
    echo ""
    echo "Usage:"
    echo "  $0 [version|type]"
    echo ""
    echo "Version Types (Semantic Versioning):"
    echo "  patch    - Bug fixes, small improvements (x.y.Z)"
    echo "  minor    - New features, backward compatible (x.Y.0)"
    echo "  major    - Breaking changes, major updates (X.0.0)"
    echo ""
    echo "Examples:"
    echo "  $0              # Default patch increment (bug fix)"
    echo "  $0 patch        # Bug fix: v1.2.3 -> v1.2.4"
    echo "  $0 minor        # New feature: v1.2.3 -> v1.3.0"
    echo "  $0 major        # Breaking change: v1.2.3 -> v2.0.0"
    echo "  $0 v1.5.0       # Specific version"
    echo ""
    echo "When to use each type:"
    echo "  patch: Bug fixes, typos, small improvements"
    echo "  minor: New features, new commands, enhancements"
    echo "  major: Breaking API changes, major rewrites"
    exit 0
fi

# Determine version
case "$VERSION_ARG" in
    "patch"|"minor"|"major")
        LATEST_VERSION=$(get_latest_version)
        VERSION=$(increment_version "$LATEST_VERSION" "$VERSION_ARG")
        log "Incrementing $VERSION_ARG version: $LATEST_VERSION -> $VERSION"
        ;;
    v*.*.*)
        VERSION="$VERSION_ARG"
        log "Using specified version: $VERSION"
        ;;
    *.*.*)
        VERSION="v$VERSION_ARG"
        log "Using specified version: $VERSION"
        ;;
    "")
        # Default to patch
        LATEST_VERSION=$(get_latest_version)
        VERSION=$(increment_version "$LATEST_VERSION" "patch")
        log "Default patch increment: $LATEST_VERSION -> $VERSION"
        ;;
    *)
        error "Invalid argument: $VERSION_ARG. Use 'patch', 'minor', 'major', or a version number (e.g., v1.2.3)"
        ;;
esac

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    error "Invalid version format. Use vX.Y.Z (e.g., v1.2.3)"
fi

# Check if version already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    error "Version $VERSION already exists"
fi

# Check dependencies
if ! command -v gh >/dev/null 2>&1; then
    error "GitHub CLI (gh) not found. Install with: brew install gh"
fi

log "Starting quick release for version $VERSION..."

# Step 1: Build
log "Building binaries..."
./build.sh "$VERSION"
success "Build completed"

# Step 2: Commit any changes (if needed)
if ! git diff-index --quiet HEAD --; then
    log "Committing changes..."
    git add .
    git commit -m "chore: prepare release $VERSION"
    git push origin main
fi

# Step 3: Create and push tag
log "Creating git tag..."
git tag -a "$VERSION" -m "Release $VERSION"
git push origin "$VERSION"
success "Tag created and pushed"

# Step 4: Create GitHub release
log "Creating GitHub release..."

# Simple changelog - just get commits since last version
LAST_VERSION=$(git tag -l "v*.*.*" | sort -V | tail -n2 | head -n1)
if [[ -z "$LAST_VERSION" ]]; then
    CHANGELOG="- Initial release"
else
    CHANGELOG=$(git log --pretty=format:"- %s" "$LAST_VERSION"..HEAD | head -10)
fi

# Create release notes
RELEASE_NOTES="## Changes
$CHANGELOG

## Installation

### One-line install:
\`\`\`bash
curl -sSL https://raw.githubusercontent.com/$REPO/main/install.sh | bash
\`\`\`

### Manual install:
1. Download the binary for your platform
2. Move it to your PATH
3. Make executable: \`chmod +x claude-helper\`

## Platforms
- Linux (amd64, arm64)
- macOS (amd64, arm64)  
- Windows (amd64)"

gh release create "$VERSION" dist/* \
    --title "$VERSION" \
    --notes "$RELEASE_NOTES"

success "Release $VERSION created successfully!"
echo ""
echo "🎉 View release: https://github.com/$REPO/releases/tag/$VERSION"
echo "📦 Install with: curl -sSL https://raw.githubusercontent.com/$REPO/main/install.sh | bash"