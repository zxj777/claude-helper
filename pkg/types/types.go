package types

// TextExpanderConfig represents the configuration for text expander
type TextExpanderConfig struct {
	Mappings map[string]string `json:"mappings"`
}

// TODO: Define other core data structures for agents, hooks, and templates as needed