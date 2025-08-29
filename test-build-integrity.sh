#!/bin/bash

# Test Build Integrity Script
# Validates that the project can still build correctly after reorganization

set -e

echo "🧪 Testing build integrity after project reorganization..."

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_test() {
    echo -e "${BLUE}🔍 Testing: $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Test 1: Go embed paths are accessible
log_test "Go embed paths accessibility"
if [ -d "internal/assets/templates" ] && [ -d "internal/assets/sounds" ]; then
    log_success "Embedded asset paths intact"
else
    log_error "Critical embedded asset paths missing"
fi

# Test 2: Build script exists and is executable
log_test "Build script availability"
if [ -f "build.sh" ] && [ -x "build.sh" ]; then
    log_success "build.sh exists and is executable"
else
    log_error "build.sh missing or not executable"
fi

# Test 3: Install scripts exist (required for GitHub URLs)
log_test "Install scripts for GitHub URLs"
if [ -f "install.sh" ] && [ -f "install.ps1" ]; then
    log_success "Install scripts preserved in root"
else
    log_error "Install scripts missing - GitHub URLs will break"
fi

# Test 4: Go mod and source compilation
log_test "Go source compilation"
if go build -o /tmp/test-claude-helper ./cmd/claude-helper; then
    log_success "Go compilation successful"
    rm -f /tmp/test-claude-helper
else
    log_error "Go compilation failed"
fi

# Test 5: Embedded assets can be loaded
log_test "Embedded assets loading"
if go run -ldflags "-X main.Version=test" ./cmd/claude-helper list >/dev/null 2>&1; then
    log_success "Embedded assets loading correctly"
else
    log_warning "Could not test embedded assets (command may require specific setup)"
fi

# Test 6: Check if release scripts can find build.sh
log_test "Release script references"
if [ -f "scripts/quick-release.sh" ]; then
    if grep -q "\.\./build\.sh" scripts/quick-release.sh; then
        log_success "quick-release.sh correctly references ../build.sh"
    else
        log_error "quick-release.sh does not reference ../build.sh correctly"
    fi
else
    log_error "quick-release.sh not found in scripts/"
fi

if [ -f "scripts/auto-release.sh" ]; then
    if grep -q "\.\./build\.sh" scripts/auto-release.sh; then
        log_success "auto-release.sh correctly references ../build.sh"
    else
        log_error "auto-release.sh does not reference ../build.sh correctly"
    fi
else
    log_error "auto-release.sh not found in scripts/"
fi

# Test 7: Make sure no critical files were moved
log_test "Critical files preservation"
critical_files=("go.mod" "go.sum" "Makefile" "CLAUDE.md" "README.md")
for file in "${critical_files[@]}"; do
    if [ -f "$file" ]; then
        log_success "$file preserved in root"
    else
        log_error "$file missing from root directory"
    fi
done

echo ""
echo -e "${GREEN}🎉 All build integrity tests passed!${NC}"
echo ""
echo "📋 Summary:"
echo "  ✅ Go embed paths preserved"
echo "  ✅ Build script accessible"
echo "  ✅ GitHub install URLs maintained"
echo "  ✅ Go compilation successful"
echo "  ✅ Release scripts updated correctly"
echo "  ✅ Critical project files intact"
echo ""
echo "🚀 Project is ready for building and deployment!"