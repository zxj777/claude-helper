package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	
	"github.com/zxj777/claude-helper/pkg/types"
)

// GetClaudeConfigPath returns the path to Claude's configuration directory
func GetClaudeConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	var claudePath string
	switch runtime.GOOS {
	case "darwin": // macOS
		claudePath = filepath.Join(homeDir, "Library", "Application Support", "Claude")
	case "linux":
		claudePath = filepath.Join(homeDir, ".config", "Claude")
	case "windows":
		claudePath = filepath.Join(os.Getenv("APPDATA"), "Claude")
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return claudePath, nil
}

// GetAgentsPath returns the path to Claude's agents directory
func GetAgentsPath() (string, error) {
	claudePath, err := GetClaudeConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(claudePath, "agents"), nil
}

// GetSettingsPath returns the path to Claude's settings file
func GetSettingsPath() (string, error) {
	// First, check if we're in a project with local Claude settings
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	
	// Use project-local settings.json (this gets committed to version control)
	localSettingsPath := filepath.Join(cwd, ".claude", "settings.json")
	if _, err := os.Stat(filepath.Dir(localSettingsPath)); err == nil {
		return localSettingsPath, nil
	}
	
	// Fall back to global Claude settings if no .claude directory
	claudePath, err := GetClaudeConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(claudePath, "settings.json"), nil
}

// IsAgentInstalled checks if an agent is installed in the project-local directory
func IsAgentInstalled(agentName string) (bool, error) {
	// Use project-local agents directory
	wd, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("failed to get working directory: %w", err)
	}
	agentFile := filepath.Join(wd, ".claude", "agents", agentName+".md")
	
	_, err = os.Stat(agentFile)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check agent file: %w", err)
	}

	return true, nil
}

// ClaudeSettings represents Claude's settings.json structure
type ClaudeSettings struct {
	Hooks map[string]interface{} `json:"hooks,omitempty"`
	// Add other settings fields as needed
}

// IsHookInstalled checks if a hook is installed in Claude's settings
func IsHookInstalled(hookName string) (bool, error) {
	settingsPath, err := GetSettingsPath()
	if err != nil {
		return false, err
	}

	// Check if settings file exists
	_, err = os.Stat(settingsPath)
	if os.IsNotExist(err) {
		return false, nil // No settings file means no hooks installed
	}
	if err != nil {
		return false, fmt.Errorf("failed to check settings file: %w", err)
	}

	// Read and parse settings file
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return false, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings ClaudeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return false, fmt.Errorf("failed to parse settings file: %w", err)
	}

	// Check if hook exists in any event
	if settings.Hooks == nil {
		return false, nil
	}

	// Search through all hook events to find the specific hook
	for _, eventHooks := range settings.Hooks {
		// eventHooks could be a slice or other structure, need to handle different types
		switch v := eventHooks.(type) {
		case []interface{}:
			for _, hookGroup := range v {
				if hookGroupMap, ok := hookGroup.(map[string]interface{}); ok {
					if hooks, exists := hookGroupMap["hooks"]; exists {
						if hooksSlice, ok := hooks.([]interface{}); ok {
							for _, hook := range hooksSlice {
								if hookMap, ok := hook.(map[string]interface{}); ok {
									if command, exists := hookMap["command"]; exists {
										if commandStr, ok := command.(string); ok {
											// Check if the command contains the hook name
											if strings.Contains(commandStr, hookName) {
												return true, nil
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return false, nil
}

// InstallHookToSettings adds a hook to Claude's settings.json file
func InstallHookToSettings(hook *types.Hook, force bool) error {
	settingsPath, err := GetSettingsPath()
	if err != nil {
		return err
	}

	// Read existing settings or create new
	var settings ClaudeSettings
	
	if data, err := os.ReadFile(settingsPath); err == nil {
		// File exists, parse it
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse existing settings: %w", err)
		}
	}
	
	// Initialize hooks map if nil
	if settings.Hooks == nil {
		settings.Hooks = make(map[string]interface{})
	}

	// Convert hook to Claude format and merge
	claudeHooks := types.MergeHooksIntoClaudeConfig([]types.Hook{*hook})
	hookConfig := claudeHooks["hooks"].(map[string][]map[string]interface{})
	
	// Merge the hook configuration
	for eventName, eventHooks := range hookConfig {
		// Convert new hooks to []interface{}
		converted := make([]interface{}, len(eventHooks))
		for i, hook := range eventHooks {
			converted[i] = hook
		}

		if existingEventHooks, exists := settings.Hooks[eventName]; exists {
			if force {
				// In force mode, find and replace existing hook with same name, or just add if not found
				if existingArray, ok := existingEventHooks.([]interface{}); ok {
					hookName := hook.Name
					var updatedHooks []interface{}
					
					// Go through existing hooks and replace the one with same name
					for _, existingHook := range existingArray {
						shouldKeep := true
						if hookMap, ok := existingHook.(map[string]interface{}); ok {
							if hooks, ok := hookMap["hooks"].([]interface{}); ok {
								for _, h := range hooks {
									if hMap, ok := h.(map[string]interface{}); ok {
										if cmd, ok := hMap["command"].(string); ok {
											// If command contains hook name, replace this hook
											if strings.Contains(cmd, hookName) {
												shouldKeep = false
												break
											}
										}
									}
								}
							}
						}
						if shouldKeep {
							updatedHooks = append(updatedHooks, existingHook)
						}
					}
					
					// Add the new hook
					settings.Hooks[eventName] = append(updatedHooks, converted...)
				} else {
					// Replace entire event hooks if format is unexpected
					settings.Hooks[eventName] = converted
				}
			} else {
				// Non-force mode: just append to existing hooks
				if existingArray, ok := existingEventHooks.([]interface{}); ok {
					settings.Hooks[eventName] = append(existingArray, converted...)
				} else {
					// Replace if existing format is unexpected
					settings.Hooks[eventName] = converted
				}
			}
		} else {
			// Add new event hooks (event doesn't exist yet)
			settings.Hooks[eventName] = converted
		}
	}

	// Write back to settings file
	return writeSettingsFile(settingsPath, &settings)
}

// RemoveHookFromSettings removes a hook from Claude's settings.json file
func RemoveHookFromSettings(hookName string) error {
	settingsPath, err := GetSettingsPath()
	if err != nil {
		return err
	}

	// Check if settings file exists
	data, err := os.ReadFile(settingsPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("settings file not found")
	}
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings ClaudeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to parse settings file: %w", err)
	}

	if settings.Hooks == nil {
		return fmt.Errorf("no hooks found in settings")
	}

	// Remove hook from all events
	removed := false
	for eventName, eventHooks := range settings.Hooks {
		if hooksArray, ok := eventHooks.([]interface{}); ok {
			var filteredHooks []interface{}
			for _, hookEntry := range hooksArray {
				if hookMap, ok := hookEntry.(map[string]interface{}); ok {
					if hooks, hasHooks := hookMap["hooks"]; hasHooks {
						if hooksList, ok := hooks.([]interface{}); ok {
							var filteredHooksList []interface{}
							for _, individualHook := range hooksList {
								if hookData, ok := individualHook.(map[string]interface{}); ok {
									// This is a simplified check - in reality, we'd need more sophisticated matching
									if command, hasCommand := hookData["command"]; hasCommand {
										if !containsHookName(command.(string), hookName) {
											filteredHooksList = append(filteredHooksList, individualHook)
										} else {
											removed = true
										}
									}
								}
							}
							if len(filteredHooksList) > 0 {
								hookMap["hooks"] = filteredHooksList
								filteredHooks = append(filteredHooks, hookEntry)
							}
						}
					}
				}
			}
			// If no hooks remain for this event, set it to an empty array instead of nil/null
			if len(filteredHooks) == 0 {
				settings.Hooks[eventName] = []interface{}{}
			} else {
				settings.Hooks[eventName] = filteredHooks
			}
		}
	}

	if !removed {
		return fmt.Errorf("hook '%s' not found in settings", hookName)
	}

	return writeSettingsFile(settingsPath, &settings)
}

func writeSettingsFile(path string, settings *ClaudeSettings) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	// Marshal settings to JSON with proper formatting
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

func containsHookName(command, hookName string) bool {
	// Check if the command contains the hook name
	// This could be more sophisticated, but for now we check if the hook script name appears in the command
	if command == "" || hookName == "" {
		return false
	}
	
	// Check if the hook name appears in the command (as script name)
	return strings.Contains(command, hookName+".py") || 
		   strings.Contains(command, hookName+".sh") ||
		   strings.Contains(command, hookName+".js") ||
		   strings.Contains(command, "/"+hookName) ||
		   strings.Contains(command, "\\"+hookName)
}

// GetNotificationConfigPath returns the path to the notification config file
func GetNotificationConfigPath() (string, error) {
	// Use project-local config directory
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return filepath.Join(wd, ".claude", "config", "notification.json"), nil
}

// GetAudioConfigPath returns the path to the legacy audio notification config file
// Kept for backward compatibility
func GetAudioConfigPath() (string, error) {
	// Use project-local config directory
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return filepath.Join(wd, ".claude", "config", "audio-notification.json"), nil
}

// IsNotificationInstalled checks if notification is configured
func IsNotificationInstalled() (bool, error) {
	configPath, err := GetNotificationConfigPath()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		// Check for legacy audio config
		return IsAudioNotificationInstalled()
	}
	if err != nil {
		return false, fmt.Errorf("failed to check notification config file: %w", err)
	}

	return true, nil
}

// IsAudioNotificationInstalled checks if legacy audio notification is configured
func IsAudioNotificationInstalled() (bool, error) {
	configPath, err := GetAudioConfigPath()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check audio config file: %w", err)
	}

	return true, nil
}

// LoadNotificationConfig loads the notification configuration
func LoadNotificationConfig() (*types.NotificationConfig, error) {
	configPath, err := GetNotificationConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		// Try to migrate from legacy audio config
		return MigrateLegacyAudioConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read notification config: %w", err)
	}

	var config types.NotificationConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse notification config: %w", err)
	}

	return &config, nil
}

// LoadAudioConfig loads the legacy audio notification configuration
func LoadAudioConfig() (*types.AudioConfig, error) {
	configPath, err := GetAudioConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("audio notification not configured")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read audio config: %w", err)
	}

	var config types.AudioConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse audio config: %w", err)
	}

	return &config, nil
}

// SaveNotificationConfig saves the notification configuration
func SaveNotificationConfig(config *types.NotificationConfig) error {
	configPath, err := GetNotificationConfigPath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON with proper formatting
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal notification config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write notification config file: %w", err)
	}

	return nil
}

// SaveAudioConfig saves the legacy audio notification configuration
func SaveAudioConfig(config *types.AudioConfig) error {
	configPath, err := GetAudioConfigPath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON with proper formatting
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal audio config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write audio config file: %w", err)
	}

	return nil
}

// MigrateLegacyAudioConfig migrates old audio config to new notification config
func MigrateLegacyAudioConfig() (*types.NotificationConfig, error) {
	audioConfig, err := LoadAudioConfig()
	if err != nil {
		return nil, fmt.Errorf("no notification configuration found")
	}

	// Create new notification config from legacy audio config
	notificationConfig := &types.NotificationConfig{
		NotificationTypes: []string{"audio"}, // Default to audio only for migration
		CooldownSecs:      2,                 // Default cooldown
		Desktop: types.DesktopConfig{
			Enabled:     false, // Disabled by default for migration
			ShowDetails: true,
		},
		Audio: *audioConfig,
	}

	// Save the migrated config
	if err := SaveNotificationConfig(notificationConfig); err != nil {
		return nil, fmt.Errorf("failed to save migrated config: %w", err)
	}

	return notificationConfig, nil
}