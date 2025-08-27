#!/usr/bin/env python3
"""
Intelligent Auto-Review Hook
Automatically prepares code review context when review-related requests are detected.
"""

import os
import sys
import json
import shutil
import subprocess
from pathlib import Path
from datetime import datetime

def log(message, level="INFO"):
    """Simple logging with timestamp"""
    timestamp = datetime.now().strftime("%H:%M:%S")
    print(f"🔍 [{timestamp}] {message}")

def run_command(cmd, capture_output=True):
    """Run shell command safely"""
    try:
        if capture_output:
            result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
            return result.returncode == 0, result.stdout.strip()
        else:
            result = subprocess.run(cmd, shell=True)
            return result.returncode == 0, ""
    except Exception as e:
        return False, str(e)

def check_git_repo():
    """Check if current directory is a git repository"""
    success, _ = run_command("git rev-parse --git-dir")
    return success

def get_recent_changes():
    """Get recent file changes from git"""
    if not check_git_repo():
        return []
    
    # Get files changed in last commit
    success, output = run_command("git diff --name-only HEAD~1 HEAD")
    if success and output:
        return output.split('\n')
    
    # If no commits, get staged files
    success, output = run_command("git diff --cached --name-only")
    if success and output:
        return output.split('\n')
        
    # If no staged files, get modified files
    success, output = run_command("git diff --name-only")
    if success and output:
        return output.split('\n')
    
    return []

def install_code_reviewer():
    """Install code-reviewer agent if not already installed"""
    
    # Check if code-reviewer agent already exists
    agent_path = Path(".claude/agents/code-reviewer.md")
    if agent_path.exists():
        log("Code reviewer agent already installed")
        return True
    
    # Try to copy from templates
    template_paths = [
        Path("assets/templates/agents/code-reviewer.md"),
        Path("../assets/templates/agents/code-reviewer.md"),
        Path("./assets/templates/agents/code-reviewer.md")
    ]
    
    for template_path in template_paths:
        if template_path.exists():
            # Create agents directory if it doesn't exist
            agent_path.parent.mkdir(parents=True, exist_ok=True)
            
            try:
                shutil.copy2(template_path, agent_path)
                log(f"✅ Code reviewer agent installed from {template_path}")
                return True
            except Exception as e:
                log(f"❌ Failed to copy template: {e}")
                continue
    
    # Try using claude-helper command if available
    success, _ = run_command("claude-helper install code-reviewer", capture_output=False)
    if success:
        log("✅ Code reviewer agent installed via claude-helper")
        return True
    else:
        log("❌ Failed to install code-reviewer agent")
        return False

def main():
    """Main hook execution"""
    if len(sys.argv) < 2:
        log("No prompt provided")
        sys.exit(0)
    
    prompt = sys.argv[1]
    
    # Check if prompt contains review-related keywords
    review_keywords = [
        'review', 'quality', 'bug', 'issue', 'problem', 'optimize', 'refactor',
        '质量', '问题', '优化', '重构', '审查', '检查'
    ]
    
    if not any(keyword.lower() in prompt.lower() for keyword in review_keywords):
        sys.exit(0)  # Not a review request, exit silently
    
    log("Review request detected, preparing context...")
    
    # Get recent changes for context
    changed_files = get_recent_changes()
    if changed_files:
        log(f"Recent changes detected in {len(changed_files)} files")
        
        # Filter for code files only
        code_extensions = {'.py', '.js', '.ts', '.go', '.java', '.cpp', '.c', '.rs', '.rb', '.php'}
        code_files = [f for f in changed_files if any(f.endswith(ext) for ext in code_extensions)]
        
        if code_files:
            log(f"Code files changed: {', '.join(code_files[:5])}")
    
    # Install code reviewer agent if needed
    install_code_reviewer()
    
    # Add context to the user prompt
    context = "\n\n🔍 **Auto-Review Context:**\n"
    if changed_files:
        context += f"Recent changes detected in: {', '.join(changed_files[:5])}\n"
    context += "Code reviewer agent is available for detailed analysis.\n"
    
    print(context)

if __name__ == "__main__":
    main()