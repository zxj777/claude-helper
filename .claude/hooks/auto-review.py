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

def detect_project_info():
    """Analyze project structure and type"""
    project_info = {
        "type": "unknown",
        "languages": [],
        "has_git": False,
        "main_dirs": []
    }
    
    # Check for Git
    success, _ = run_command("git rev-parse --git-dir 2>/dev/null")
    project_info["has_git"] = success
    
    # Detect project type and languages
    cwd = Path.cwd()
    
    # Language detection
    extensions = {
        ".go": "Go", ".py": "Python", ".js": "JavaScript", ".ts": "TypeScript",
        ".java": "Java", ".cpp": "C++", ".c": "C", ".rs": "Rust",
        ".php": "PHP", ".rb": "Ruby", ".swift": "Swift", ".kt": "Kotlin"
    }
    
    found_extensions = set()
    for file in cwd.rglob("*"):
        if file.suffix in extensions and not any(part.startswith('.') for part in file.parts[1:]):
            found_extensions.add(file.suffix)
    
    project_info["languages"] = [extensions[ext] for ext in found_extensions]
    
    # Project type detection
    if (cwd / "go.mod").exists():
        project_info["type"] = "Go"
    elif (cwd / "package.json").exists():
        project_info["type"] = "Node.js"
    elif (cwd / "requirements.txt").exists() or (cwd / "pyproject.toml").exists():
        project_info["type"] = "Python"
    elif (cwd / "Cargo.toml").exists():
        project_info["type"] = "Rust"
    elif (cwd / "pom.xml").exists() or (cwd / "build.gradle").exists():
        project_info["type"] = "Java"
    
    # Important directories
    common_dirs = ["src", "internal", "pkg", "lib", "app", "components", "utils"]
    project_info["main_dirs"] = [d for d in common_dirs if (cwd / d).exists()]
    
    return project_info

def get_git_context():
    """Collect relevant Git information"""
    git_info = {
        "has_changes": False,
        "modified_files": [],
        "staged_files": [],
        "recent_commits": [],
        "current_branch": "unknown"
    }
    
    # Get current branch
    success, branch = run_command("git branch --show-current 2>/dev/null")
    if success:
        git_info["current_branch"] = branch
    
    # Get modified files
    success, output = run_command("git status --porcelain 2>/dev/null")
    if success and output:
        git_info["has_changes"] = True
        for line in output.split('\n'):
            if line.strip():
                status = line[:2]
                file_path = line[3:]
                if 'M' in status or 'A' in status:
                    if status[0] != ' ':  # Staged
                        git_info["staged_files"].append(file_path)
                    else:  # Modified
                        git_info["modified_files"].append(file_path)
    
    # Get recent commits
    success, output = run_command("git log --oneline -5 2>/dev/null")
    if success and output:
        git_info["recent_commits"] = output.split('\n')
    
    return git_info

def ensure_code_reviewer_agent():
    """Ensure code-reviewer agent is available"""
    claude_dir = Path(".claude")
    agents_dir = claude_dir / "agents"
    agent_file = agents_dir / "code-reviewer.md"
    
    # Create directories if needed
    agents_dir.mkdir(parents=True, exist_ok=True)
    
    if agent_file.exists():
        log("Code reviewer agent already installed")
        return True
    
    # Try to copy from templates
    template_sources = [
        Path("assets/templates/agents/code-reviewer.md"),
        Path("../assets/templates/agents/code-reviewer.md"),
        Path("./assets/templates/agents/code-reviewer.md")
    ]
    
    for template_path in template_sources:
        if template_path.exists():
            try:
                shutil.copy2(template_path, agent_file)
                log(f"✅ Installed code-reviewer agent from {template_path}")
                return True
            except Exception as e:
                log(f"Failed to copy from {template_path}: {e}", "WARN")
    
    # Create a minimal code reviewer agent if template not found
    minimal_agent = """---
name: code-reviewer
description: Expert code review specialist focused on quality, security, and best practices
tools: [Read, Grep, Glob, Bash]
---

You are an expert code reviewer with extensive experience across multiple programming languages and frameworks.

## Your Role
- Analyze code for quality, performance, security, and maintainability issues
- Provide constructive feedback with specific suggestions
- Focus on best practices and potential improvements
- Consider the broader context of the codebase

## Review Guidelines
1. **Code Quality**: Check for readability, maintainability, and adherence to conventions
2. **Security**: Identify potential security vulnerabilities or risks
3. **Performance**: Spot performance bottlenecks or inefficient patterns
4. **Best Practices**: Ensure code follows language and framework best practices
5. **Documentation**: Verify adequate commenting and documentation

## Output Format
For each issue found, provide:
- **Issue**: Brief description of the problem
- **Location**: File and line number if applicable  
- **Suggestion**: Specific recommendation for improvement
- **Priority**: High/Medium/Low based on impact

Focus on being helpful and educational rather than overly critical.
"""
    
    try:
        with open(agent_file, 'w', encoding='utf-8') as f:
            f.write(minimal_agent)
        log("✅ Created minimal code-reviewer agent")
        return True
    except Exception as e:
        log(f"❌ Failed to create code reviewer agent: {e}", "ERROR")
        return False

def create_review_context(prompt, project_info, git_info):
    """Create comprehensive review context file"""
    context_dir = Path(".claude/context")
    context_dir.mkdir(parents=True, exist_ok=True)
    
    context = {
        "trigger": "auto-review",
        "timestamp": datetime.now().isoformat(),
        "user_prompt": prompt,
        "project_info": project_info,
        "git_info": git_info,
        "review_suggestions": generate_review_suggestions(prompt, project_info, git_info)
    }
    
    context_file = context_dir / "review-context.json"
    try:
        with open(context_file, 'w', encoding='utf-8') as f:
            json.dump(context, f, indent=2, ensure_ascii=False)
        log(f"📝 Created review context: {context_file}")
        return True
    except Exception as e:
        log(f"❌ Failed to create context: {e}", "ERROR")
        return False

def generate_review_suggestions(prompt, project_info, git_info):
    """Generate intelligent review suggestions based on context"""
    suggestions = []
    
    # Analyze prompt for specific requests
    prompt_lower = prompt.lower()
    
    if any(keyword in prompt_lower for keyword in ["security", "安全", "vulnerability"]):
        suggestions.append("Focus on security vulnerabilities and potential exploits")
    
    if any(keyword in prompt_lower for keyword in ["performance", "性能", "optimization", "优化"]):
        suggestions.append("Analyze performance bottlenecks and optimization opportunities")
    
    if any(keyword in prompt_lower for keyword in ["refactor", "重构", "clean", "improve"]):
        suggestions.append("Identify refactoring opportunities and code structure improvements")
    
    # Project-specific suggestions
    if project_info["type"] == "Go":
        suggestions.append("Check for Go best practices: error handling, goroutine usage, interface design")
    elif project_info["type"] == "Python":
        suggestions.append("Review Python conventions: PEP 8, type hints, error handling")
    elif project_info["type"] == "Node.js":
        suggestions.append("Check JavaScript/TypeScript patterns: async/await usage, error handling, type safety")
    
    # Git-based suggestions
    if git_info["has_changes"]:
        if git_info["modified_files"]:
            suggestions.append(f"Focus review on recently modified files: {', '.join(git_info['modified_files'][:3])}")
        if len(git_info["modified_files"]) > 10:
            suggestions.append("Large changeset detected - consider breaking into smaller reviews")
    
    return suggestions

def main():
    """Main execution function"""
    prompt = sys.argv[1] if len(sys.argv) > 1 else ""
    
    log("🎯 Auto-review triggered")
    
    # Collect project context
    log("🔍 Analyzing project structure...")
    project_info = detect_project_info()
    
    log("📊 Collecting Git information...")
    git_info = get_git_context()
    
    # Ensure agent is ready
    log("🤖 Preparing code reviewer agent...")
    if not ensure_code_reviewer_agent():
        log("❌ Could not prepare code reviewer agent", "ERROR")
        return 1
    
    # Create review context
    log("📋 Creating review context...")
    if create_review_context(prompt, project_info, git_info):
        # Provide helpful output
        print("\n🎯 Code Review Ready!")
        print(f"📁 Project Type: {project_info['type']} ({', '.join(project_info['languages'])})")
        
        if git_info["has_changes"]:
            print(f"📝 Changes detected: {len(git_info['modified_files'])} modified, {len(git_info['staged_files'])} staged")
            
        if project_info["main_dirs"]:
            print(f"🗂️  Key directories: {', '.join(project_info['main_dirs'])}")
            
        print("\n💡 The code-reviewer agent is now active and ready to help!")
        print("   You can ask for specific reviews, security checks, or general code quality feedback.")
        
        return 0
    else:
        log("❌ Failed to create review context", "ERROR")
        return 1

if __name__ == "__main__":
    exit(main())