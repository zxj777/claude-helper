#!/usr/bin/env python3
"""
Commit Helper Hook
Intelligently analyzes Git changes and suggests commit messages.
"""

import os
import sys
import re
import subprocess
from pathlib import Path
from datetime import datetime
from collections import defaultdict

def log(message, level="INFO"):
    """Simple logging"""
    timestamp = datetime.now().strftime("%H:%M:%S")
    symbols = {"INFO": "💬", "WARN": "⚠️", "ERROR": "❌", "SUCCESS": "✅"}
    symbol = symbols.get(level, "ℹ️")
    print(f"{symbol} [{timestamp}] {message}")

def run_command(cmd):
    """Execute git command safely"""
    try:
        result = subprocess.run(
            cmd, shell=True, capture_output=True, text=True, timeout=10
        )
        return result.returncode == 0, result.stdout.strip(), result.stderr.strip()
    except subprocess.TimeoutExpired:
        return False, "", "Command timed out"
    except Exception as e:
        return False, "", str(e)

class GitAnalyzer:
    """Analyzes Git repository state and changes"""
    
    def __init__(self):
        self.is_git_repo = self._check_git_repo()
        self.staged_files = []
        self.modified_files = []
        self.added_files = []
        self.deleted_files = []
        self.renamed_files = []
        self.diff_stats = {}
        
        if self.is_git_repo:
            self._analyze_status()
            self._analyze_diff()
    
    def _check_git_repo(self):
        """Check if current directory is a git repository"""
        success, _, _ = run_command("git rev-parse --git-dir")
        return success
    
    def _analyze_status(self):
        """Parse git status to categorize changes"""
        success, output, _ = run_command("git status --porcelain")
        if not success or not output:
            return
        
        for line in output.split('\n'):
            if not line.strip():
                continue
            
            status = line[:2]
            file_path = line[3:]
            
            # Check if file is staged (first character)
            if status[0] in 'MARC':
                self.staged_files.append(file_path)
                
                if status[0] == 'M':
                    self.modified_files.append(file_path)
                elif status[0] == 'A':
                    self.added_files.append(file_path)
                elif status[0] == 'D':
                    self.deleted_files.append(file_path)
                elif status[0] == 'R':
                    self.renamed_files.append(file_path)
    
    def _analyze_diff(self):
        """Analyze staged changes diff"""
        if not self.staged_files:
            return
        
        success, output, _ = run_command("git diff --cached --stat")
        if success and output:
            self.diff_stats['summary'] = output
        
        # Get detailed diff for analysis
        success, output, _ = run_command("git diff --cached")
        if success and output:
            self.diff_stats['detailed'] = output
    
    def get_change_summary(self):
        """Generate a summary of changes"""
        if not self.staged_files:
            return "No staged changes"
        
        summary = []
        if self.added_files:
            summary.append(f"{len(self.added_files)} added")
        if self.modified_files:
            summary.append(f"{len(self.modified_files)} modified")
        if self.deleted_files:
            summary.append(f"{len(self.deleted_files)} deleted")
        if self.renamed_files:
            summary.append(f"{len(self.renamed_files)} renamed")
        
        return f"{len(self.staged_files)} files: " + ", ".join(summary)
    
    def analyze_change_type(self):
        """Determine the primary type of changes"""
        if not self.staged_files:
            return "misc"
        
        file_extensions = defaultdict(int)
        change_patterns = {
            'feat': 0, 'fix': 0, 'docs': 0, 'style': 0, 'refactor': 0, 
            'test': 0, 'chore': 0, 'config': 0
        }
        
        # Analyze files
        for file_path in self.staged_files:
            path = Path(file_path)
            ext = path.suffix.lower()
            file_extensions[ext] += 1
            
            # Pattern matching for change types
            name_lower = path.name.lower()
            
            if any(x in name_lower for x in ['test', 'spec', '.test.', '_test']):
                change_patterns['test'] += 1
            elif ext in ['.md', '.txt', '.rst']:
                change_patterns['docs'] += 1
            elif ext in ['.json', '.yaml', '.yml', '.toml', '.ini', '.conf']:
                change_patterns['config'] += 1
            elif 'readme' in name_lower:
                change_patterns['docs'] += 1
        
        # Analyze diff content for more patterns
        if 'detailed' in self.diff_stats:
            diff_content = self.diff_stats['detailed'].lower()
            
            # Look for bug fix patterns
            if any(word in diff_content for word in ['fix', 'bug', 'error', 'issue', 'problem']):
                change_patterns['fix'] += 2
            
            # Look for new feature patterns
            if any(word in diff_content for word in ['add', 'new', 'feature', 'implement']):
                change_patterns['feat'] += 2
            
            # Look for refactoring patterns
            if any(word in diff_content for word in ['refactor', 'clean', 'optimize', 'improve']):
                change_patterns['refactor'] += 1
            
            # Look for style changes
            if any(word in diff_content for word in ['format', 'style', 'indent', 'whitespace']):
                change_patterns['style'] += 1
        
        # Determine primary change type
        max_score = max(change_patterns.values())
        if max_score == 0:
            return 'chore'
        
        return max(change_patterns.keys(), key=change_patterns.get)

class CommitMessageGenerator:
    """Generates intelligent commit message suggestions"""
    
    def __init__(self, git_analyzer):
        self.git = git_analyzer
        self.change_type = git_analyzer.analyze_change_type()
    
    def generate_conventional_commit(self):
        """Generate conventional commit format message"""
        if not self.git.staged_files:
            return "chore: update files"
        
        # Get scope if possible
        scope = self._detect_scope()
        scope_str = f"({scope})" if scope else ""
        
        # Generate description based on changes
        description = self._generate_description()
        
        return f"{self.change_type}{scope_str}: {description}"
    
    def generate_simple_commit(self):
        """Generate simple, direct commit message"""
        if not self.git.staged_files:
            return "Update files"
        
        description = self._generate_description()
        return description.capitalize()
    
    def generate_detailed_commit(self):
        """Generate detailed commit message with body"""
        header = self.generate_conventional_commit()
        
        body_parts = []
        
        if self.git.added_files:
            body_parts.append(f"Added: {', '.join(self.git.added_files[:3])}")
        if self.git.modified_files:
            body_parts.append(f"Modified: {', '.join(self.git.modified_files[:3])}")
        if self.git.deleted_files:
            body_parts.append(f"Deleted: {', '.join(self.git.deleted_files[:3])}")
        
        if len(self.git.staged_files) > 6:
            body_parts.append(f"...and {len(self.git.staged_files) - 6} more files")
        
        if body_parts:
            body = "\\n\\n" + "\\n".join(body_parts)
            return header + body
        
        return header
    
    def _detect_scope(self):
        """Detect scope from file paths"""
        if not self.git.staged_files:
            return None
        
        # Look for common directory patterns
        common_scopes = {}
        for file_path in self.git.staged_files:
            parts = Path(file_path).parts
            if len(parts) > 1:
                first_dir = parts[0]
                if first_dir in ['src', 'internal', 'pkg', 'lib', 'app', 'components']:
                    if len(parts) > 2:
                        scope = parts[1]
                    else:
                        scope = first_dir
                else:
                    scope = first_dir
                
                common_scopes[scope] = common_scopes.get(scope, 0) + 1
        
        if common_scopes:
            # Return most common scope
            return max(common_scopes.keys(), key=common_scopes.get)
        
        return None
    
    def _generate_description(self):
        """Generate description based on file analysis"""
        if not self.git.staged_files:
            return "update files"
        
        # File-based descriptions
        if len(self.git.staged_files) == 1:
            file_path = Path(self.git.staged_files[0])
            
            if self.git.added_files and file_path.name in self.git.added_files:
                return f"add {file_path.name}"
            elif self.git.deleted_files and file_path.name in self.git.deleted_files:
                return f"remove {file_path.name}"
            else:
                return f"update {file_path.name}"
        
        # Multiple files - use change type
        descriptions = {
            'feat': 'add new features',
            'fix': 'fix bugs and issues',
            'docs': 'update documentation',
            'style': 'improve code formatting',
            'refactor': 'refactor code structure',
            'test': 'update tests',
            'chore': 'update configuration and maintenance',
            'config': 'update configuration files'
        }
        
        base_desc = descriptions.get(self.change_type, 'update files')
        
        # Add context if possible
        if len(self.git.staged_files) <= 3:
            file_names = [Path(f).name for f in self.git.staged_files]
            return f"{base_desc} in {', '.join(file_names)}"
        
        return base_desc

def create_commit_context(suggestions, git_analyzer):
    """Create context file with commit suggestions"""
    context_dir = Path(".claude/context")
    context_dir.mkdir(parents=True, exist_ok=True)
    
    context = {
        "trigger": "commit-helper",
        "timestamp": datetime.now().isoformat(),
        "suggestions": suggestions,
        "changes": {
            "staged_files": git_analyzer.staged_files,
            "summary": git_analyzer.get_change_summary(),
            "change_type": git_analyzer.analyze_change_type()
        }
    }
    
    context_file = context_dir / "commit-context.json"
    try:
        import json
        with open(context_file, 'w', encoding='utf-8') as f:
            json.dump(context, f, indent=2, ensure_ascii=False)
        return True
    except Exception as e:
        log(f"Could not save context: {e}", "WARN")
        return False

def main():
    """Main execution function"""
    prompt = sys.argv[1] if len(sys.argv) > 1 else ""
    
    log("📝 Commit Helper activated")
    
    # Analyze git repository
    git_analyzer = GitAnalyzer()
    
    if not git_analyzer.is_git_repo:
        log("Not a Git repository", "WARN")
        return 0
    
    if not git_analyzer.staged_files:
        print("\\n⚠️  No staged changes found!")
        print("💡 Use 'git add <files>' to stage changes before committing")
        
        # Check if there are unstaged changes
        success, output, _ = run_command("git status --porcelain")
        if success and output:
            print("\\n📋 Unstaged changes detected:")
            for line in output.split('\\n')[:5]:  # Show first 5
                if line.strip():
                    print(f"   {line}")
            if len(output.split('\\n')) > 5:
                print("   ...")
        return 0
    
    # Generate commit message suggestions
    generator = CommitMessageGenerator(git_analyzer)
    
    suggestions = {
        "conventional": generator.generate_conventional_commit(),
        "simple": generator.generate_simple_commit(),
        "detailed": generator.generate_detailed_commit()
    }
    
    # Create context for potential AI assistance
    create_commit_context(suggestions, git_analyzer)
    
    # Display results
    print(f"\\n🎯 Commit Message Suggestions")
    print(f"📊 Changes: {git_analyzer.get_change_summary()}")
    print(f"🏷️  Type: {git_analyzer.analyze_change_type()}")
    
    print("\\n📝 Suggested Messages:")
    print(f"\\n1. **Conventional Commits:**")
    print(f"   {suggestions['conventional']}")
    
    print(f"\\n2. **Simple & Direct:**")
    print(f"   {suggestions['simple']}")
    
    print(f"\\n3. **Detailed:**")
    print(f"   {suggestions['detailed']}")
    
    if 'summary' in git_analyzer.diff_stats:
        print(f"\\n📈 **Change Summary:**")
        for line in git_analyzer.diff_stats['summary'].split('\\n')[:3]:
            if line.strip():
                print(f"   {line}")
    
    print(f"\\n💡 **To commit:** `git commit -m \"<your chosen message>\"`")
    print(f"🔄 **To modify:** Edit the message before committing")
    
    return 0

if __name__ == "__main__":
    exit(main())