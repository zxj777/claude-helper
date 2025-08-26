package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"github.com/zxj777/claude-helper/internal/config"
	"github.com/zxj777/claude-helper/pkg/types"
)

var installCmd = &cobra.Command{
	Use:   "install <component-name>",
	Short: "Install a component to Claude Code",
	Long:  `Install an agent or hook template to your Claude Code configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  installComponent,
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().BoolP("force", "f", false, "Force install even if component already exists")
}

func installComponent(cmd *cobra.Command, args []string) error {
	componentName := args[0]
	force, _ := cmd.Flags().GetBool("force")

	fmt.Printf("Installing component: %s\n", componentName)

	// Find the component template
	templatePath, componentType, err := findComponentTemplate(componentName)
	if err != nil {
		return fmt.Errorf("failed to find component '%s': %w", componentName, err)
	}

	fmt.Printf("Found %s template at: %s\n", componentType, templatePath)

	// Check if already installed (unless force is used)
	if !force {
		if installed, err := isComponentInstalled(componentName, componentType); err == nil && installed {
			return fmt.Errorf("component '%s' is already installed. Use --force to reinstall", componentName)
		}
	}

	// Install based on component type
	switch componentType {
	case "agent":
		err = installAgent(componentName, templatePath)
	case "hook":
		err = installHook(componentName, templatePath)
	default:
		return fmt.Errorf("unsupported component type: %s", componentType)
	}

	if err != nil {
		return fmt.Errorf("failed to install %s '%s': %w", componentType, componentName, err)
	}

	fmt.Printf("✓ Successfully installed %s '%s'\n", componentType, componentName)
	return nil
}

func findComponentTemplate(name string) (templatePath string, componentType string, err error) {
	// Get current working directory and build templates path
	wd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get working directory: %w", err)
	}
	templatesDir := filepath.Join(wd, "assets", "templates")

	// Check for agent template
	agentPath := filepath.Join(templatesDir, "agents", name+".md")
	if _, err := os.Stat(agentPath); err == nil {
		return agentPath, "agent", nil
	}

	// Check for hook template
	hookPath := filepath.Join(templatesDir, "hooks", name+".yaml")
	if _, err := os.Stat(hookPath); err == nil {
		return hookPath, "hook", nil
	}

	return "", "", fmt.Errorf("template not found")
}

func isComponentInstalled(name, componentType string) (bool, error) {
	switch componentType {
	case "agent":
		return config.IsAgentInstalled(name)
	case "hook":
		return config.IsHookInstalled(name)
	default:
		return false, fmt.Errorf("unknown component type: %s", componentType)
	}
}

func installAgent(name, templatePath string) error {
	// Get Claude agents directory
	agentsDir, err := config.GetAgentsPath()
	if err != nil {
		return err
	}

	// Create agents directory if it doesn't exist
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	// Read template content
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Write to Claude agents directory
	targetPath := filepath.Join(agentsDir, name+".md")
	if err := os.WriteFile(targetPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write agent file: %w", err)
	}

	return nil
}

func installHook(name, templatePath string) error {
	// Special handling for text-expander hook
	if name == "text-expander" {
		if err := configureTextExpander(); err != nil {
			return fmt.Errorf("failed to configure text expander: %w", err)
		}
	}

	// Read hook template
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read hook template: %w", err)
	}

	// Parse hook from YAML
	hook, err := parseHookFromYAML(content)
	if err != nil {
		return fmt.Errorf("failed to parse hook template: %w", err)
	}

	// Install hook to Claude settings
	return installHookToSettings(hook)
}

func configureTextExpander() error {
	fmt.Println("🔧 Configuring Text Expander mappings...")
	fmt.Println("You can create shortcuts that expand to longer text.")
	fmt.Println("Example: -d -> 'Please provide detailed explanation'")
	fmt.Println("Press Enter with empty marker to finish configuration.")
	fmt.Println()

	// Get config directory path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	
	configDir := filepath.Join(homeDir, ".claude-helper")
	configPath := filepath.Join(configDir, "text-expander-config.json")

	// Create config directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config or create new
	var textConfig TextExpanderConfig
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &textConfig); err != nil {
			fmt.Printf("Warning: existing config file is corrupted, creating new one\n")
		}
	}
	if textConfig.Mappings == nil {
		textConfig.Mappings = make(map[string]string)
	}

	// Interactive input
	reader := bufio.NewReader(os.Stdin)
	
	for {
		fmt.Print("Enter marker (e.g., -d, -v, --explain): ")
		marker, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		
		marker = strings.TrimSpace(marker)
		if marker == "" {
			break // Empty input ends configuration
		}

		// Validate marker
		if !isValidMarker(marker) {
			fmt.Println("❌ Invalid marker. Use format like: -d, -v, --explain, debug")
			continue
		}

		fmt.Printf("Enter replacement text for '%s': ", marker)
		replacement, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read replacement: %w", err)
		}
		
		replacement = strings.TrimSpace(replacement)
		if replacement == "" {
			fmt.Println("❌ Replacement text cannot be empty")
			continue
		}

		// Check if marker already exists
		if existing, exists := textConfig.Mappings[marker]; exists {
			fmt.Printf("⚠️  Marker '%s' already exists with value: '%s'\n", marker, existing)
			fmt.Print("Overwrite? (y/N): ")
			confirm, _ := reader.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
				continue
			}
		}

		textConfig.Mappings[marker] = replacement
		fmt.Printf("✅ Added mapping: '%s' → '%s'\n", marker, replacement)
		fmt.Println()
	}

	if len(textConfig.Mappings) == 0 {
		fmt.Println("No mappings configured. Text expander will be installed but inactive.")
	} else {
		fmt.Printf("📝 Total mappings configured: %d\n", len(textConfig.Mappings))
	}

	// Save configuration
	return saveTextExpanderConfig(configPath, &textConfig)
}

func isValidMarker(marker string) bool {
	if len(marker) == 0 {
		return false
	}
	
	// Allow markers starting with - or -- or just alphanumeric words
	if strings.HasPrefix(marker, "-") {
		return len(marker) > 1
	}
	
	// Allow simple word markers
	for _, char := range marker {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || char == '_' || char == '-') {
			return false
		}
	}
	
	return true
}

func saveTextExpanderConfig(configPath string, textConfig *TextExpanderConfig) error {
	data, err := json.MarshalIndent(textConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("💾 Configuration saved to: %s\n", configPath)
	return nil
}

func parseHookFromYAML(content []byte) (*types.Hook, error) {
	var hook types.Hook
	
	if err := yaml.Unmarshal(content, &hook); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	
	// Validate required fields
	if hook.Name == "" {
		return nil, fmt.Errorf("hook name is required")
	}
	if hook.Event == "" {
		return nil, fmt.Errorf("hook event is required")
	}
	if hook.Command == "" {
		return nil, fmt.Errorf("hook command is required")
	}
	
	// Set default values
	if hook.Timeout == 0 {
		hook.Timeout = 30 // Default 30 seconds timeout
	}
	
	return &hook, nil
}

func installHookToSettings(hook *types.Hook) error {
	return config.InstallHookToSettings(hook)
}