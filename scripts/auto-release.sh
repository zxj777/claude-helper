#!/bin/bash

# Automated Build and Release Script for claude-helper
# Usage: ./auto-release.sh [version] [--dry-run] [--force]
#
# Examples:
#   ./auto-release.sh v1.2.0        # Release version v1.2.0
#   ./auto-release.sh               # Auto-increment patch version
#   ./auto-release.sh --dry-run     # Test run without actually releasing
#   ./auto-release.sh v1.2.0 --force # Force release even with uncommitted changes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="zxj777/claude-helper"
MAIN_BRANCH="main"

# Parse command line arguments
VERSION=""
DRY_RUN=false
FORCE=false

for arg in "$@"; do
    case $arg in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        v*.*.*)
            VERSION="$arg"
            shift
            ;;
        *)
            if [[ "$arg" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
                VERSION="v$arg"
            else
                echo -e "${RED}Error: Unknown argument '$arg'${NC}"
                echo "Usage: $0 [version] [--dry-run] [--force]"
                exit 1
            fi
            shift
            ;;
    esac
done

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir >/dev/null 2>&1; then
        log_error "Not in a git repository"
        exit 1
    fi
}

# Check if we're on the main branch
check_main_branch() {
    current_branch=$(git rev-parse --abbrev-ref HEAD)
    if [[ "$current_branch" != "$MAIN_BRANCH" ]]; then
        log_warn "Not on '$MAIN_BRANCH' branch (currently on '$current_branch')"
        if [[ "$FORCE" != true ]]; then
            log_error "Use --force to release from a non-main branch"
            exit 1
        fi
    fi
}

# Check for uncommitted changes
check_clean_working_tree() {
    if ! git diff-index --quiet HEAD --; then
        log_warn "You have uncommitted changes"
        if [[ "$FORCE" != true ]]; then
            log_error "Please commit your changes before releasing, or use --force"
            git status --porcelain
            exit 1
        fi
    fi
}

# Get the latest tag to auto-increment version
get_latest_version() {
    # Get the latest tag that matches vX.Y.Z format
    latest_tag=$(git tag -l "v*.*.*" | sort -V | tail -n1)
    if [[ -z "$latest_tag" ]]; then
        echo "v0.0.0"
    else
        echo "$latest_tag"
    fi
}

# Auto-increment patch version
auto_increment_version() {
    local current_version="$1"
    if [[ "$current_version" =~ v([0-9]+)\.([0-9]+)\.([0-9]+) ]]; then
        major="${BASH_REMATCH[1]}"
        minor="${BASH_REMATCH[2]}"
        patch="${BASH_REMATCH[3]}"
        new_patch=$((patch + 1))
        echo "v${major}.${minor}.${new_patch}"
    else
        echo "v0.1.0"
    fi
}

# Check if version already exists
check_version_exists() {
    local version="$1"
    if git rev-parse "$version" >/dev/null 2>&1; then
        log_error "Version $version already exists"
        exit 1
    fi
}

# Check if required tools are installed
check_dependencies() {
    local missing_deps=()
    
    if ! command -v go >/dev/null 2>&1; then
        missing_deps+=("go")
    fi
    
    if ! command -v gh >/dev/null 2>&1; then
        missing_deps+=("gh")
    fi
    
    if [[ ${#missing_deps[@]} -ne 0 ]]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        echo ""
        echo "Install missing dependencies:"
        for dep in "${missing_deps[@]}"; do
            case $dep in
                gh)
                    echo "  GitHub CLI: brew install gh (or visit https://cli.github.com/)"
                    ;;
                go)
                    echo "  Go: brew install go (or visit https://golang.org/dl/)"
                    ;;
            esac
        done
        exit 1
    fi
}

# Generate changelog since last version
generate_changelog() {
    local last_version="$1"
    local new_version="$2"
    
    echo "## Changes"
    echo ""
    
    if [[ "$last_version" == "v0.0.0" ]]; then
        echo "- Initial release"
    else
        # Get commits since last version
        git log --pretty=format:"- %s" "$last_version"..HEAD | head -20
        
        # If there are more than 20 commits, add a note
        commit_count=$(git rev-list --count "$last_version"..HEAD)
        if [[ $commit_count -gt 20 ]]; then
            echo "- ... and $((commit_count - 20)) more commits"
        fi
    fi
    
    echo ""
    echo "## Installation"
    echo ""
    echo "### One-line install script:"
    echo '```bash'
    echo "curl -sSL https://raw.githubusercontent.com/$REPO/main/install.sh | bash"
    echo '```'
    echo ""
    echo "### Manual installation:"
    echo "1. Download the binary for your platform"
    echo "2. Move it to a directory in your PATH"
    echo '3. Make it executable: `chmod +x claude-helper`'
    echo ""
    echo "## Platform Support"
    echo "- Linux (amd64, arm64)"
    echo "- macOS (amd64, arm64)"
    echo "- Windows (amd64)"
}

# Main execution
main() {
    log_info "Starting automated release process..."
    
    # Perform checks
    check_git_repo
    check_dependencies
    check_main_branch
    check_clean_working_tree
    
    # Determine version
    if [[ -z "$VERSION" ]]; then
        latest_version=$(get_latest_version)
        VERSION=$(auto_increment_version "$latest_version")
        log_info "Auto-incremented version from $latest_version to $VERSION"
    else
        log_info "Using specified version: $VERSION"
    fi
    
    # Check if version already exists
    check_version_exists "$VERSION"
    
    if [[ "$DRY_RUN" == true ]]; then
        log_warn "DRY RUN MODE - No actual changes will be made"
        echo ""
    fi
    
    log_info "Release summary:"
    echo "  Repository: $REPO"
    echo "  Version: $VERSION"
    echo "  Branch: $(git rev-parse --abbrev-ref HEAD)"
    echo "  Commit: $(git rev-parse --short HEAD)"
    echo ""
    
    if [[ "$DRY_RUN" != true ]]; then
        read -p "Continue with release? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Release cancelled"
            exit 0
        fi
    fi
    
    # Pull latest changes
    log_info "Pulling latest changes..."
    if [[ "$DRY_RUN" != true ]]; then
        git pull origin "$MAIN_BRANCH"
    fi
    
    # Build binaries
    log_info "Building binaries..."
    if [[ "$DRY_RUN" != true ]]; then
        ../build.sh "$VERSION"
    else
        echo "Would run: ../build.sh $VERSION"
    fi
    log_success "Build completed"
    
    # Create and push tag
    log_info "Creating git tag..."
    if [[ "$DRY_RUN" != true ]]; then
        git tag -a "$VERSION" -m "Release $VERSION"
        git push origin "$VERSION"
    else
        echo "Would run: git tag -a $VERSION -m 'Release $VERSION'"
        echo "Would run: git push origin $VERSION"
    fi
    log_success "Git tag created and pushed"
    
    # Generate release notes
    log_info "Generating release notes..."
    latest_version=$(get_latest_version)
    if [[ "$latest_version" == "$VERSION" ]]; then
        # If we just created this version, get the previous one
        latest_version=$(git tag -l "v*.*.*" | sort -V | tail -n2 | head -n1)
    fi
    
    changelog_file="/tmp/claude-helper-changelog-$VERSION.md"
    generate_changelog "$latest_version" "$VERSION" > "$changelog_file"
    
    # Create GitHub release
    log_info "Creating GitHub release..."
    if [[ "$DRY_RUN" != true ]]; then
        if [[ -d "dist" ]]; then
            gh release create "$VERSION" dist/* \
                --title "$VERSION" \
                --notes-file "$changelog_file"
            
            # Clean up
            rm -f "$changelog_file"
        else
            log_error "dist/ directory not found. Build may have failed."
            exit 1
        fi
    else
        echo "Would run: gh release create $VERSION dist/* --title $VERSION --notes-file $changelog_file"
        echo ""
        echo "Generated release notes:"
        cat "$changelog_file"
        rm -f "$changelog_file"
    fi
    
    if [[ "$DRY_RUN" != true ]]; then
        log_success "Release $VERSION created successfully!"
        echo ""
        echo "🎉 Release URL: https://github.com/$REPO/releases/tag/$VERSION"
        echo "📦 Install with: curl -sSL https://raw.githubusercontent.com/$REPO/main/install.sh | bash"
    else
        log_info "Dry run completed. Use without --dry-run to actually release."
    fi
}

# Show help
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    echo "Automated Build and Release Script for claude-helper"
    echo ""
    echo "Usage: $0 [version] [options]"
    echo ""
    echo "Arguments:"
    echo "  version       Version to release (e.g., v1.2.0 or 1.2.0)"
    echo "                If not specified, auto-increments patch version"
    echo ""
    echo "Options:"
    echo "  --dry-run     Test run without making actual changes"
    echo "  --force       Force release even with uncommitted changes or on non-main branch"
    echo "  --help, -h    Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                    # Auto-increment patch version and release"
    echo "  $0 v1.2.0             # Release version v1.2.0"
    echo "  $0 1.2.0              # Release version v1.2.0 (auto-adds 'v' prefix)"
    echo "  $0 --dry-run          # Test what would happen without releasing"
    echo "  $0 v1.2.0 --force     # Force release even with uncommitted changes"
    exit 0
fi

# Run main function
main "$@"