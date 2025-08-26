package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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
	// Use project-local agents directory (.claude/agents/)
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	agentsDir := filepath.Join(wd, ".claude", "agents")

	// Create agents directory if it doesn't exist
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	// Read template content
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Write to project-local agents directory
	targetPath := filepath.Join(agentsDir, name+".md")
	if err := os.WriteFile(targetPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write agent file: %w", err)
	}

	fmt.Printf("Agent installed at: %s\n", targetPath)
	return nil
}

func installHook(name, templatePath string) error {
	// Special handling for text-expander hook - configure mappings before setup
	if name == "text-expander" {
		if err := configureTextExpanderMappings(); err != nil {
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

	// Execute setup script if present
	if hook.Setup != "" {
		if err := executeSetupScript(hook.Setup); err != nil {
			return fmt.Errorf("failed to execute setup script: %w", err)
		}
	}

	// Install hook to Claude settings
	return installHookToSettings(hook)
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

func executeSetupScript(setupScript string) error {
	fmt.Println("🔧 Executing setup script...")
	
	// Create a temporary script file
	tmpFile, err := os.CreateTemp("", "claude-helper-setup-*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temp script file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	
	// Write the setup script to the temp file
	if _, err := tmpFile.WriteString(setupScript); err != nil {
		return fmt.Errorf("failed to write setup script: %w", err)
	}
	tmpFile.Close()
	
	// Make the script executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}
	
	// Execute the script
	cmd := exec.Command("/bin/bash", tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir, _ = os.Getwd() // Set working directory to current directory
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("setup script failed: %w", err)
	}
	
	fmt.Println("✓ Setup script executed successfully")
	return nil
}

func configureTextExpanderMappings() error {
	fmt.Println("🔧 Configuring Text Expander mappings...")
	fmt.Println("You can create shortcuts that expand to longer text.")
	fmt.Println("Example: -d -> '详细解释这段代码的功能、实现原理和使用方法'")
	fmt.Println("Press Enter with empty marker to finish configuration.")
	fmt.Println()

	// Create a temporary mappings map for interactive configuration
	mappings := make(map[string]string)

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

		// Validate marker (use the function from config.go)
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
		if existing, exists := mappings[marker]; exists {
			fmt.Printf("⚠️  Marker '%s' already exists with value: '%s'\n", marker, existing)
			fmt.Print("Overwrite? (y/N): ")
			confirm, _ := reader.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
				continue
			}
		}

		mappings[marker] = replacement
		fmt.Printf("✅ Added mapping: '%s' → '%s'\n", marker, replacement)
		fmt.Println()
	}

	// Store mappings in a temporary file for the setup script to use
	tmpMappingsFile := ".claude-temp-mappings.txt"
	if len(mappings) > 0 {
		// Write mappings to temporary file
		tmpFile, err := os.Create(tmpMappingsFile)
		if err != nil {
			return fmt.Errorf("failed to create temp mappings file: %w", err)
		}
		defer tmpFile.Close()
		// Don't remove the file here - let the setup script handle cleanup

		for marker, replacement := range mappings {
			_, err := tmpFile.WriteString(fmt.Sprintf("%s\t%s\n", marker, replacement))
			if err != nil {
				return fmt.Errorf("failed to write mappings: %w", err)
			}
		}
		
		fmt.Printf("📝 Total mappings configured: %d\n", len(mappings))
	} else {
		fmt.Println("No mappings configured. Default mappings will be used.")
	}

	return nil
}

