package types

// HookEvent represents the available hook events in Claude Code
type HookEvent string

// Claude Code hook events
const (
	PreToolUse       HookEvent = "PreToolUse"
	PostToolUse      HookEvent = "PostToolUse"
	UserPromptSubmit HookEvent = "UserPromptSubmit"
	Notification     HookEvent = "Notification"
	Stop             HookEvent = "Stop"
	SubagentStop     HookEvent = "SubagentStop"
	PreCompact       HookEvent = "PreCompact"
	SessionStart     HookEvent = "SessionStart"
	SessionEnd       HookEvent = "SessionEnd"
)

// Hook represents a Claude Code hook configuration
type Hook struct {
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description" yaml:"description"`
	Event       HookEvent `json:"event" yaml:"event"`
	Matcher     string    `json:"matcher,omitempty" yaml:"matcher,omitempty"`
	Setup       string    `json:"setup,omitempty" yaml:"setup,omitempty"`
	Command     string    `json:"command" yaml:"command"`
	Timeout     int       `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Enabled     bool      `json:"enabled" yaml:"enabled"`
}

// ToClaudeHookEntry converts the Hook to a single hook entry format
func (h *Hook) ToClaudeHookEntry() map[string]interface{} {
	hook := map[string]interface{}{
		"type":    "command", 
		"command": h.Command,
	}
	
	if h.Timeout > 0 {
		hook["timeout"] = h.Timeout
	}
	
	return map[string]interface{}{
		"matcher": h.Matcher,
		"hooks":   []map[string]interface{}{hook},
	}
}

// MergeHooksIntoClaudeConfig merges multiple hooks into Claude's settings format
func MergeHooksIntoClaudeConfig(hooks []Hook) map[string]interface{} {
	claudeHooks := make(map[string][]map[string]interface{})
	
	// Group hooks by Event and Matcher
	type EventMatcher struct {
		Event   string
		Matcher string
	}
	grouped := make(map[EventMatcher][]map[string]interface{})
	
	for _, hook := range hooks {
		if !hook.Enabled {
			continue // Skip disabled hooks
		}
		
		key := EventMatcher{
			Event:   string(hook.Event),
			Matcher: hook.Matcher,
		}
		
		hookCmd := map[string]interface{}{
			"type":    "command",
			"command": hook.Command,
		}
		if hook.Timeout > 0 {
			hookCmd["timeout"] = hook.Timeout
		}
		
		grouped[key] = append(grouped[key], hookCmd)
	}
	
	// Convert grouped hooks to Claude format
	for key, hookCmds := range grouped {
		eventKey := key.Event
		if claudeHooks[eventKey] == nil {
			claudeHooks[eventKey] = []map[string]interface{}{}
		}
		
		entry := map[string]interface{}{
			"matcher": key.Matcher,
			"hooks":   hookCmds,
		}
		claudeHooks[eventKey] = append(claudeHooks[eventKey], entry)
	}
	
	return map[string]interface{}{
		"hooks": claudeHooks,
	}
}
