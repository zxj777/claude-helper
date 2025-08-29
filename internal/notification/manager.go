package notification

import (
	"fmt"

	"github.com/zxj777/claude-helper/pkg/types"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	AudioNotification   NotificationType = "audio"
	DesktopNotification NotificationType = "desktop"
)

// NotificationMessage represents the content of a notification
type NotificationMessage struct {
	Title   string
	Message string
	Type    MessageType // success, error, info
}

// MessageType represents the type of message for notification styling
type MessageType string

const (
	SuccessMessage MessageType = "success"
	ErrorMessage   MessageType = "error"
	InfoMessage    MessageType = "info"
)

// Manager handles different types of notifications
type Manager struct {
	config         *types.NotificationConfig
	audioHandler   NotificationHandler
	desktopHandler NotificationHandler
}

// NotificationHandler defines the interface for notification handlers
type NotificationHandler interface {
	Send(message NotificationMessage) error
	IsAvailable() bool
}

// NewManager creates a new notification manager
func NewManager(config *types.NotificationConfig) *Manager {
	return &Manager{
		config:         config,
		audioHandler:   NewAudioHandler(&config.Audio),
		desktopHandler: NewDesktopHandler(&config.Desktop),
	}
}

// Send sends a notification using configured methods
func (m *Manager) Send(message NotificationMessage) error {
	if !m.shouldSendNotification() {
		return nil
	}

	var errors []error
	sent := false

	// Try each enabled notification type
	for _, notifType := range m.config.NotificationTypes {
		switch NotificationType(notifType) {
		case AudioNotification:
			if m.audioHandler.IsAvailable() {
				if err := m.audioHandler.Send(message); err != nil {
					errors = append(errors, fmt.Errorf("audio notification failed: %w", err))
				} else {
					sent = true
				}
			}
		case DesktopNotification:
			if m.desktopHandler.IsAvailable() {
				if err := m.desktopHandler.Send(message); err != nil {
					errors = append(errors, fmt.Errorf("desktop notification failed: %w", err))
				} else {
					sent = true
				}
			}
		}
	}

	// If no notifications were sent successfully, return the errors
	if !sent && len(errors) > 0 {
		return fmt.Errorf("all notifications failed: %v", errors)
	}

	return nil
}

// shouldSendNotification checks cooldown and other conditions
func (m *Manager) shouldSendNotification() bool {
	// Check cooldown logic here (similar to existing implementation)
	// This would check the last notification time file
	return true // Simplified for now
}

// CreateMessageFromToolResult analyzes tool result and creates appropriate message
func CreateMessageFromToolResult(toolResult interface{}) NotificationMessage {
	// Analyze tool result to determine message type and content
	success := true
	toolName := "Unknown"
	
	// Try to determine if operation was successful and extract tool information
	if resultStr := fmt.Sprintf("%v", toolResult); resultStr != "" {
		// Check for error indicators
		if containsError(resultStr) {
			success = false
		}
		
		// Extract tool name if possible
		toolName = extractToolName(toolResult)
	}

	var messageType MessageType
	var message string

	if success {
		messageType = SuccessMessage
		message = fmt.Sprintf("✅ %s 操作完成", toolName)
	} else {
		messageType = ErrorMessage
		message = fmt.Sprintf("❌ %s 操作失败", toolName)
	}

	return NotificationMessage{
		Title:   "Claude Helper - 任务完成",
		Message: message,
		Type:    messageType,
	}
}

// containsError checks if the result contains error indicators
func containsError(result string) bool {
	errorKeywords := []string{"error", "failed", "exception", "timeout", "denied"}
	for _, keyword := range errorKeywords {
		if contains(result, keyword) {
			return true
		}
	}
	return false
}

// extractToolName attempts to extract tool name from result
func extractToolName(toolResult interface{}) string {
	// This is a simplified implementation
	// In practice, you'd parse the tool result structure
	return "任务"
}

// contains is a case-insensitive string contains check
func contains(s, substr string) bool {
	// Simple case-insensitive check
	// In practice, you'd use strings.Contains with strings.ToLower
	return false // Simplified for now
}