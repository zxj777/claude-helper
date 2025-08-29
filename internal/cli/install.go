package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"github.com/zxj777/claude-helper/internal/assets"
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
		err = installHook(componentName, templatePath, force)
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
	// Try to find agent template first
	if agentPath, err := assets.GetTemplatePath("agent", name); err == nil {
		return agentPath, "agent", nil
	}

	// Try to find hook template
	if hookPath, err := assets.GetTemplatePath("hook", name); err == nil {
		return hookPath, "hook", nil
	}

	return "", "", fmt.Errorf("template not found: %s", name)
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

func installHook(name, templatePath string, force bool) error {
	// Special handling for text-expander hook - configure mappings before setup
	if name == "text-expander" {
		if err := configureTextExpanderMappings(); err != nil {
			return fmt.Errorf("failed to configure text expander: %w", err)
		}
	}

	// Special handling for audio-notification hook - configure audio settings before setup  
	if name == "audio-notification" {
		if err := configureAudioNotificationSettings(); err != nil {
			return fmt.Errorf("failed to configure audio notification: %w", err)
		}
	}

	// Special handling for task-notification hook - configure notification settings before setup
	if name == "task-notification" {
		if err := configureTaskNotificationSettings(); err != nil {
			return fmt.Errorf("failed to configure task notification: %w", err)
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

	// Execute setup script if present (skip for hooks we handle with Go code)
	skipSetupHooks := []string{"text-expander", "audio-notification", "task-notification"}
	shouldSkip := false
	for _, skipHook := range skipSetupHooks {
		if name == skipHook {
			shouldSkip = true
			break
		}
	}

	if hook.Setup != "" && !shouldSkip {
		if err := executeSetupScript(hook.Setup); err != nil {
			return fmt.Errorf("failed to execute setup script: %w", err)
		}
	} else if shouldSkip {
		fmt.Printf("⏭️  Skipping setup script for %s (using Go configuration instead)\n", name)
	}


	// Copy associated Python/shell script files if they exist
	if err := copyHookScriptFiles(name, filepath.Dir(templatePath)); err != nil {
		return fmt.Errorf("failed to copy hook script files: %w", err)
	}

	// Ensure cross-platform run-python scripts exist
	if err := ensureCrossPlatformRunPythonScripts(); err != nil {
		return fmt.Errorf("failed to create cross-platform run-python scripts: %w", err)
	}

	// Special handling for text-expander - create Python script and config file
	if name == "text-expander" {
		if err := createTextExpanderPythonScript(); err != nil {
			return fmt.Errorf("failed to create text-expander Python script: %w", err)
		}
		if err := createTextExpanderConfig(); err != nil {
			return fmt.Errorf("failed to create text-expander config: %w", err)
		}
	}

	// Special handling for audio-notification - create Python script and config file
	if name == "audio-notification" {
		if err := createAudioNotificationPythonScript(); err != nil {
			return fmt.Errorf("failed to create audio-notification Python script: %w", err)
		}
		if err := createAudioNotificationConfig(); err != nil {
			return fmt.Errorf("failed to create audio-notification config: %w", err)
		}
		if err := copyAudioFiles(); err != nil {
			return fmt.Errorf("failed to copy audio files: %w", err)
		}
	}

	// Special handling for task-notification - create Python script and config file
	if name == "task-notification" {
		if err := createTaskNotificationPythonScript(); err != nil {
			return fmt.Errorf("failed to create task-notification Python script: %w", err)
		}
		if err := createTaskNotificationConfig(); err != nil {
			return fmt.Errorf("failed to create task-notification config: %w", err)
		}
		if err := copyAudioFiles(); err != nil {
			return fmt.Errorf("failed to copy audio files: %w", err)
		}
	}

	// Install hook to Claude settings
	return installHookToSettings(hook, force)
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


func copyHookScriptFiles(hookName, templateDir string) error {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	
	// Create .claude/hooks directory
	hooksDir := filepath.Join(wd, ".claude", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	// Look for associated script files (.py, .sh, .js, etc.)
	scriptExtensions := []string{".py", ".sh", ".js", ".ts"}
	
	for _, ext := range scriptExtensions {
		scriptName := hookName + ext
		sourcePath := filepath.Join(templateDir, scriptName)
		
		// Check if script file exists
		if _, err := os.Stat(sourcePath); err == nil {
			// Copy the script file
			targetPath := filepath.Join(hooksDir, scriptName)
			
			if err := copyFile(sourcePath, targetPath); err != nil {
				return fmt.Errorf("failed to copy script file %s: %w", scriptName, err)
			}
			
			// Make executable for shell scripts and Python scripts
			if ext == ".sh" || ext == ".py" {
				if err := os.Chmod(targetPath, 0755); err != nil {
					return fmt.Errorf("failed to make script executable: %w", err)
				}
			}
			
			fmt.Printf("Hook script copied to: %s\n", targetPath)
		}
	}
	
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Use io.Copy instead of WriteTo for compatibility
	if _, err := sourceFile.Seek(0, 0); err != nil {
		return err
	}
	
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	if _, err := destFile.Write(content); err != nil {
		return err
	}

	return destFile.Sync()
}

func installHookToSettings(hook *types.Hook, force bool) error {
	return config.InstallHookToSettings(hook, force)
}

func executeSetupScript(setupScript string) error {
	fmt.Println("🔧 Executing setup script...")
	
	var cmd *exec.Cmd
	var tmpFile *os.File
	var err error
	
	if runtime.GOOS == "windows" {
		// Convert bash script to PowerShell equivalent for Windows
		powershellScript := convertBashToPowerShell(setupScript)
		
		// Create a temporary PowerShell script file
		tmpFile, err = os.CreateTemp("", "claude-helper-setup-*.ps1")
		if err != nil {
			return fmt.Errorf("failed to create temp script file: %w", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()
		
		// Write the PowerShell script to the temp file
		if _, err := tmpFile.WriteString(powershellScript); err != nil {
			return fmt.Errorf("failed to write setup script: %w", err)
		}
		tmpFile.Close()
		
		// Find available PowerShell executable
		powerShellCmd := findPowerShellExecutable()
		if powerShellCmd == "" {
			return fmt.Errorf("PowerShell not found. Please install PowerShell or add it to PATH")
		}
		
		// Execute the PowerShell script
		cmd = exec.Command(powerShellCmd, "-ExecutionPolicy", "Bypass", "-File", tmpFile.Name())
	} else {
		// Unix/Linux/macOS - use bash
		tmpFile, err = os.CreateTemp("", "claude-helper-setup-*.sh")
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
		cmd = exec.Command("/bin/bash", tmpFile.Name())
	}
	
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
	// Get config path
	configPath, err := getTextExpanderConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Check if config already exists - if so, skip interactive configuration
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("🔧 Text Expander config already exists, skipping interactive configuration")
		return nil
	}

	fmt.Println("🔧 Configuring Text Expander mappings...")
	fmt.Println("You can create shortcuts that expand to longer text.")
	fmt.Println("Example: -d -> '详细解释这段代码的功能、实现原理和使用方法'")
	fmt.Println("Press Enter with empty marker to finish configuration.")
	fmt.Println()

	// Create new config with default mappings
	textConfig := &TextExpanderConfig{
		Mappings: map[string]string{
			"-d": "该睡觉了",
			"-z": "该睡觉了", 
			"-v": "查看详细信息",
			"-h": "显示帮助信息",
			"-l": "列出所有项目",
			"-s": "显示状态信息",
		},
		EscapeChar: "\\",
	}

	// Interactive input
	reader := bufio.NewReader(os.Stdin)
	added := 0
	
	for {
		fmt.Print("Enter marker (e.g., -d, -v, --explain): ")
		marker, err := reader.ReadString('\n')
		if err != nil {
			// If we can't read input (e.g., not in a terminal), use default config
			fmt.Println("\nNo interactive input available, using default configuration")
			break
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
		added++
		fmt.Println()
	}

	// Save config to JSON file directly (no more temp files)
	if err := saveTextExpanderConfig(configPath, textConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	
	if added > 0 {
		fmt.Printf("📝 Total new mappings added: %d\n", added)
	}
	fmt.Printf("📝 Total mappings configured: %d\n", len(textConfig.Mappings))

	return nil
}

func ensureCrossPlatformRunPythonScripts() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	
	hooksDir := filepath.Join(wd, ".claude", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	// Create run-python.bat for Windows
	batPath := filepath.Join(hooksDir, "run-python.bat")
	batContent := `@echo off
setlocal

REM Try different Python commands
python3 %* 2>nul
if %errorlevel% == 0 goto :eof

python %* 2>nul
if %errorlevel% == 0 goto :eof

py -3 %* 2>nul
if %errorlevel% == 0 goto :eof

py %* 2>nul
if %errorlevel% == 0 goto :eof

echo Python not found. Please install Python or add it to PATH.
exit /b 1
`
	
	if err := os.WriteFile(batPath, []byte(batContent), 0644); err != nil {
		return fmt.Errorf("failed to create run-python.bat: %w", err)
	}

	// Create run-python.sh for Unix-like systems
	shPath := filepath.Join(hooksDir, "run-python.sh")
	shContent := `#!/bin/bash

# Try different Python commands
if command -v python3 > /dev/null 2>&1; then
    python3 "$@"
elif command -v python > /dev/null 2>&1; then
    python "$@"
elif command -v py > /dev/null 2>&1; then
    py -3 "$@"
else
    echo "Python not found. Please install Python or add it to PATH." >&2
    exit 1
fi
`
	
	if err := os.WriteFile(shPath, []byte(shContent), 0755); err != nil {
		return fmt.Errorf("failed to create run-python.sh: %w", err)
	}

	fmt.Println("Cross-platform run-python scripts created successfully!")
	return nil
}

func createTextExpanderPythonScript() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	
	hooksDir := filepath.Join(wd, ".claude", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	pythonScriptPath := filepath.Join(hooksDir, "text-expander.py")
	
	// Python script content (extracted from the YAML setup script)
	pythonContent := `#!/usr/bin/env python3
import json
import sys
import os
import re

def apply_text_expansions_with_escape(text, mappings, escape_char='\\'):
    r"""Apply text expansions with support for escape characters.
    
    Rules:
    - \marker -> literal marker (no expansion)
    - \\marker -> literal \ + expand marker  
    - \\\marker -> literal \ + literal marker
    - \\\\marker -> literal \\ + expand marker
    """
    if not mappings:
        return text
    
    result = text
    
    # Process each mapping
    for marker, replacement in mappings.items():
        # Create a pattern that matches the marker with potential escaping
        # We need to handle sequences of backslashes before the marker
        pattern = r'(\\*)' + re.escape(marker)
        
        def replace_func(match):
            backslashes = match.group(1)
            backslash_count = len(backslashes)
            
            if backslash_count == 0:
                # No backslashes, normal expansion
                return replacement
            elif backslash_count % 2 == 1:
                # Odd number of backslashes: last one escapes the marker
                # Return half the backslashes (rounded down) + literal marker
                return '\\' * (backslash_count // 2) + marker
            else:
                # Even number of backslashes: marker is not escaped
                # Return half the backslashes + expanded marker
                return '\\' * (backslash_count // 2) + replacement
        
        result = re.sub(pattern, replace_func, result)
    
    return result

try:
    # Read JSON input from stdin
    input_data = json.load(sys.stdin)
    
    # Extract prompt
    prompt = input_data.get('prompt', '')
    
    # Clean the prompt to remove invalid Unicode characters
    try:
        # Encode and decode to clean up any invalid UTF-8 characters
        prompt = prompt.encode('utf-8', errors='ignore').decode('utf-8')
    except:
        # If that fails, use ascii encoding as fallback
        prompt = prompt.encode('ascii', errors='ignore').decode('ascii')
    
    if not prompt:
        sys.exit(0)
    
    # Load text expansion config from project .claude/config directory
    config_file = '.claude/config/text-expander.json'
    if not os.path.exists(config_file):
        sys.exit(0)
    
    with open(config_file, 'r', encoding='utf-8') as f:
        config = json.load(f)
    
    # Get escape character (default to backslash)
    escape_char = config.get('escape_char', '\\')
    mappings = config.get('mappings', {})
    
    # Apply text expansions with escape support
    expanded_prompt = apply_text_expansions_with_escape(prompt, mappings, escape_char)
    
    # If prompt changed, add expanded text as additional context
    if prompt != expanded_prompt:
        result = {
            "hookSpecificOutput": {
                "hookEventName": "UserPromptSubmit",
                "additionalContext": f"用户的意思是: {expanded_prompt}"
            }
        }
        print(json.dumps(result, ensure_ascii=True), flush=True)
        sys.exit(0)
    
    # No change needed, allow original prompt through
    sys.exit(0)
        
except Exception as e:
    # On any error, allow original prompt through
    # For debugging, could log error to a file
    try:
        with open('.claude/hook-error.log', 'a', encoding='utf-8', errors='replace') as f:
            import traceback
            f.write(f"Text-expander error: {type(e).__name__}: {str(e)}\\n")
            f.write(f"Traceback: {traceback.format_exc()}\\n")
    except:
        pass  # Ignore logging errors
    sys.exit(0)
`

	if err := os.WriteFile(pythonScriptPath, []byte(pythonContent), 0755); err != nil {
		return fmt.Errorf("failed to write Python script: %w", err)
	}

	fmt.Printf("Text expander Python script created at: %s\n", pythonScriptPath)
	return nil
}

func createTextExpanderConfig() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	
	configDir := filepath.Join(wd, ".claude", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "text-expander.json")
	
	// Check if config file already exists - don't overwrite user's interactive configuration!
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("Text expander config already exists, skipping default config creation")
		return nil
	}
	
	// Only create default config if file doesn't exist
	config := map[string]interface{}{
		"mappings": map[string]string{
			"-d": "该睡觉了",
			"-z": "该睡觉了",
			"-v": "查看详细信息",
			"-h": "显示帮助信息",
			"-l": "列出所有项目",
			"-s": "显示状态信息",
		},
		"escape_char": "\\",
	}
	
	// Use JSON encoder with proper UTF-8 handling
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false) // Don't escape HTML characters
	encoder.SetIndent("", "  ")  // Pretty print with 2-space indentation
	
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}
	
	if err := os.WriteFile(configPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	fmt.Println("Text expander config created with proper UTF-8 encoding!")
	return nil
}

// Note: Using shared functions from config.go in the same package

// findPowerShellExecutable tries to find an available PowerShell executable
func findPowerShellExecutable() string {
	// Try different PowerShell executables in order of preference
	powerShellCmds := []string{
		"pwsh",        // PowerShell 7+ (cross-platform)
		"powershell",  // Windows PowerShell 5.x
		"pwsh.exe",    // PowerShell 7+ with .exe extension
		"powershell.exe", // Windows PowerShell 5.x with .exe extension
	}
	
	for _, cmd := range powerShellCmds {
		if _, err := exec.LookPath(cmd); err == nil {
			return cmd
		}
	}
	
	return "" // No PowerShell executable found
}

func convertBashToPowerShell(bashScript string) string {
	// For Windows, we'll create a simplified PowerShell script
	// that mimics the bash script functionality
	return `# PowerShell setup for text-expander hook

# Create directories
New-Item -ItemType Directory -Force -Path ".claude\hooks" | Out-Null
New-Item -ItemType Directory -Force -Path ".claude\config" | Out-Null

# Create Python script content
$pythonContent = @'
#!/usr/bin/env python3
import json
import sys
import os
import re

def apply_text_expansions_with_escape(text, mappings, escape_char='\\'):
    if not mappings:
        return text
    result = text
    for marker, replacement in mappings.items():
        pattern = r'(\\*)' + re.escape(marker)
        def replace_func(match):
            backslashes = match.group(1)
            backslash_count = len(backslashes)
            if backslash_count == 0:
                return replacement
            elif backslash_count % 2 == 1:
                return '\\' * (backslash_count // 2) + marker
            else:
                return '\\' * (backslash_count // 2) + replacement
        result = re.sub(pattern, replace_func, result)
    return result

try:
    input_data = json.load(sys.stdin)
    prompt = input_data.get('prompt', '')
    if not prompt:
        sys.exit(0)
    config_file = '.claude/config/text-expander.json'
    if not os.path.exists(config_file):
        sys.exit(0)
    with open(config_file, 'r', encoding='utf-8') as f:
        config = json.load(f)
    escape_char = config.get('escape_char', '\\')
    mappings = config.get('mappings', {})
    expanded_prompt = apply_text_expansions_with_escape(prompt, mappings, escape_char)
    if prompt != expanded_prompt:
        print(f"用户的意思是: {expanded_prompt}")
        sys.exit(0)
    sys.exit(0)
except Exception as e:
    sys.exit(0)
'@

# Write Python script
$pythonContent | Out-File -FilePath ".claude\hooks\text-expander.py" -Encoding UTF8

# Handle mappings
if (Test-Path ".claude-temp-mappings.txt") {
    # Convert temp mappings to JSON format
    $mappings = @{}
    Get-Content ".claude-temp-mappings.txt" | ForEach-Object {
        $parts = $_ -split "\t", 2
        if ($parts.Length -eq 2) {
            $mappings[$parts[0]] = $parts[1]
        }
    }
    
    # Create config object
    $config = @{
        mappings = $mappings
        escape_char = "\\"
    }
    
    # Convert to JSON and save
    $config | ConvertTo-Json -Depth 2 | Out-File -FilePath ".claude\config\text-expander.json" -Encoding UTF8
    Remove-Item ".claude-temp-mappings.txt" -ErrorAction SilentlyContinue
} else {
    # Default config - create JSON manually to avoid encoding issues
    $jsonBuilder = [System.Text.StringBuilder]::new()
    [void]$jsonBuilder.AppendLine('{')
    [void]$jsonBuilder.AppendLine('  "mappings": {')
    [void]$jsonBuilder.AppendLine('    "-d": "该睡觉了",')
    [void]$jsonBuilder.AppendLine('    "-z": "该睡觉了",')
    [void]$jsonBuilder.AppendLine('    "-v": "查看详细信息",')
    [void]$jsonBuilder.AppendLine('    "-h": "显示帮助信息",')
    [void]$jsonBuilder.AppendLine('    "-l": "列出所有项目",')
    [void]$jsonBuilder.AppendLine('    "-s": "显示状态信息"')
    [void]$jsonBuilder.AppendLine('  },')
    [void]$jsonBuilder.AppendLine('  "escape_char": "\\"')
    [void]$jsonBuilder.AppendLine('}')
    
    # Write with explicit UTF-8 encoding
    $jsonContent = $jsonBuilder.ToString()
    [System.IO.File]::WriteAllText(".claude\config\text-expander.json", $jsonContent, [System.Text.Encoding]::UTF8)
}

Write-Host "Text expander hook installed successfully!"
`
}

// configureAudioNotificationSettings handles interactive audio notification configuration
func configureAudioNotificationSettings() error {
	fmt.Println("🔊 Configuring Audio Notification Settings...")
	fmt.Println("Choose when to play notification sounds after Claude completes tasks.")
	fmt.Println()

	// Check if config already exists
	configPath, err := getAudioNotificationConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("🔧 Audio notification config already exists, skipping interactive configuration")
		return nil
	}

	// Create default audio config
	audioConfig := &types.AudioConfig{
		Enabled:      true,
		SuccessSound: "success.wav",
		ErrorSound:   "error.wav", 
		DefaultSound: "complete.wav",
		Volume:       70,
	}

	// Interactive sound selection
	fmt.Println("Available notification sounds:")
	fmt.Println("1. success.wav    - 成功铃声 (愉悦的成功提示)")
	fmt.Println("2. complete.wav   - 完成提示 (中性的任务完成)")
	fmt.Println("3. subtle.wav     - 轻柔提醒 (不打扰工作)")
	fmt.Println("4. chime.wav      - 清脆铃声 (清脆悦耳)")
	fmt.Println("5. bell.wav       - 传统铃声 (经典铃铛声)")
	fmt.Println("6. attention.wav  - 注意提醒 (明显的提醒音)")
	fmt.Println("7. 禁用音频通知")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("请选择默认提示音 (1-7): ")
		choice, err := reader.ReadString('\n')
		if err != nil {
			// If we can't read input, use default settings
			fmt.Println("\n无法读取输入，使用默认配置")
			break
		}
		
		choice = strings.TrimSpace(choice)
		switch choice {
		case "1":
			audioConfig.DefaultSound = "success.wav"
			fmt.Println("✅ 选择了成功铃声作为默认提示音")
		case "2":
			audioConfig.DefaultSound = "complete.wav"
			fmt.Println("✅ 选择了完成提示作为默认提示音")
		case "3":
			audioConfig.DefaultSound = "subtle.wav"
			fmt.Println("✅ 选择了轻柔提醒作为默认提示音")
		case "4":
			audioConfig.DefaultSound = "chime.wav"
			fmt.Println("✅ 选择了清脆铃声作为默认提示音")
		case "5":
			audioConfig.DefaultSound = "bell.wav"
			fmt.Println("✅ 选择了传统铃声作为默认提示音")
		case "6":
			audioConfig.DefaultSound = "attention.wav"
			fmt.Println("✅ 选择了注意提醒作为默认提示音")
		case "7":
			audioConfig.Enabled = false
			fmt.Println("✅ 禁用了音频通知")
		default:
			fmt.Println("❌ 无效选择，请输入 1-7")
			continue
		}
		break
	}

	// Volume setting  
	fmt.Print("设置音量 (1-100，默认70): ")
	volumeStr, err := reader.ReadString('\n')
	if err == nil {
		volumeStr = strings.TrimSpace(volumeStr)
		if volumeStr != "" {
			if volume := parseVolume(volumeStr); volume > 0 {
				audioConfig.Volume = volume
				fmt.Printf("✅ 音量设置为: %d\n", volume)
			}
		}
	}

	// Save config
	if err := saveAudioNotificationConfig(configPath, audioConfig); err != nil {
		return fmt.Errorf("failed to save audio config: %w", err)
	}

	fmt.Printf("📝 音频通知配置已保存到: %s\n", configPath)
	return nil
}

func parseVolume(volumeStr string) int {
	var volume int
	if _, err := fmt.Sscanf(volumeStr, "%d", &volume); err == nil && volume >= 1 && volume <= 100 {
		return volume
	}
	return 0
}

func getAudioNotificationConfigPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return filepath.Join(wd, ".claude", "config", "audio-notification.json"), nil
}

func saveAudioNotificationConfig(configPath string, audioConfig *types.AudioConfig) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Use JSON encoder with proper formatting
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(audioConfig); err != nil {
		return fmt.Errorf("failed to encode audio config: %w", err)
	}

	if err := os.WriteFile(configPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write audio config file: %w", err)
	}

	return nil
}

func createAudioNotificationPythonScript() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	hooksDir := filepath.Join(wd, ".claude", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	pythonScriptPath := filepath.Join(hooksDir, "audio-notification.py")

	// Python script content from the YAML template
	pythonContent := `#!/usr/bin/env python3
import json
import sys
import os
import subprocess
import time
import platform

def get_audio_config():
    """Load audio configuration from config file"""
    config_file = '.claude/config/audio-notification.json'
    if not os.path.exists(config_file):
        return None
        
    try:
        with open(config_file, 'r', encoding='utf-8') as f:
            return json.load(f)
    except Exception as e:
        return None

def should_play_notification(config):
    """Check if we should play a notification based on cooldown"""
    if not config.get('enabled', True):
        return False
        
    cooldown = config.get('cooldown_seconds', 2)
    if cooldown <= 0:
        return True
        
    # Check last notification time
    last_file = '.claude/last-audio-notification'
    if os.path.exists(last_file):
        try:
            with open(last_file, 'r') as f:
                last_time = float(f.read().strip())
            if time.time() - last_time < cooldown:
                return False
        except:
            pass
            
    # Update last notification time
    try:
        with open(last_file, 'w') as f:
            f.write(str(time.time()))
    except:
        pass
        
    return True

def get_sound_file(config, tool_result):
    """Determine which sound file to play based on tool result"""
    # Try to determine if operation was successful
    success = True
    try:
        # Check for common error indicators
        if 'error' in str(tool_result).lower():
            success = False
        elif 'failed' in str(tool_result).lower():
            success = False
        elif 'exception' in str(tool_result).lower():
            success = False
    except:
        pass
    
    if success:
        return config.get('success_sound', config.get('default_sound', 'complete.wav'))
    else:
        return config.get('error_sound', config.get('default_sound', 'complete.wav'))

def get_sound_path(sound_file):
    """Get the full path to the sound file"""
    # First check if it's an absolute path
    if os.path.isabs(sound_file):
        return sound_file
        
    # Check in project .claude/sounds directory
    project_sound = os.path.join('.claude', 'sounds', sound_file)
    if os.path.exists(project_sound):
        return project_sound
        
    # Check in embedded sounds directory (relative to hook script)
    script_dir = os.path.dirname(os.path.abspath(__file__))
    embedded_sound = os.path.join(script_dir, '..', 'sounds', sound_file)
    if os.path.exists(embedded_sound):
        return embedded_sound
        
    return None

def play_audio_file(sound_path, volume=70):
    """Play audio file using platform-appropriate command"""
    if not sound_path or not os.path.exists(sound_path):
        return False
        
    try:
        system = platform.system().lower()
        
        if system == 'darwin':  # macOS
            subprocess.run(['afplay', sound_path], 
                         check=False, 
                         stdout=subprocess.DEVNULL, 
                         stderr=subprocess.DEVNULL)
        elif system == 'linux':
            # Try different Linux audio players
            players = ['aplay', 'paplay', 'play']
            for player in players:
                try:
                    subprocess.run([player, sound_path], 
                                 check=True,
                                 stdout=subprocess.DEVNULL, 
                                 stderr=subprocess.DEVNULL)
                    break
                except (subprocess.CalledProcessError, FileNotFoundError):
                    continue
        else:  # Windows (assume Git Bash environment)
            # Use PowerShell to play sound
            ps_command = f"(New-Object Media.SoundPlayer '{sound_path}').PlaySync()"
            subprocess.run(['powershell', '-c', ps_command], 
                         check=False,
                         stdout=subprocess.DEVNULL, 
                         stderr=subprocess.DEVNULL)
        return True
    except Exception as e:
        return False

def main():
    try:
        # Load configuration
        config = get_audio_config()
        if not config:
            sys.exit(0)  # No config, exit silently
            
        # Check if notifications should be played
        if not should_play_notification(config):
            sys.exit(0)
            
        # Read tool use data from stdin
        try:
            input_data = json.load(sys.stdin)
        except:
            sys.exit(0)
            
        # Determine appropriate sound
        sound_file = get_sound_file(config, input_data)
        sound_path = get_sound_path(sound_file)
        
        if sound_path:
            volume = config.get('volume', 70)
            play_audio_file(sound_path, volume)
            
    except Exception as e:
        # On any error, fail silently
        pass
    
    sys.exit(0)

if __name__ == '__main__':
    main()
`

	if err := os.WriteFile(pythonScriptPath, []byte(pythonContent), 0755); err != nil {
		return fmt.Errorf("failed to write Python script: %w", err)
	}

	fmt.Printf("Audio notification Python script created at: %s\n", pythonScriptPath)
	return nil
}

func createAudioNotificationConfig() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	configDir := filepath.Join(wd, ".claude", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "audio-notification.json")

	// Check if config already exists - don't overwrite user's configuration
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("Audio notification config already exists, skipping default config creation")
		return nil
	}

	// Create default config (this shouldn't happen if configureAudioNotificationSettings ran first)
	config := &types.AudioConfig{
		Enabled:      true,
		SuccessSound: "success.wav",
		ErrorSound:   "error.wav",
		DefaultSound: "complete.wav",
		Volume:       70,
	}

	return saveAudioNotificationConfig(configPath, config)
}

func copyAudioFiles() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Create sounds directory in project
	soundsDir := filepath.Join(wd, ".claude", "sounds")
	if err := os.MkdirAll(soundsDir, 0755); err != nil {
		return fmt.Errorf("failed to create sounds directory: %w", err)
	}

	// Get the source sounds directory from embedded assets
	// For now, just create a placeholder since we can't embed actual audio files
	soundFiles := []string{
		"success.wav", "error.wav", "complete.wav", 
		"attention.wav", "subtle.wav", "chime.wav", "bell.wav",
	}

	for _, soundFile := range soundFiles {
		targetPath := filepath.Join(soundsDir, soundFile)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			// Create empty placeholder files for now
			// In production, these would be actual audio files
			if err := os.WriteFile(targetPath, []byte(""), 0644); err != nil {
				return fmt.Errorf("failed to create sound file %s: %w", soundFile, err)
			}
		}
	}

	fmt.Printf("Audio files copied to: %s\n", soundsDir)
	fmt.Println("📝 Note: Run 'python3 generate-sounds.py' in the project root to generate actual audio files")
	
	return nil
}

// configureTaskNotificationSettings handles interactive task notification configuration
func configureTaskNotificationSettings() error {
	fmt.Println("🔔 Configuring Task Notification Settings...")
	fmt.Println("Choose how you want to be notified when Claude completes tasks.")
	fmt.Println()

	// Check if config already exists
	configPath, err := getTaskNotificationConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("🔧 Task notification config already exists, skipping interactive configuration")
		return nil
	}

	// Create default notification config
	notificationConfig := &types.NotificationConfig{
		NotificationTypes: []string{"desktop"}, // Default to desktop notifications
		CooldownSecs:      2,
		Desktop: types.DesktopConfig{
			Enabled:     true,
			ShowDetails: true,
		},
		Audio: types.AudioConfig{
			Enabled:      false,
			SuccessSound: "success.wav",
			ErrorSound:   "error.wav",
			DefaultSound: "complete.wav",
			Volume:       70,
		},
	}

	// Interactive notification type selection
	fmt.Println("选择任务完成提醒方式:")
	fmt.Println("1. 仅桌面通知 (推荐)")
	fmt.Println("2. 仅音频通知")
	fmt.Println("3. 桌面通知 + 音频通知")
	fmt.Println("4. 禁用通知")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("请选择 (1-4): ")
		choice, err := reader.ReadString('\n')
		if err != nil {
			// If we can't read input, use default settings
			fmt.Println("\n无法读取输入，使用默认配置 (桌面通知)")
			break
		}

		choice = strings.TrimSpace(choice)
		switch choice {
		case "1":
			notificationConfig.NotificationTypes = []string{"desktop"}
			notificationConfig.Desktop.Enabled = true
			notificationConfig.Audio.Enabled = false
			fmt.Println("✅ 选择了桌面通知")
		case "2":
			notificationConfig.NotificationTypes = []string{"audio"}
			notificationConfig.Desktop.Enabled = false
			notificationConfig.Audio.Enabled = true
			fmt.Println("✅ 选择了音频通知")
			// Ask for audio settings
			if err := configureAudioSettings(&notificationConfig.Audio, reader); err != nil {
				return fmt.Errorf("failed to configure audio settings: %w", err)
			}
		case "3":
			notificationConfig.NotificationTypes = []string{"desktop", "audio"}
			notificationConfig.Desktop.Enabled = true
			notificationConfig.Audio.Enabled = true
			fmt.Println("✅ 选择了桌面通知 + 音频通知")
			// Ask for audio settings
			if err := configureAudioSettings(&notificationConfig.Audio, reader); err != nil {
				return fmt.Errorf("failed to configure audio settings: %w", err)
			}
		case "4":
			notificationConfig.NotificationTypes = []string{}
			notificationConfig.Desktop.Enabled = false
			notificationConfig.Audio.Enabled = false
			fmt.Println("✅ 禁用了所有通知")
		default:
			fmt.Println("❌ 无效选择，请输入 1-4")
			continue
		}
		break
	}

	// Desktop notification details setting
	if notificationConfig.Desktop.Enabled {
		fmt.Print("显示详细信息? (Y/n): ")
		detailChoice, err := reader.ReadString('\n')
		if err == nil {
			detailChoice = strings.TrimSpace(strings.ToLower(detailChoice))
			if detailChoice == "n" || detailChoice == "no" {
				notificationConfig.Desktop.ShowDetails = false
				fmt.Println("✅ 禁用详细信息显示")
			} else {
				fmt.Println("✅ 启用详细信息显示")
			}
		}
	}

	// Save config
	if err := saveTaskNotificationConfig(configPath, notificationConfig); err != nil {
		return fmt.Errorf("failed to save notification config: %w", err)
	}

	fmt.Printf("📝 任务通知配置已保存到: %s\n", configPath)
	return nil
}

func configureAudioSettings(audioConfig *types.AudioConfig, reader *bufio.Reader) error {
	fmt.Println("\n音频设置:")
	fmt.Println("1. success.wav    - 成功铃声")
	fmt.Println("2. complete.wav   - 完成提示") 
	fmt.Println("3. subtle.wav     - 轻柔提醒")
	fmt.Println("4. chime.wav      - 清脆铃声")
	fmt.Println("5. bell.wav       - 传统铃声")

	fmt.Print("选择默认提示音 (1-5): ")
	soundChoice, err := reader.ReadString('\n')
	if err == nil {
		soundChoice = strings.TrimSpace(soundChoice)
		switch soundChoice {
		case "1":
			audioConfig.DefaultSound = "success.wav"
		case "2":
			audioConfig.DefaultSound = "complete.wav"
		case "3":
			audioConfig.DefaultSound = "subtle.wav"
		case "4":
			audioConfig.DefaultSound = "chime.wav"
		case "5":
			audioConfig.DefaultSound = "bell.wav"
		}
		fmt.Printf("✅ 选择了 %s 作为默认提示音\n", audioConfig.DefaultSound)
	}

	fmt.Print("设置音量 (1-100，默认70): ")
	volumeStr, err := reader.ReadString('\n')
	if err == nil {
		volumeStr = strings.TrimSpace(volumeStr)
		if volumeStr != "" {
			if volume := parseVolume(volumeStr); volume > 0 {
				audioConfig.Volume = volume
				fmt.Printf("✅ 音量设置为: %d\n", volume)
			}
		}
	}

	return nil
}

func getTaskNotificationConfigPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return filepath.Join(wd, ".claude", "config", "notification.json"), nil
}

func saveTaskNotificationConfig(configPath string, config *types.NotificationConfig) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Use JSON encoder with proper formatting
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode notification config: %w", err)
	}

	if err := os.WriteFile(configPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write notification config file: %w", err)
	}

	return nil
}

func createTaskNotificationPythonScript() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	hooksDir := filepath.Join(wd, ".claude", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	pythonScriptPath := filepath.Join(hooksDir, "task-notification.py")

	// The complete Python script with full notification logic
	pythonContent := `#!/usr/bin/env python3
import json
import sys
import os
import subprocess
import time
import platform

def get_notification_config():
    """Load notification configuration from config file"""
    config_file = '.claude/config/notification.json'
    if not os.path.exists(config_file):
        # Try legacy audio config
        legacy_config_file = '.claude/config/audio-notification.json'
        if os.path.exists(legacy_config_file):
            return migrate_legacy_config(legacy_config_file)
        return None
        
    try:
        with open(config_file, 'r', encoding='utf-8') as f:
            return json.load(f)
    except Exception as e:
        return None

def migrate_legacy_config(legacy_path):
    """Migrate legacy audio config to new notification config"""
    try:
        with open(legacy_path, 'r', encoding='utf-8') as f:
            legacy_config = json.load(f)
        
        # Convert to new format
        new_config = {
            "notification_types": ["audio"] if legacy_config.get("enabled", True) else [],
            "cooldown_seconds": legacy_config.get("cooldown_seconds", 2),
            "desktop": {
                "enabled": False,
                "show_details": True
            },
            "audio": {
                "enabled": legacy_config.get("enabled", True),
                "success_sound": legacy_config.get("success_sound", "success.wav"),
                "error_sound": legacy_config.get("error_sound", "error.wav"),
                "default_sound": legacy_config.get("default_sound", "complete.wav"),
                "volume": legacy_config.get("volume", 70)
            }
        }
        return new_config
    except Exception as e:
        return None

def should_send_notification(config):
    """Check if we should send a notification based on cooldown"""
    if not config.get('notification_types'):
        return False
        
    cooldown = config.get('cooldown_seconds', 2)
    if cooldown <= 0:
        return True
        
    # Check last notification time
    last_file = '.claude/last-notification-time'
    if os.path.exists(last_file):
        try:
            with open(last_file, 'r') as f:
                last_time = float(f.read().strip())
            if time.time() - last_time < cooldown:
                return False
        except:
            pass
            
    # Update last notification time
    try:
        with open(last_file, 'w') as f:
            f.write(str(time.time()))
    except:
        pass
        
    return True

def analyze_tool_result(tool_result):
    """Analyze tool result to determine success/failure and extract info"""
    success = True
    tool_name = "任务"
    details = ""
    
    try:
        result_str = str(tool_result).lower()
        
        # Check for error indicators
        error_keywords = ['error', 'failed', 'exception', 'timeout', 'denied', 'not found']
        for keyword in error_keywords:
            if keyword in result_str:
                success = False
                break
        
        # Try to extract tool information
        if isinstance(tool_result, dict):
            if 'tool_name' in tool_result:
                tool_name = tool_result['tool_name']
            elif 'command' in tool_result:
                tool_name = f"命令执行"
            elif 'file' in result_str:
                tool_name = f"文件操作"
        
    except Exception as e:
        pass
    
    return success, tool_name, details

def send_desktop_notification(title, message, message_type="info"):
    """Send desktop notification across platforms"""
    try:
        system = platform.system().lower()
        
        if system == 'darwin':  # macOS
            script = f'display notification "{message}" with title "{title}"'
            subprocess.run(['osascript', '-e', script], 
                         check=False, 
                         stdout=subprocess.DEVNULL, 
                         stderr=subprocess.DEVNULL)
        elif system == 'linux':
            # Try notify-send first
            try:
                subprocess.run(['notify-send', title, message], 
                             check=True,
                             stdout=subprocess.DEVNULL, 
                             stderr=subprocess.DEVNULL)
            except (subprocess.CalledProcessError, FileNotFoundError):
                # Try zenity as fallback
                try:
                    notification_text = f"{title}\\n{message}"
                    subprocess.run(['zenity', '--notification', f'--text={notification_text}'], 
                                 check=True,
                                 stdout=subprocess.DEVNULL, 
                                 stderr=subprocess.DEVNULL)
                except (subprocess.CalledProcessError, FileNotFoundError):
                    return False
        else:  # Windows
            # Use PowerShell balloon notification
            ps_script = f"""
                Add-Type -AssemblyName System.Windows.Forms
                $balloon = New-Object System.Windows.Forms.NotifyIcon
                $balloon.Icon = [System.Drawing.SystemIcons]::Information
                $balloon.BalloonTipTitle = "{title}"
                $balloon.BalloonTipText = "{message}"
                $balloon.Visible = $true
                $balloon.ShowBalloonTip(3000)
                Start-Sleep -Seconds 1
                $balloon.Dispose()
            """
            subprocess.run(['powershell', '-Command', ps_script], 
                         check=False,
                         stdout=subprocess.DEVNULL, 
                         stderr=subprocess.DEVNULL)
        return True
    except Exception as e:
        return False

def get_sound_file(config, success):
    """Get appropriate sound file based on result"""
    audio_config = config.get('audio', {})
    
    if success:
        return audio_config.get('success_sound', audio_config.get('default_sound', 'complete.wav'))
    else:
        return audio_config.get('error_sound', audio_config.get('default_sound', 'complete.wav'))

def get_sound_path(sound_file):
    """Get the full path to the sound file"""
    if os.path.isabs(sound_file):
        return sound_file if os.path.exists(sound_file) else None
        
    # Check in project .claude/sounds directory
    project_sound = os.path.join('.claude', 'sounds', sound_file)
    if os.path.exists(project_sound):
        return project_sound
        
    # Check relative to hooks directory
    script_dir = os.path.dirname(os.path.abspath(__file__))
    embedded_sound = os.path.join(script_dir, '..', 'sounds', sound_file)
    if os.path.exists(embedded_sound):
        return embedded_sound
        
    return None

def send_audio_notification(config, success):
    """Send audio notification"""
    audio_config = config.get('audio', {})
    if not audio_config.get('enabled', False):
        return False
    
    sound_file = get_sound_file(config, success)
    sound_path = get_sound_path(sound_file)
    
    if not sound_path:
        return False
    
    try:
        system = platform.system().lower()
        
        if system == 'darwin':  # macOS
            subprocess.run(['afplay', sound_path], 
                         check=False,
                         stdout=subprocess.DEVNULL, 
                         stderr=subprocess.DEVNULL)
        elif system == 'linux':
            # Try different audio players
            players = ['aplay', 'paplay', 'play']
            for player in players:
                try:
                    subprocess.run([player, sound_path], 
                                 check=True,
                                 stdout=subprocess.DEVNULL, 
                                 stderr=subprocess.DEVNULL)
                    break
                except (subprocess.CalledProcessError, FileNotFoundError):
                    continue
        else:  # Windows
            ps_script = f'$sound = New-Object Media.SoundPlayer "{sound_path}"; $sound.PlaySync()'
            subprocess.run(['powershell', '-Command', ps_script], 
                         check=False,
                         stdout=subprocess.DEVNULL, 
                         stderr=subprocess.DEVNULL)
        return True
    except Exception as e:
        return False

def main():
    try:
        # Load configuration
        config = get_notification_config()
        if not config:
            sys.exit(0)  # No config, exit silently
            
        # Check if notifications should be sent
        if not should_send_notification(config):
            sys.exit(0)
            
        # Read tool use data from stdin
        try:
            input_data = json.load(sys.stdin)
        except:
            sys.exit(0)
            
        # Analyze the tool result
        success, tool_name, details = analyze_tool_result(input_data)
        
        # Prepare notification content
        if success:
            title = "Claude Helper - 任务完成"
            message = f"✅ {tool_name} 操作完成"
            message_type = "success"
        else:
            title = "Claude Helper - 任务失败"  
            message = f"❌ {tool_name} 操作失败"
            message_type = "error"
        
        # Send notifications based on configured types
        notification_types = config.get('notification_types', [])
        
        # Send desktop notification
        if 'desktop' in notification_types:
            desktop_config = config.get('desktop', {})
            if desktop_config.get('enabled', False):
                send_desktop_notification(title, message, message_type)
        
        # Send audio notification
        if 'audio' in notification_types:
            send_audio_notification(config, success)
            
    except Exception as e:
        # On any error, fail silently
        try:
            with open('.claude/notification-error.log', 'a', encoding='utf-8', errors='replace') as f:
                import traceback
                f.write(f"Task notification error: {type(e).__name__}: {str(e)}\\n")
                f.write(f"Traceback: {traceback.format_exc()}\\n")
        except:
            pass  # Ignore logging errors
    
    sys.exit(0)

if __name__ == '__main__':
    main()
`

	if err := os.WriteFile(pythonScriptPath, []byte(pythonContent), 0755); err != nil {
		return fmt.Errorf("failed to write Python script: %w", err)
	}

	fmt.Printf("Task notification Python script created at: %s\n", pythonScriptPath)
	return nil
}

func createTaskNotificationConfig() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	configDir := filepath.Join(wd, ".claude", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "notification.json")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("Task notification config already exists, skipping default config creation")
		return nil
	}

	// Create default config
	config := &types.NotificationConfig{
		NotificationTypes: []string{"desktop"},
		CooldownSecs:      2,
		Desktop: types.DesktopConfig{
			Enabled:     true,
			ShowDetails: true,
		},
		Audio: types.AudioConfig{
			Enabled:      false,
			SuccessSound: "success.wav",
			ErrorSound:   "error.wav",
			DefaultSound: "complete.wav",
			Volume:       70,
		},
	}

	return saveTaskNotificationConfig(configPath, config)
}


