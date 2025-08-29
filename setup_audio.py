#!/usr/bin/env python3
import shutil
import os
import sys

def setup_audio_files():
    """Copy system audio file to internal assets"""
    
    # Source and destination paths
    source = '/System/Library/Sounds/Glass.aiff'
    sounds_dir = 'internal/assets/sounds'
    dest = os.path.join(sounds_dir, 'notification.aiff')
    
    # Ensure sounds directory exists
    os.makedirs(sounds_dir, exist_ok=True)
    
    try:
        # Copy the Glass.aiff file
        shutil.copy2(source, dest)
        print(f"✅ Successfully copied {source} to {dest}")
        
        # Verify the file exists and get its size
        if os.path.exists(dest):
            size = os.path.getsize(dest)
            print(f"📁 File size: {size} bytes")
        
        return True
        
    except Exception as e:
        print(f"❌ Failed to copy audio file: {e}")
        return False

if __name__ == '__main__':
    success = setup_audio_files()
    sys.exit(0 if success else 1)