package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure installed components",
	Long:  `Configure settings for installed components like text-expander.`,
}

var configTextExpanderCmd = &cobra.Command{
	Use:   "text-expander",
	Short: "Configure text expander mappings",
	Long:  `Add, remove, or list text expander mappings.`,
}

var addMappingCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new text expander mappings",
	Long:  `Interactively add new text expander mappings.`,
	RunE:  addTextExpanderMappings,
}

var listMappingsCmd = &cobra.Command{
	Use:   "list",
	Short: "List current text expander mappings",
	Long:  `Display all configured text expander mappings.`,
	RunE:  listTextExpanderMappings,
}

var removeMappingCmd = &cobra.Command{
	Use:   "remove <marker>",
	Short: "Remove a text expander mapping",
	Long:  `Remove a specific text expander mapping by its marker.`,
	Args:  cobra.ExactArgs(1),
	RunE:  removeTextExpanderMapping,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configTextExpanderCmd)
	configTextExpanderCmd.AddCommand(addMappingCmd)
	configTextExpanderCmd.AddCommand(listMappingsCmd)
	configTextExpanderCmd.AddCommand(removeMappingCmd)
}

func addTextExpanderMappings(cmd *cobra.Command, args []string) error {
	fmt.Println("🔧 Adding new Text Expander mappings...")
	fmt.Println("Press Enter with empty marker to finish.")
	fmt.Println()

	// Get config path
	configPath, err := getTextExpanderConfigPath()
	if err != nil {
		return err
	}

	// Load existing config
	textConfig, err := loadTextExpanderConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Interactive input
	reader := bufio.NewReader(os.Stdin)
	added := 0

	for {
		fmt.Print("Enter marker (e.g., -d, -v, --explain): ")
		marker, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		marker = strings.TrimSpace(marker)
		if marker == "" {
			break
		}

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

	if added > 0 {
		if err := saveTextExpanderConfig(configPath, textConfig); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("💾 Added %d new mappings\n", added)
	} else {
		fmt.Println("No new mappings added")
	}

	return nil
}

func listTextExpanderMappings(cmd *cobra.Command, args []string) error {
	configPath, err := getTextExpanderConfigPath()
	if err != nil {
		return err
	}

	textConfig, err := loadTextExpanderConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(textConfig.Mappings) == 0 {
		fmt.Println("No text expander mappings configured.")
		fmt.Println("Use 'claude-helper config text-expander add' to add mappings.")
		return nil
	}

	fmt.Printf("📝 Text Expander Mappings (%d total):\n\n", len(textConfig.Mappings))
	for marker, replacement := range textConfig.Mappings {
		fmt.Printf("  %s → %s\n", marker, replacement)
	}
	fmt.Printf("\nConfig file: %s\n", configPath)

	return nil
}

func removeTextExpanderMapping(cmd *cobra.Command, args []string) error {
	marker := args[0]

	configPath, err := getTextExpanderConfigPath()
	if err != nil {
		return err
	}

	textConfig, err := loadTextExpanderConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if _, exists := textConfig.Mappings[marker]; !exists {
		return fmt.Errorf("mapping for marker '%s' not found", marker)
	}

	delete(textConfig.Mappings, marker)

	if err := saveTextExpanderConfig(configPath, textConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Removed mapping for marker: '%s'\n", marker)
	return nil
}

func getTextExpanderConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".claude-helper", "text-expander-config.json"), nil
}

func loadTextExpanderConfig(configPath string) (*TextExpanderConfig, error) {
	var textConfig TextExpanderConfig

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, return empty config
		textConfig.Mappings = make(map[string]string)
		return &textConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, &textConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if textConfig.Mappings == nil {
		textConfig.Mappings = make(map[string]string)
	}

	return &textConfig, nil
}

// TextExpanderConfig represents the configuration for text expander
type TextExpanderConfig struct {
	Mappings map[string]string `json:"mappings"`
}