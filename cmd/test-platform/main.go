package main

import (
	"fmt"
	"runtime"
	"strings"
	"github.com/zxj777/claude-helper/pkg/types"
)

func main() {
	fmt.Printf("Testing on platform: %s\n", runtime.GOOS)
	fmt.Println(strings.Repeat("=", 50))
	
	// Test the exact command from text-expander template
	hook := types.Hook{
		Name:        "text-expander",
		Event:       types.UserPromptSubmit,
		Matcher:     "*",
		Command:     ".claude/hooks/run-python.sh .claude/hooks/text-expander.py",
		Timeout:     10,
		Enabled:     true,
	}

	fmt.Printf("Original command: %s\n", hook.Command)
	fmt.Printf("Platform command: %s\n", hook.GetPlatformCommand())
	
	// Test auto-review command
	hook2 := types.Hook{
		Name:        "auto-review", 
		Event:       types.UserPromptSubmit,
		Matcher:     "*review*",
		Command:     ".claude/hooks/run-python.sh .claude/hooks/auto-review.py \"$PROMPT\"",
		Timeout:     15,
		Enabled:     true,
	}

	fmt.Printf("\nAuto-review original: %s\n", hook2.Command)
	fmt.Printf("Auto-review platform: %s\n", hook2.GetPlatformCommand())

	// Test with a non run-python command
	hook3 := types.Hook{
		Command: "echo hello world",
		Enabled: true,
	}
	
	fmt.Printf("\nNon-python original: %s\n", hook3.Command)
	fmt.Printf("Non-python platform: %s\n", hook3.GetPlatformCommand())
}