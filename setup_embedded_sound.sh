#!/bin/bash

# Setup script to copy system sound to embedded assets
echo "🔊 Setting up embedded notification sound..."

# Source and destination
SOURCE="/System/Library/Sounds/Glass.aiff"
DEST="internal/assets/sounds/notification.aiff"

# Check if source exists
if [ ! -f "$SOURCE" ]; then
    echo "❌ System sound not found: $SOURCE"
    echo "Creating placeholder instead..."
    
    # Create placeholder content
    mkdir -p "$(dirname "$DEST")"
    echo "PLACEHOLDER_SOUND_FILE - Replace with actual Glass.aiff" > "$DEST"
    echo "📝 Created placeholder at: $DEST"
    echo "Please manually copy Glass.aiff to this location"
    exit 0
fi

# Ensure destination directory exists
mkdir -p "$(dirname "$DEST")"

# Copy the file
if cp "$SOURCE" "$DEST"; then
    echo "✅ Successfully copied system sound to: $DEST"
    
    # Get file size
    SIZE=$(stat -f%z "$DEST" 2>/dev/null || stat -c%s "$DEST" 2>/dev/null || echo "unknown")
    echo "📁 File size: $SIZE bytes"
    
    echo "🎵 Embedded sound ready for build!"
else
    echo "❌ Failed to copy sound file"
    exit 1
fi