package notification

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/zxj777/claude-helper/pkg/types"
)

// DesktopHandler implements desktop notifications across platforms
type DesktopHandler struct {
	config *types.DesktopConfig
}

// NewDesktopHandler creates a new desktop notification handler
func NewDesktopHandler(config *types.DesktopConfig) *DesktopHandler {
	return &DesktopHandler{
		config: config,
	}
}

// Send sends a desktop notification
func (d *DesktopHandler) Send(message NotificationMessage) error {
	if !d.config.Enabled {
		return nil
	}

	switch runtime.GOOS {
	case "darwin":
		return d.sendMacOSNotification(message)
	case "linux":
		return d.sendLinuxNotification(message)
	case "windows":
		return d.sendWindowsNotification(message)
	default:
		return fmt.Errorf("desktop notifications not supported on %s", runtime.GOOS)
	}
}

// IsAvailable checks if desktop notifications are available on this system
func (d *DesktopHandler) IsAvailable() bool {
	if !d.config.Enabled {
		return false
	}

	switch runtime.GOOS {
	case "darwin":
		// osascript is always available on macOS
		return true
	case "linux":
		// Check for notify-send or zenity
		if _, err := exec.LookPath("notify-send"); err == nil {
			return true
		}
		if _, err := exec.LookPath("zenity"); err == nil {
			return true
		}
		return false
	case "windows":
		// Check for PowerShell
		if _, err := exec.LookPath("powershell"); err == nil {
			return true
		}
		return false
	default:
		return false
	}
}

// sendMacOSNotification sends notification on macOS using osascript
func (d *DesktopHandler) sendMacOSNotification(message NotificationMessage) error {
	title := message.Title
	text := message.Message
	
	if d.config.ShowDetails {
		// Add emoji based on message type
		switch message.Type {
		case SuccessMessage:
			text = "✅ " + text
		case ErrorMessage:
			text = "❌ " + text
		case InfoMessage:
			text = "ℹ️ " + text
		}
	}

	script := fmt.Sprintf(`display notification "%s" with title "%s"`, text, title)
	cmd := exec.Command("osascript", "-e", script)
	
	return cmd.Run()
}

// sendLinuxNotification sends notification on Linux
func (d *DesktopHandler) sendLinuxNotification(message NotificationMessage) error {
	title := message.Title
	text := message.Message
	
	if d.config.ShowDetails {
		// Add emoji based on message type
		switch message.Type {
		case SuccessMessage:
			text = "✅ " + text
		case ErrorMessage:
			text = "❌ " + text
		case InfoMessage:
			text = "ℹ️ " + text
		}
	}

	// Try notify-send first
	if _, err := exec.LookPath("notify-send"); err == nil {
		cmd := exec.Command("notify-send", title, text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Try zenity as fallback
	if _, err := exec.LookPath("zenity"); err == nil {
		notificationText := fmt.Sprintf("%s\n%s", title, text)
		cmd := exec.Command("zenity", "--notification", "--text="+notificationText)
		return cmd.Run()
	}

	return fmt.Errorf("no desktop notification system found (tried notify-send, zenity)")
}

// sendWindowsNotification sends notification on Windows
func (d *DesktopHandler) sendWindowsNotification(message NotificationMessage) error {
	title := message.Title
	text := message.Message
	
	if d.config.ShowDetails {
		// Add emoji based on message type
		switch message.Type {
		case SuccessMessage:
			text = "✅ " + text
		case ErrorMessage:
			text = "❌ " + text
		case InfoMessage:
			text = "ℹ️ " + text
		}
	}

	// Use PowerShell to show a simple message box or balloon tip
	// For Git Bash compatibility, we'll use a simple message approach
	psScript := fmt.Sprintf(`
		Add-Type -AssemblyName System.Windows.Forms
		$balloon = New-Object System.Windows.Forms.NotifyIcon
		$balloon.Icon = [System.Drawing.SystemIcons]::Information
		$balloon.BalloonTipTitle = "%s"
		$balloon.BalloonTipText = "%s"
		$balloon.Visible = $true
		$balloon.ShowBalloonTip(3000)
		Start-Sleep -Seconds 1
		$balloon.Dispose()
	`, title, text)

	cmd := exec.Command("powershell", "-Command", psScript)
	return cmd.Run()
}

// GetIconForMessageType returns appropriate icon for message type (Linux)
func (d *DesktopHandler) getIconForMessageType(msgType MessageType) string {
	switch msgType {
	case SuccessMessage:
		return "dialog-information"
	case ErrorMessage:
		return "dialog-error" 
	case InfoMessage:
		return "dialog-information"
	default:
		return "dialog-information"
	}
}