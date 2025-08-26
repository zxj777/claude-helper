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
	// Get Claude agents directory
	agentsDir, err := config.GetAgentsPath()
	if err != nil {
		return err
	}

	// Build agent file path
	agentPath := filepath.Join(agentsDir, name+".md")

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
	return config.RemoveHookFromSettings(name)
}