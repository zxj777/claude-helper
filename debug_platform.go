package main

import (
	"fmt"
	"strings"
)

// Simulate the GetPlatformCommand logic for different platforms
func getPlatformCommandWindows(command string) string {
	if !containsRunPython(command) {
		return command
	}
	
	// Windows logic (Fixed)
	cmd := strings.Replace(command, ".claude/hooks/run-python.sh", ".claude\\hooks\\run-python.bat", -1)
	return cmd
}

func getPlatformCommandMac(command string) string {
	if !containsRunPython(command) {
		return command
	}
	
	// Mac/Linux logic (Fixed) 
	cmd := strings.Replace(command, ".claude\\hooks\\run-python.bat", ".claude/hooks/run-python.sh", -1)
	return cmd
}

func containsRunPython(command string) bool {
	return strings.Contains(command, "run-python.bat") || strings.Contains(command, "run-python.sh")
}

func main() {
	fmt.Println("=== Cross-Platform Command Testing ===")
	
	// Test with template command (Unix-style paths)
	templateCommand := ".claude/hooks/run-python.sh .claude/hooks/text-expander.py"
	fmt.Printf("Template command: %s\n", templateCommand)
	
	// Test Windows conversion
	windowsResult := getPlatformCommandWindows(templateCommand)
	fmt.Printf("Windows result:   %s\n", windowsResult)
	
	// Test Mac conversion (should keep original)
	macResult := getPlatformCommandMac(templateCommand)
	fmt.Printf("Mac result:       %s\n", macResult)
	
	fmt.Println("\n=== Reverse Test (Windows -> Mac) ===")
	
	// Test with Windows-style command
	windowsCommand := ".claude\\hooks\\run-python.bat .claude\\hooks\\text-expander.py"
	fmt.Printf("Windows command:  %s\n", windowsCommand)
	
	// Test Mac conversion from Windows format
	macFromWindows := getPlatformCommandMac(windowsCommand)
	fmt.Printf("Mac from Windows: %s\n", macFromWindows)
	
	// Test Windows conversion (should keep original)
	windowsFromWindows := getPlatformCommandWindows(windowsCommand)
	fmt.Printf("Windows from Win: %s\n", windowsFromWindows)
	
	fmt.Println("\n=== Auto-review Command Test ===")
	autoReviewTemplate := ".claude/hooks/run-python.sh .claude/hooks/auto-review.py \"$PROMPT\""
	fmt.Printf("Template:   %s\n", autoReviewTemplate)
	fmt.Printf("Windows:    %s\n", getPlatformCommandWindows(autoReviewTemplate))
	fmt.Printf("Mac:        %s\n", getPlatformCommandMac(autoReviewTemplate))
	
	fmt.Println("\n=== Non-Python Command Test ===")
	nonPythonCommand := "echo hello world"
	fmt.Printf("Template:   %s\n", nonPythonCommand)
	fmt.Printf("Windows:    %s\n", getPlatformCommandWindows(nonPythonCommand))
	fmt.Printf("Mac:        %s\n", getPlatformCommandMac(nonPythonCommand))
}