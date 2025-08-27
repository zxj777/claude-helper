package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"github.com/zxj777/claude-helper/pkg/types"
)

func main() {
	fmt.Printf("Testing cross-platform command conversion on %s\n", runtime.GOOS)
	fmt.Println("====================================================")
	
	// Test hook with run-python command
	hook := types.Hook{
		Name:        "text-expander",
		Description: "Test hook for cross-platform commands",
		Event:       types.UserPromptSubmit,
		Matcher:     "*",
		Command:     ".claude/hooks/run-python.sh .claude/hooks/text-expander.py",
		Timeout:     10,
		Enabled:     true,
	}

	fmt.Printf("Original command: %s\n", hook.Command)
	fmt.Printf("Platform-specific command: %s\n", hook.GetPlatformCommand())

	// Test ToClaudeHookEntry
	entry := hook.ToClaudeHookEntry()
	entryJson, _ := json.MarshalIndent(entry, "", "  ")
	fmt.Printf("Hook entry JSON:\n%s\n", string(entryJson))

	// Test MergeHooksIntoClaudeConfig
	hooks := []types.Hook{hook}
	config := types.MergeHooksIntoClaudeConfig(hooks)
	configJson, _ := json.MarshalIndent(config, "", "  ")
	fmt.Printf("Claude config JSON:\n%s\n", string(configJson))
	
	fmt.Println("\n====================================================")
	
	// Test with different command
	hook2 := types.Hook{
		Name:        "auto-review",
		Description: "Test hook 2",
		Event:       types.UserPromptSubmit,
		Matcher:     "*review*",
		Command:     ".claude/hooks/run-python.sh .claude/hooks/auto-review.py \"$PROMPT\"",
		Timeout:     15,
		Enabled:     true,
	}

	fmt.Printf("Original command 2: %s\n", hook2.Command)
	fmt.Printf("Platform-specific command 2: %s\n", hook2.GetPlatformCommand())
}