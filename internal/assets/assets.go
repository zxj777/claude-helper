package assets

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed templates/*
var templatesFS embed.FS

// GetTemplatesDir returns the path to the templates directory
// If running from source (development), it uses the local assets/templates
// If running from built binary, it extracts embedded files to a temp location
func GetTemplatesDir() (string, error) {
	// First try to find local templates directory (for development)
	if wd, err := os.Getwd(); err == nil {
		localTemplatesDir := filepath.Join(wd, "assets", "templates")
		if _, err := os.Stat(localTemplatesDir); err == nil {
			return localTemplatesDir, nil
		}
	}

	// If not found locally, extract embedded files to temp directory
	tempDir, err := os.MkdirTemp("", "claude-helper-templates-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Extract embedded templates to temp directory
	err = fs.WalkDir(templatesFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Create the target path in temp directory
		targetPath := filepath.Join(tempDir, path)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Read file content from embedded FS
		content, err := templatesFS.ReadFile(path)
		if err != nil {
			return err
		}

		// Create parent directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// Write file to temp location
		return os.WriteFile(targetPath, content, 0644)
	})

	if err != nil {
		os.RemoveAll(tempDir) // Clean up on error
		return "", fmt.Errorf("failed to extract embedded templates: %w", err)
	}

	return filepath.Join(tempDir, "templates"), nil
}

// ListAgentTemplates returns a list of available agent templates
func ListAgentTemplates() ([]string, error) {
	templatesDir, err := GetTemplatesDir()
	if err != nil {
		return nil, err
	}

	agentsDir := filepath.Join(templatesDir, "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read agents directory: %w", err)
	}

	var agents []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			name := entry.Name()[:len(entry.Name())-3] // Remove .md extension
			agents = append(agents, name)
		}
	}

	return agents, nil
}

// ListHookTemplates returns a list of available hook templates
func ListHookTemplates() ([]string, error) {
	templatesDir, err := GetTemplatesDir()
	if err != nil {
		return nil, err
	}

	hooksDir := filepath.Join(templatesDir, "hooks")
	entries, err := os.ReadDir(hooksDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read hooks directory: %w", err)
	}

	var hooks []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			name := entry.Name()[:len(entry.Name())-5] // Remove .yaml extension
			hooks = append(hooks, name)
		}
	}

	return hooks, nil
}

// GetTemplatePath returns the full path to a template file
func GetTemplatePath(templateType, name string) (string, error) {
	templatesDir, err := GetTemplatesDir()
	if err != nil {
		return "", err
	}

	var templatePath string
	switch templateType {
	case "agent":
		templatePath = filepath.Join(templatesDir, "agents", name+".md")
	case "hook":
		templatePath = filepath.Join(templatesDir, "hooks", name+".yaml")
	default:
		return "", fmt.Errorf("unknown template type: %s", templateType)
	}

	if _, err := os.Stat(templatePath); err != nil {
		return "", fmt.Errorf("template not found: %s", templatePath)
	}

	return templatePath, nil
}
