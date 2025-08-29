#!/usr/bin/env python3
import os
import shutil

def main():
    source = "/System/Library/Sounds/Glass.aiff"
    dest = "internal/assets/sounds/notification.aiff"
    
    # Create directory
    os.makedirs(os.path.dirname(dest), exist_ok=True)
    
    try:
        # Copy the file
        shutil.copy2(source, dest)
        print(f"✅ Successfully copied: {source} -> {dest}")
        
        # Get file info
        size = os.path.getsize(dest)
        print(f"📁 File size: {size} bytes")
        
        return True
    except Exception as e:
        print(f"❌ Failed to copy: {e}")
        
        # Create placeholder
        with open(dest, 'w') as f:
            f.write("PLACEHOLDER - Need to manually copy Glass.aiff here")
        print(f"📝 Created placeholder at: {dest}")
        return False

if __name__ == "__main__":
    main()