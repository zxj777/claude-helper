package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable <component-name>",
	Short: "Enable a disabled component",
	Long:  `Enable a previously installed but disabled agent or hook.`,
	Args:  cobra.ExactArgs(1),
	RunE:  enableComponent,
}

func init() {
	rootCmd.AddCommand(enableCmd)
}

func enableComponent(cmd *cobra.Command, args []string) error {
	componentName := args[0]

	fmt.Printf("Enabling component: %s\n", componentName)

	// Check what type of component this is and if it's installed
	componentType, err := detectInstalledComponentType(componentName)
	if err != nil {
		return fmt.Errorf("component '%s' not found or not installed: %w", componentName, err)
	}

	// For now, enabling/disabling is primarily relevant for hooks
	// Agents are enabled/disabled by their presence in the filesystem
	switch componentType {
	case "agent":
		fmt.Printf("Agent '%s' is controlled by file presence. It's already enabled if installed.\n", componentName)
		return nil
	case "hook":
		return enableHook(componentName)
	default:
		return fmt.Errorf("unsupported component type: %s", componentType)
	}
}

func enableHook(name string) error {
	// TODO: Implement hook enabling in settings.json
	// This would require modifying the hook's "enabled" status or re-adding it
	fmt.Printf("Hook enabling functionality not yet fully implemented for '%s'\n", name)
	fmt.Println("Note: Hooks are enabled by default when installed.")
	return nil
}