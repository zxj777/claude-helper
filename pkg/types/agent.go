package types

import (
	"fmt"
	"strings"
	
	"gopkg.in/yaml.v3"
)

// Agent represents a Claude Code sub-agent configuration
type Agent struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Tools       []string `json:"tools,omitempty" yaml:"tools,omitempty"`
	Prompt      string   `json:"prompt" yaml:"prompt"`
	Enabled     bool     `json:"enabled" yaml:"enabled"`
}

// ToMarkdown converts the Agent to Claude Code's agent file format
func (a *Agent) ToMarkdown() string {
	var frontmatter strings.Builder
	frontmatter.WriteString("---\n")
	frontmatter.WriteString(fmt.Sprintf("name: %s\n", a.Name))
	frontmatter.WriteString(fmt.Sprintf("description: %s\n", a.Description))
	
	if len(a.Tools) > 0 {
		frontmatter.WriteString(fmt.Sprintf("tools: %s\n", strings.Join(a.Tools, ", ")))
	}
	
	frontmatter.WriteString("---\n\n")
	frontmatter.WriteString(a.Prompt)
	
	return frontmatter.String()
}

// AgentFrontmatter represents the YAML frontmatter in agent files
type AgentFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Tools       string `yaml:"tools,omitempty"` // Tools as comma-separated string
}

// ParseAgentFromMarkdown parses agent configuration from markdown content
func ParseAgentFromMarkdown(content string) (*Agent, error) {
	// Split frontmatter and content
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid agent format: missing frontmatter")
	}
	
	// Parse YAML frontmatter
	var frontmatter AgentFrontmatter
	yamlContent := strings.TrimSpace(parts[1])
	if err := yaml.Unmarshal([]byte(yamlContent), &frontmatter); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}
	
	// Parse tools from comma-separated string
	var tools []string
	if frontmatter.Tools != "" {
		toolsList := strings.Split(frontmatter.Tools, ",")
		for _, tool := range toolsList {
			tools = append(tools, strings.TrimSpace(tool))
		}
	}
	
	return &Agent{
		Name:        frontmatter.Name,
		Description: frontmatter.Description,
		Tools:       tools,
		Prompt:      strings.TrimSpace(parts[2]),
		Enabled:     true, // Default to enabled
	}, nil
}