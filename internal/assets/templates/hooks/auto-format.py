#!/usr/bin/env python3
"""
Auto-Format Hook
Automatically formats code files using appropriate language-specific formatters.
"""

import os
import sys
import json
import subprocess
from pathlib import Path
from datetime import datetime

def log(message, level="INFO"):
    """Simple logging"""
    timestamp = datetime.now().strftime("%H:%M:%S")
    symbols = {"INFO": "✨", "WARN": "⚠️", "ERROR": "❌", "SUCCESS": "✅"}
    symbol = symbols.get(level, "ℹ️")
    print(f"{symbol} [{timestamp}] {message}")

def run_command(cmd, cwd=None):
    """Execute command and return success status and output"""
    try:
        result = subprocess.run(
            cmd, shell=True, cwd=cwd, 
            capture_output=True, text=True, timeout=30
        )
        return result.returncode == 0, result.stdout, result.stderr
    except subprocess.TimeoutExpired:
        return False, "", "Command timed out"
    except Exception as e:
        return False, "", str(e)

def check_tool_available(tool):
    """Check if a formatting tool is available in PATH"""
    success, _, _ = run_command(f"which {tool} >/dev/null 2>&1")
    return success

def format_file(file_path):
    """Format a single file based on its extension"""
    path = Path(file_path)
    
    if not path.exists():
        log(f"File not found: {file_path}", "WARN")
        return False
    
    extension = path.suffix.lower()
    formatted = False
    
    # Python files
    if extension == '.py':
        if check_tool_available('black'):
            success, stdout, stderr = run_command(f'black "{file_path}"')
            if success:
                log(f"Formatted Python file with black: {file_path}", "SUCCESS")
                formatted = True
            else:
                log(f"Black failed: {stderr}", "WARN")
        
        if not formatted and check_tool_available('autopep8'):
            success, stdout, stderr = run_command(f'autopep8 -i "{file_path}"')
            if success:
                log(f"Formatted Python file with autopep8: {file_path}", "SUCCESS")
                formatted = True
    
    # Go files
    elif extension == '.go':
        if check_tool_available('gofmt'):
            success, stdout, stderr = run_command(f'gofmt -w "{file_path}"')
            if success:
                log(f"Formatted Go file: {file_path}", "SUCCESS")
                formatted = True
    
    # JavaScript/TypeScript files
    elif extension in ['.js', '.ts', '.jsx', '.tsx']:
        if check_tool_available('prettier'):
            success, stdout, stderr = run_command(f'prettier --write "{file_path}"')
            if success:
                log(f"Formatted JS/TS file with Prettier: {file_path}", "SUCCESS")
                formatted = True
    
    # Rust files
    elif extension == '.rs':
        if check_tool_available('rustfmt'):
            success, stdout, stderr = run_command(f'rustfmt "{file_path}"')
            if success:
                log(f"Formatted Rust file: {file_path}", "SUCCESS")
                formatted = True
    
    if not formatted:
        log(f"No formatter available for: {file_path}", "INFO")
    
    return formatted

def main():
    """Main hook execution"""
    if len(sys.argv) < 2:
        log("No file path provided", "ERROR")
        sys.exit(1)
    
    file_path = sys.argv[1]
    
    # Skip non-code files
    code_extensions = {'.py', '.go', '.js', '.ts', '.jsx', '.tsx', '.rs', '.java', '.cpp', '.c', '.h'}
    path = Path(file_path)
    
    if path.suffix.lower() not in code_extensions:
        sys.exit(0)  # Exit silently for non-code files
    
    log(f"Formatting file: {file_path}")
    format_file(file_path)

if __name__ == "__main__":
    main()