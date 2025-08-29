package main

import (
	"fmt"
	"log"
	"os"
	
	"github.com/zxj777/claude-helper/internal/assets"
)

func main() {
	fmt.Println("🧪 Testing embedded assets...")
	
	// Test sound file path
	soundPath, err := assets.GetSoundFilePath("notification.aiff")
	if err != nil {
		log.Printf("❌ Failed to get sound file path: %v", err)
	} else {
		fmt.Printf("✅ Sound file path: %s\n", soundPath)
		
		// Check if file exists
		if _, err := os.Stat(soundPath); err == nil {
			fmt.Printf("✅ Sound file exists and is accessible\n")
		} else {
			fmt.Printf("❌ Sound file not accessible: %v\n", err)
		}
	}
	
	// Test templates
	templatesDir, err := assets.GetTemplatesDir()
	if err != nil {
		log.Printf("❌ Failed to get templates dir: %v", err)
	} else {
		fmt.Printf("✅ Templates directory: %s\n", templatesDir)
	}
	
	fmt.Println("🎯 Asset test completed")
}