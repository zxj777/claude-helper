package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zxj777/claude-helper/pkg/types"
)

var createCmd = &cobra.Command{
	Use:   "create <type> <name>",
	Short: "Create a new component template",
	Long: `Create a new agent or hook template in the assets/templates directory.

Examples:
  claude-helper create agent my-reviewer
  claude-helper create hook my-formatter`,
	Args: cobra.ExactArgs(2),
	RunE: createComponent,
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP("description", "d", "", "Component description")
	createCmd.Flags().StringSliceP("tools", "t", []string{}, "Agent tools (comma-separated)")
}

func createComponent(cmd *cobra.Command, args []string) error {
	componentType := args[0]
	componentName := args[1]

	// Validate component type
	if componentType != "agent" && componentType != "hook" {
		return fmt.Errorf("invalid component type '%s'. Must be 'agent' or 'hook'", componentType)
	}

	// Validate component name
	if !isValidComponentName(componentName) {
		return fmt.Errorf("invalid component name '%s'. Use lowercase letters, numbers, and hyphens only", componentName)
	}

	description, _ := cmd.Flags().GetString("description")
	tools, _ := cmd.Flags().GetStringSlice("tools")

	fmt.Printf("Creating %s template: %s\n", componentType, componentName)

	switch componentType {
	case "agent":
		return createAgentTemplate(componentName, description, tools)
	case "hook":
		return createHookTemplate(componentName, description)
	default:
		return fmt.Errorf("unsupported component type: %s", componentType)
	}
}

func createAgentTemplate(name, description string, tools []string) error {
	// Get templates directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	
	agentsDir := filepath.Join(wd, "assets", "templates", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	agentPath := filepath.Join(agentsDir, name+".md")
	
	// Check if file already exists
	if _, err := os.Stat(agentPath); err == nil {
		return fmt.Errorf("agent template '%s' already exists", name)
	}

	// Set defaults if not provided
	if description == "" {
		description = fmt.Sprintf("A custom Claude agent: %s", name)
	}
	if len(tools) == 0 {
		tools = []string{"Read", "Write", "Bash"}
	}

	// Create agent structure
	agent := types.Agent{
		Name:        name,
		Description: description,
		Tools:       tools,
		Prompt:      generateAgentPromptTemplate(name, description),
		Enabled:     true,
	}

	// Convert to markdown
	content := agent.ToMarkdown()

	// Write to file
	if err := os.WriteFile(agentPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write agent template: %w", err)
	}

	fmt.Printf("✓ Created agent template: %s\n", agentPath)
	return nil
}

func createHookTemplate(name, description string) error {
	// Get templates directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	
	hooksDir := filepath.Join(wd, "assets", "templates", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	hookPath := filepath.Join(hooksDir, name+".yaml")
	
	// Check if file already exists
	if _, err := os.Stat(hookPath); err == nil {
		return fmt.Errorf("hook template '%s' already exists", name)
	}

	// Set defaults if not provided
	if description == "" {
		description = fmt.Sprintf("A custom Claude hook: %s", name)
	}

	// Create hook structure
	hook := types.Hook{
		Name:        name,
		Description: description,
		Event:       types.PostToolUse, // Default event
		Matcher:     "Edit|Write",      // Default matcher
		Command:     generateHookCommandTemplate(name),
		Timeout:     30,
		Enabled:     true,
	}

	// Convert to YAML
	content, err := generateHookYAMLTemplate(hook)
	if err != nil {
		return fmt.Errorf("failed to generate hook YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(hookPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write hook template: %w", err)
	}

	fmt.Printf("✓ Created hook template: %s\n", hookPath)
	return nil
}

func isValidComponentName(name string) bool {
	// Check if name contains only lowercase letters, numbers, and hyphens
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}
	return len(name) > 0 && !strings.HasPrefix(name, "-") && !strings.HasSuffix(name, "-")
}

func generateAgentPromptTemplate(name, description string) string {
	return fmt.Sprintf(`You are %s, %s.

## Your Role
Describe your specific role and responsibilities here.

## Guidelines
- Be helpful and accurate in your responses
- Follow best practices in your domain
- Provide clear explanations when needed

## Examples
Add specific examples of how you should behave or respond.`, strings.Title(strings.ReplaceAll(name, "-", " ")), description)
}

func generateHookCommandTemplate(name string) string {
	return fmt.Sprintf(`#!/bin/bash
# %s hook script
echo "Executing %s hook on file: $1"

# Add your custom logic here
# Example:
# if [[ "$1" == *.go ]]; then
#   echo "Processing Go file: $1"
# fi

echo "Hook completed successfully"`, name, name)
}

func generateHookYAMLTemplate(hook types.Hook) (string, error) {
	return fmt.Sprintf(`name: %s
description: %s
event: %s
matcher: %s
command: |
%s
timeout: %d
enabled: %t`, 
		hook.Name,
		hook.Description,
		hook.Event,
		hook.Matcher,
		indentString(hook.Command, "  "),
		hook.Timeout,
		hook.Enabled), nil
}

func indentString(s, indent string) string {
	lines := strings.Split(s, "\n")
	var indented []string
	for _, line := range lines {
		if line != "" {
			indented = append(indented, indent+line)
		} else {
			indented = append(indented, "")
		}
	}
	return strings.Join(indented, "\n")
}