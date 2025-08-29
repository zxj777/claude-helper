package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zxj777/claude-helper/internal/config"
)

var removeCmd = &cobra.Command{
	Use:   "remove <component-name>",
	Short: "Remove a component from Claude Code",
	Long:  `Remove an installed agent or hook from your Claude Code configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  removeComponent,
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
}

func removeComponent(cmd *cobra.Command, args []string) error {
	componentName := args[0]
	skipConfirm, _ := cmd.Flags().GetBool("yes")

	fmt.Printf("Removing component: %s\n", componentName)

	// Check what type of component this is and if it's installed
	componentType, err := detectInstalledComponentType(componentName)
	if err != nil {
		return fmt.Errorf("component '%s' not found or not installed: %w", componentName, err)
	}

	fmt.Printf("Found installed %s: %s\n", componentType, componentName)

	// Confirm removal (unless -y flag is used)
	if !skipConfirm {
		if !confirmRemoval(componentName, componentType) {
			fmt.Println("Removal cancelled.")
			return nil
		}
	}

	// Remove based on component type
	switch componentType {
	case "agent":
		err = removeAgent(componentName)
	case "hook":
		err = removeHook(componentName)
	default:
		return fmt.Errorf("unsupported component type: %s", componentType)
	}

	if err != nil {
		return fmt.Errorf("failed to remove %s '%s': %w", componentType, componentName, err)
	}

	fmt.Printf("✓ Successfully removed %s '%s'\n", componentType, componentName)
	return nil
}

func detectInstalledComponentType(name string) (string, error) {
	// Check if it's an installed agent
	if installed, err := config.IsAgentInstalled(name); err == nil && installed {
		return "agent", nil
	}

	// Check if it's an installed hook
	if installed, err := config.IsHookInstalled(name); err == nil && installed {
		return "hook", nil
	}

	return "", fmt.Errorf("component not installed")
}

func confirmRemoval(name, componentType string) bool {
	fmt.Printf("\nThis will remove the %s '%s' from your Claude Code configuration.\n", componentType, name)
	fmt.Print("Are you sure you want to continue? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func removeAgent(name string) error {
	// Use project-local agents directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	agentPath := filepath.Join(wd, ".claude", "agents", name+".md")

	// Check if file exists
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		return fmt.Errorf("agent file not found: %s", agentPath)
	}

	// Remove the file
	if err := os.Remove(agentPath); err != nil {
		return fmt.Errorf("failed to remove agent file: %w", err)
	}

	return nil
}

func removeHook(name string) error {
	// Remove hook from Claude settings first
	if err := config.RemoveHookFromSettings(name); err != nil {
		return fmt.Errorf("failed to remove hook from settings: %w", err)
	}

	// Remove hook-related files
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Remove hook script files
	hooksDir := filepath.Join(wd, ".claude", "hooks")
	scriptExtensions := []string{".py", ".sh", ".js", ".ts"}
	for _, ext := range scriptExtensions {
		scriptPath := filepath.Join(hooksDir, name+ext)
		if _, err := os.Stat(scriptPath); err == nil {
			if err := os.Remove(scriptPath); err != nil {
				fmt.Printf("Warning: failed to remove hook script %s: %v\n", scriptPath, err)
			} else {
				fmt.Printf("Removed hook script: %s\n", scriptPath)
			}
		}
	}

	// Remove hook-specific config files
	configDir := filepath.Join(wd, ".claude", "config")
	configFiles := []string{
		name + ".json",
		name + "-config.json",
	}
	
	// Special handling for known hooks
	switch name {
	case "audio-notification":
		configFiles = append(configFiles, "audio-notification.json")
	case "task-notification":
		configFiles = append(configFiles, "notification.json")
	case "text-expander":
		configFiles = append(configFiles, "text-expander.json")
	}

	for _, configFile := range configFiles {
		configPath := filepath.Join(configDir, configFile)
		if _, err := os.Stat(configPath); err == nil {
			if err := os.Remove(configPath); err != nil {
				fmt.Printf("Warning: failed to remove config file %s: %v\n", configPath, err)
			} else {
				fmt.Printf("Removed config file: %s\n", configPath)
			}
		}
	}

	// Remove sound files for audio-related hooks
	if name == "audio-notification" || name == "task-notification" {
		soundsDir := filepath.Join(wd, ".claude", "sounds")
		if _, err := os.Stat(soundsDir); err == nil {
			// Ask user if they want to remove sound files
			fmt.Print("Do you want to remove audio files? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err == nil {
				response = strings.ToLower(strings.TrimSpace(response))
				if response == "y" || response == "yes" {
					if err := os.RemoveAll(soundsDir); err != nil {
						fmt.Printf("Warning: failed to remove sounds directory: %v\n", err)
					} else {
						fmt.Printf("Removed sounds directory: %s\n", soundsDir)
					}
				}
			}
		}
	}

	// Remove any temporary or state files
	tempFiles := []string{
		".claude/last-notification-time",
		".claude/last-audio-notification",
		".claude/hook-error.log",
		".claude/notification-error.log",
	}

	for _, tempFile := range tempFiles {
		tempPath := filepath.Join(wd, tempFile)
		if _, err := os.Stat(tempPath); err == nil {
			if err := os.Remove(tempPath); err != nil {
				fmt.Printf("Warning: failed to remove temp file %s: %v\n", tempPath, err)
			} else {
				fmt.Printf("Removed temp file: %s\n", tempPath)
			}
		}
	}

	// Clean up empty directories
	cleanupEmptyDirectories(wd, name)

	return nil
}

func cleanupEmptyDirectories(wd, hookName string) {
	// Check and remove empty directories
	dirsToCheck := []string{
		filepath.Join(wd, ".claude", "hooks"),
		filepath.Join(wd, ".claude", "config"),
		filepath.Join(wd, ".claude", "sounds"),
	}

	for _, dir := range dirsToCheck {
		if isEmpty, err := isDirEmpty(dir); err == nil && isEmpty {
			// Don't remove the directory itself as other components might need it
			// Just leave it empty for now
			fmt.Printf("Directory %s is now empty (keeping for other components)\n", dir)
		}
	}
}

func isDirEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

