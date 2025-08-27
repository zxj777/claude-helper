package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/zxj777/claude-helper/internal/assets"
	"github.com/zxj777/claude-helper/internal/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates and installed components",
	Long:  `Display all available agent and hook templates, showing their status.`,
	RunE:  listComponents,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("agents", "a", false, "Show only agents")
	listCmd.Flags().BoolP("hooks", "k", false, "Show only hooks")
	listCmd.Flags().BoolP("installed", "i", false, "Show only installed components")
}

type Component struct {
	Name        string
	Type        string
	Description string
	Status      string
}

func listComponents(cmd *cobra.Command, args []string) error {
	// Get command flags
	showAgents, _ := cmd.Flags().GetBool("agents")
	showHooks, _ := cmd.Flags().GetBool("hooks")
	showInstalled, _ := cmd.Flags().GetBool("installed")

	// Get templates directory from assets
	templatesDir, err := assets.GetTemplatesDir()
	if err != nil {
		return fmt.Errorf("failed to get templates directory: %w", err)
	}

	fmt.Println("Scanning for templates...")

	var components []Component

	// Scan agents directory
	agentsDir := filepath.Join(templatesDir, "agents")
	if !showHooks { // Only scan agents if not specifically showing hooks
		err := filepath.Walk(agentsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Continue walking even if some paths fail
			}
			
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
				name := strings.TrimSuffix(info.Name(), ".md")
				
				// Check if agent is installed
				status := "Available"
				if installed, err := config.IsAgentInstalled(name); err == nil && installed {
					status = "Installed"
				}
				
				components = append(components, Component{
					Name:        name,
					Type:        "agent",
					Description: "Claude agent template", // TODO: Parse from file
					Status:      status,
				})
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Warning: failed to scan agents directory: %v\n", err)
		}
	}

	// Scan hooks directory  
	hooksDir := filepath.Join(templatesDir, "hooks")
	if !showAgents { // Only scan hooks if not specifically showing agents
		err := filepath.Walk(hooksDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Continue walking even if some paths fail
			}
			
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".yaml") {
				name := strings.TrimSuffix(info.Name(), ".yaml")
				
				// Check if hook is installed
				status := "Available"
				if installed, err := config.IsHookInstalled(name); err == nil && installed {
					status = "Installed"
				}
				
				components = append(components, Component{
					Name:        name,
					Type:        "hook", 
					Description: "Claude hook template", // TODO: Parse from file
					Status:      status,
				})
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Warning: failed to scan hooks directory: %v\n", err)
		}
	}

	// Filter by installed status if requested
	if showInstalled {
		var installedComponents []Component
		for _, comp := range components {
			if comp.Status == "Installed" {
				installedComponents = append(installedComponents, comp)
			}
		}
		components = installedComponents
	}

	// Display results in table format
	if len(components) == 0 {
		fmt.Println("No components found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tDESCRIPTION\tSTATUS")
	fmt.Fprintln(w, "----\t----\t-----------\t------")
	
	for _, comp := range components {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", 
			comp.Name, comp.Type, comp.Description, comp.Status)
	}
	
	w.Flush()
	return nil
}
