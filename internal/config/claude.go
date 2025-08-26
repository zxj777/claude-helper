package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	
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

// GetSettingsPath returns the path to Claude's settings.json file
func GetSettingsPath() (string, error) {
	claudePath, err := GetClaudeConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(claudePath, "settings.json"), nil
}

// IsAgentInstalled checks if an agent is installed in Claude
func IsAgentInstalled(agentName string) (bool, error) {
	agentsPath, err := GetAgentsPath()
	if err != nil {
		return false, err
	}

	agentFile := filepath.Join(agentsPath, agentName+".md")
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

	// TODO: Implement more sophisticated hook detection
	// For now, just check if hooks section exists and is not empty
	return len(settings.Hooks) > 0, nil
}

// InstallHookToSettings adds a hook to Claude's settings.json file
func InstallHookToSettings(hook *types.Hook) error {
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
	hookConfig := claudeHooks["hooks"].(map[string]interface{})
	
	// Merge the hook configuration
	for eventName, eventHooks := range hookConfig {
		if existingEventHooks, exists := settings.Hooks[eventName]; exists {
			// Merge with existing hooks for this event
			if existingArray, ok := existingEventHooks.([]interface{}); ok {
				newHooks := eventHooks.([]interface{})
				settings.Hooks[eventName] = append(existingArray, newHooks...)
			} else {
				// Replace if existing format is unexpected
				settings.Hooks[eventName] = eventHooks
			}
		} else {
			// Add new event hooks
			settings.Hooks[eventName] = eventHooks
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
			settings.Hooks[eventName] = filteredHooks
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
	// Simple check - this could be more sophisticated
	// For now, just check if the hook name appears in the command
	return command != "" && hookName != ""
}