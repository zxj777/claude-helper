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

	// Execute setup script if present (skip for text-expander as we handle it with Go code)
	if hook.Setup != "" && name != "text-expander" {
		if err := executeSetupScript(hook.Setup); err != nil {
			return fmt.Errorf("failed to execute setup script: %w", err)
		}
	} else if name == "text-expander" {
		fmt.Println("⏭️  Skipping setup script for text-expander (using Go configuration instead)")
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


