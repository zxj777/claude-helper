package main

import (
	"fmt"
	"log"
	"os"

	"github.com/zxj777/claude-helper/internal/assets"
)

func main() {
	// Test template directory discovery
	templatesDir, err := assets.GetTemplatesDir()
	if err != nil {
		log.Fatalf("Failed to get templates directory: %v", err)
	}

	fmt.Printf("Templates directory: %s\n", templatesDir)

	// Check if directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		log.Fatalf("Templates directory does not exist: %s", templatesDir)
	}

	// Test listing hook templates
	hooks, err := assets.ListHookTemplates()
	if err != nil {
		log.Fatalf("Failed to list hook templates: %v", err)
	}

	fmt.Printf("Available hooks: %v\n", hooks)

	// Test listing agent templates
	agents, err := assets.ListAgentTemplates()
	if err != nil {
		log.Fatalf("Failed to list agent templates: %v", err)
	}

	fmt.Printf("Available agents: %v\n", agents)

	// Test getting a specific template path
	if len(hooks) > 0 {
		templatePath, err := assets.GetTemplatePath("hook", hooks[0])
		if err != nil {
			log.Fatalf("Failed to get template path for %s: %v", hooks[0], err)
		}
		fmt.Printf("Template path for %s: %s\n", hooks[0], templatePath)

		// Verify file exists
		if _, err := os.Stat(templatePath); err != nil {
			log.Fatalf("Template file does not exist: %s", templatePath)
		}
	}

	fmt.Println("Template system test passed!")
}