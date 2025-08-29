package notification

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/zxj777/claude-helper/pkg/types"
)

// AudioHandler implements audio notifications
type AudioHandler struct {
	config *types.AudioConfig
}

// NewAudioHandler creates a new audio notification handler
func NewAudioHandler(config *types.AudioConfig) *AudioHandler {
	return &AudioHandler{
		config: config,
	}
}

// Send sends an audio notification
func (a *AudioHandler) Send(message NotificationMessage) error {
	if !a.config.Enabled {
		return nil
	}

	// Select appropriate sound file based on message type
	soundFile := a.getSoundFileForMessage(message)
	soundPath := a.findSoundFile(soundFile)
	
	if soundPath == "" {
		return fmt.Errorf("sound file not found: %s", soundFile)
	}

	return a.playAudioFile(soundPath)
}

// IsAvailable checks if audio notifications are available
func (a *AudioHandler) IsAvailable() bool {
	if !a.config.Enabled {
		return false
	}

	switch runtime.GOOS {
	case "darwin":
		// Check for afplay
		_, err := exec.LookPath("afplay")
		return err == nil
	case "linux":
		// Check for common audio players
		players := []string{"aplay", "paplay", "play"}
		for _, player := range players {
			if _, err := exec.LookPath(player); err == nil {
				return true
			}
		}
		return false
	case "windows":
		// Check for PowerShell (for Media.SoundPlayer)
		_, err := exec.LookPath("powershell")
		return err == nil
	default:
		return false
	}
}

// getSoundFileForMessage returns the appropriate sound file for the message
func (a *AudioHandler) getSoundFileForMessage(message NotificationMessage) string {
	switch message.Type {
	case SuccessMessage:
		if a.config.SuccessSound != "" {
			return a.config.SuccessSound
		}
	case ErrorMessage:
		if a.config.ErrorSound != "" {
			return a.config.ErrorSound
		}
	}
	
	// Default sound
	if a.config.DefaultSound != "" {
		return a.config.DefaultSound
	}
	
	return "complete.wav"
}

// findSoundFile locates the sound file in various directories
func (a *AudioHandler) findSoundFile(soundFile string) string {
	// If absolute path, use as-is
	if filepath.IsAbs(soundFile) {
		if _, err := os.Stat(soundFile); err == nil {
			return soundFile
		}
		return ""
	}

	// Check in project .claude/sounds directory
	wd, err := os.Getwd()
	if err == nil {
		projectSound := filepath.Join(wd, ".claude", "sounds", soundFile)
		if _, err := os.Stat(projectSound); err == nil {
			return projectSound
		}
	}

	// Check in sounds directory relative to hooks
	if wd != "" {
		hooksSound := filepath.Join(wd, ".claude", "hooks", "..", "sounds", soundFile)
		if _, err := os.Stat(hooksSound); err == nil {
			return hooksSound
		}
	}

	return ""
}

// playAudioFile plays the audio file using platform-appropriate command
func (a *AudioHandler) playAudioFile(soundPath string) error {
	switch runtime.GOOS {
	case "darwin":
		return a.playMacOSAudio(soundPath)
	case "linux":
		return a.playLinuxAudio(soundPath)
	case "windows":
		return a.playWindowsAudio(soundPath)
	default:
		return fmt.Errorf("audio playback not supported on %s", runtime.GOOS)
	}
}

// playMacOSAudio plays audio on macOS
func (a *AudioHandler) playMacOSAudio(soundPath string) error {
	cmd := exec.Command("afplay", soundPath)
	return cmd.Run()
}

// playLinuxAudio plays audio on Linux
func (a *AudioHandler) playLinuxAudio(soundPath string) error {
	// Try different audio players in order of preference
	players := []string{"aplay", "paplay", "play"}
	
	for _, player := range players {
		if _, err := exec.LookPath(player); err == nil {
			cmd := exec.Command(player, soundPath)
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
	}
	
	return fmt.Errorf("no working audio player found (tried: %v)", players)
}

// playWindowsAudio plays audio on Windows
func (a *AudioHandler) playWindowsAudio(soundPath string) error {
	// Use PowerShell with Media.SoundPlayer for Git Bash compatibility
	psScript := fmt.Sprintf(`
		$sound = New-Object Media.SoundPlayer '%s'
		$sound.PlaySync()
	`, soundPath)
	
	cmd := exec.Command("powershell", "-Command", psScript)
	return cmd.Run()
}