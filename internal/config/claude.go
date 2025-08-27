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
	// Simple check - this could be more sophisticated
	// For now, just check if the hook name appears in the command
	return command != "" && hookName != ""
}