#!/bin/bash

# Safe Project Organization Script for claude-helper
# This script safely reorganizes the project structure while preserving build integrity
# 🚨 CRITICAL: Preserves Go embed paths and GitHub-referenced files

set -e

echo "🧹 Safely organizing claude-helper project structure..."

# Create new directories
echo "📁 Creating directories..."
mkdir -p docs scripts

# Move documentation files to docs/
echo "📄 Moving documentation files..."
mv AUDIO_NOTIFICATION_USAGE.md docs/
mv DESKTOP_NOTIFICATION_SUMMARY.md docs/ 
mv IMPLEMENTATION_SUMMARY.md docs/
mv audio-notification-design.md docs/
mv implement.md docs/
mv project-description.md docs/
mv test-remove-functionality.md docs/
mv text-expander-implementation.md docs/

# Move SAFE scripts to scripts/ (excluding critical build-referenced files)
echo "🔧 Moving non-critical scripts..."
mv auto-release.sh scripts/
mv cleanup.sh scripts/
mv cleanup_templates.sh scripts/
mv copy_audio_manual.py scripts/
mv copy_sound.py scripts/
mv fix-hook.sh scripts/
mv generate-sounds.py scripts/
mv quick-release.sh scripts/
mv release.sh scripts/
mv setup_audio.py scripts/
mv setup_embedded_sound.sh scripts/

# ⚠️ KEEP these files in root directory (required for build/deploy):
echo "⚠️  Keeping critical files in root directory:"
echo "   - build.sh (referenced by release scripts)"
echo "   - install.sh (GitHub URL hardcoded: /main/install.sh)" 
echo "   - install.ps1 (GitHub URL hardcoded: /main/install.ps1)"

# Remove temporary/build artifacts
echo "🗑️  Removing temporary files..."
rm -f claude-helper  # Binary that should be in bin/
rm -f test_assets.go test_escape_logic.py test_settings.json test_templates.go

# Update script references in moved scripts
echo "🔄 Updating script references..."

# Update auto-release.sh to reference build.sh in root
sed -i.bak 's|./build\.sh|../build.sh|g' scripts/auto-release.sh

# Update quick-release.sh to reference build.sh in root
sed -i.bak 's|./build\.sh|../build.sh|g' scripts/quick-release.sh

# Update release.sh to reference build.sh in root
sed -i.bak 's|./build\.sh|../build.sh|g' scripts/release.sh

# Clean up backup files
rm -f scripts/*.bak

echo "✅ Safe project organization complete!"
echo ""
echo "📋 New structure:"
echo "  docs/              - Documentation files"  
echo "  scripts/           - Development and utility scripts"
echo "  build.sh           - 🔒 KEPT IN ROOT (referenced by release scripts)"
echo "  install.sh         - 🔒 KEPT IN ROOT (GitHub URL: /main/install.sh)"
echo "  install.ps1        - 🔒 KEPT IN ROOT (GitHub URL: /main/install.ps1)"
echo "  cmd/               - Entry points (unchanged)"
echo "  internal/          - Internal packages (unchanged)"
echo "    ├── assets/      - 🔒 CRITICAL: Contains Go embed paths"
echo "    │   ├── templates/ - 🔒 Required for //go:embed templates/*"
echo "    │   └── sounds/    - 🔒 Required for //go:embed sounds/*"
echo "  pkg/               - Public packages (unchanged)" 
echo ""
echo "🛡️  Build integrity preserved:"
echo "   ✅ Go embed paths intact (internal/assets/{templates,sounds}/)"
echo "   ✅ GitHub install URLs preserved (/main/install.sh, /main/install.ps1)"
echo "   ✅ Release scripts updated to reference ../build.sh"
echo ""
echo "🎯 Cleaned up temporary files:"
echo "   - claude-helper (binary build artifact)"
echo "   - test_*.go, test_*.py, test_*.json (temporary test files)"