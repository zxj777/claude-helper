package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zxj777/claude-helper/internal/config"
)

var disableCmd = &cobra.Command{
	Use:   "disable <component-name>",
	Short: "Disable an enabled component",
	Long:  `Disable an installed agent or hook without removing it completely.`,
	Args:  cobra.ExactArgs(1),
	RunE:  disableComponent,
}

func init() {
	rootCmd.AddCommand(disableCmd)
}

func disableComponent(cmd *cobra.Command, args []string) error {
	componentName := args[0]

	fmt.Printf("Disabling component: %s\n", componentName)

	// Check what type of component this is and if it's installed
	componentType, err := detectInstalledComponentType(componentName)
	if err != nil {
		return fmt.Errorf("component '%s' not found or not installed: %w", componentName, err)
	}

	switch componentType {
	case "agent":
		return disableAgent(componentName)
	case "hook":
		return disableHook(componentName)
	default:
		return fmt.Errorf("unsupported component type: %s", componentType)
	}
}

func disableAgent(name string) error {
	// For agents, disabling means moving the file to a .disabled extension
	agentsDir, err := config.GetAgentsPath()
	if err != nil {
		return err
	}

	agentPath := fmt.Sprintf("%s/%s.md", agentsDir, name)
	disabledPath := fmt.Sprintf("%s/%s.md.disabled", agentsDir, name)

	// Check if agent file exists
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		return fmt.Errorf("agent file not found: %s", agentPath)
	}

	// Rename to .disabled
	if err := os.Rename(agentPath, disabledPath); err != nil {
		return fmt.Errorf("failed to disable agent: %w", err)
	}

	fmt.Printf("✓ Agent '%s' has been disabled\n", name)
	return nil
}

func disableHook(name string) error {
	// TODO: Implement hook disabling in settings.json
	// This would require modifying the hook's configuration to set enabled: false
	fmt.Printf("Hook disabling functionality not yet fully implemented for '%s'\n", name)
	fmt.Println("Note: Use 'remove' command to completely remove the hook.")
	return nil
}