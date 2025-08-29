package assets

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed sounds/*
var soundsFS embed.FS

// GetTemplatesDir returns the path to the templates directory
// If running from source (development), it uses the local internal/assets/templates
// If running from built binary, it extracts embedded files to a temp location
func GetTemplatesDir() (string, error) {
	// First try to find local templates directory (for development)
	if wd, err := os.Getwd(); err == nil {
		localTemplatesDir := filepath.Join(wd, "internal", "assets", "templates")
		if _, err := os.Stat(localTemplatesDir); err == nil {
			return localTemplatesDir, nil
		}
	}

	// If not found locally, extract embedded files to temp directory
	tempDir, err := os.MkdirTemp("", "claude-helper-templates-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Extract embedded templates to temp directory
	err = fs.WalkDir(templatesFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Create the target path in temp directory
		targetPath := filepath.Join(tempDir, path)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Read file content from embedded FS
		content, err := templatesFS.ReadFile(path)
		if err != nil {
			return err
		}

		// Create parent directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// Write file to temp location
		return os.WriteFile(targetPath, content, 0644)
	})

	if err != nil {
		os.RemoveAll(tempDir) // Clean up on error
		return "", fmt.Errorf("failed to extract embedded templates: %w", err)
	}

	return filepath.Join(tempDir, "templates"), nil
}

// ListAgentTemplates returns a list of available agent templates
func ListAgentTemplates() ([]string, error) {
	templatesDir, err := GetTemplatesDir()
	if err != nil {
		return nil, err
	}

	agentsDir := filepath.Join(templatesDir, "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read agents directory: %w", err)
	}

	var agents []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			name := entry.Name()[:len(entry.Name())-3] // Remove .md extension
			agents = append(agents, name)
		}
	}

	return agents, nil
}

// ListHookTemplates returns a list of available hook templates
func ListHookTemplates() ([]string, error) {
	templatesDir, err := GetTemplatesDir()
	if err != nil {
		return nil, err
	}

	hooksDir := filepath.Join(templatesDir, "hooks")
	entries, err := os.ReadDir(hooksDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read hooks directory: %w", err)
	}

	var hooks []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			name := entry.Name()[:len(entry.Name())-5] // Remove .yaml extension
			hooks = append(hooks, name)
		}
	}

	return hooks, nil
}

// GetTemplatePath returns the full path to a template file
func GetTemplatePath(templateType, name string) (string, error) {
	templatesDir, err := GetTemplatesDir()
	if err != nil {
		return "", err
	}

	var templatePath string
	switch templateType {
	case "agent":
		templatePath = filepath.Join(templatesDir, "agents", name+".md")
	case "hook":
		templatePath = filepath.Join(templatesDir, "hooks", name+".yaml")
	default:
		return "", fmt.Errorf("unknown template type: %s", templateType)
	}

	if _, err := os.Stat(templatePath); err != nil {
		return "", fmt.Errorf("template not found: %s", templatePath)
	}

	return templatePath, nil
}

// GetSoundsDir returns the path to the sounds directory
// If running from source (development), it uses the local internal/assets/sounds
// If running from built binary, it extracts embedded sounds to a temp location
func GetSoundsDir() (string, error) {
	// First try to find local sounds directory (for development)
	if wd, err := os.Getwd(); err == nil {
		localSoundsDir := filepath.Join(wd, "internal", "assets", "sounds")
		if _, err := os.Stat(localSoundsDir); err == nil {
			return localSoundsDir, nil
		}
	}

	// If not found locally, extract embedded files to temp directory
	tempDir, err := os.MkdirTemp("", "claude-helper-sounds-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Extract embedded sounds to temp directory
	err = fs.WalkDir(soundsFS, "sounds", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Create the target path in temp directory
		targetPath := filepath.Join(tempDir, path)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Read file content from embedded FS
		content, err := soundsFS.ReadFile(path)
		if err != nil {
			return err
		}

		// Create parent directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// Write file to temp location
		return os.WriteFile(targetPath, content, 0644)
	})

	if err != nil {
		os.RemoveAll(tempDir) // Clean up on error
		return "", fmt.Errorf("failed to extract embedded sounds: %w", err)
	}

	return filepath.Join(tempDir, "sounds"), nil
}

// GetSoundFilePath returns the full path to a sound file
func GetSoundFilePath(filename string) (string, error) {
	soundsDir, err := GetSoundsDir()
	if err != nil {
		return "", err
	}

	soundPath := filepath.Join(soundsDir, filename)
	
	if _, err := os.Stat(soundPath); err != nil {
		return "", fmt.Errorf("sound file not found: %s", soundPath)
	}

	return soundPath, nil
}

// GetPlatformNotificationSound returns the platform-appropriate system notification sound
func GetPlatformNotificationSound() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS: Use the classic Glass sound (compatible with existing AIFF-based logic)
		return "/System/Library/Sounds/Glass.aiff"
	case "windows":
		// Windows: Use the classic Ding sound, which is pleasant and widely recognized
		return "C:\\Windows\\Media\\Windows Ding.wav"
	case "linux":
		// Linux: Try common system notification sounds
		// Check for common locations in order of preference
		candidates := []string{
			"/usr/share/sounds/alsa/Side_Left.wav",           // ALSA default
			"/usr/share/sounds/ubuntu/notifications/Blip.ogg", // Ubuntu
			"/usr/share/sounds/generic/notifications/complete.oga", // Generic
			"/usr/share/sounds/freedesktop/stereo/complete.oga",    // Freedesktop
		}
		
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
		
		// Fallback: return a generic path that most Linux systems can handle
		return "/usr/share/sounds/alsa/Side_Left.wav"
	default:
		// For unknown platforms, return a generic filename
		return "notification.wav"
	}
}

// GetPlatformNotificationSoundWithFallback returns platform-appropriate notification sound with fallback logic
func GetPlatformNotificationSoundWithFallback() (string, error) {
	// First try platform-specific system sound
	systemSound := GetPlatformNotificationSound()
	if _, err := os.Stat(systemSound); err == nil {
		return systemSound, nil
	}

	// If system sound not found, try project-specific sounds
	projectSounds := []string{"notification.wav", "notification.aiff", "complete.wav"}
	
	// Check in project .claude/sounds directory
	if wd, err := os.Getwd(); err == nil {
		for _, sound := range projectSounds {
			projectSound := filepath.Join(wd, ".claude", "sounds", sound)
			if _, err := os.Stat(projectSound); err == nil {
				return projectSound, nil
			}
		}
	}

	// Check embedded sounds
	for _, sound := range projectSounds {
		if soundPath, err := GetSoundFilePath(sound); err == nil {
			return soundPath, nil
		}
	}

	// Last resort: return system sound path even if it doesn't exist
	// (the audio player will handle the error gracefully)
	return systemSound, nil
}
