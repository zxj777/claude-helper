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

def get_file_info(file_path):
    """Analyze file for formatting decisions"""
    path = Path(file_path)
    
    if not path.exists():
        return None
    
    try:
        stat = path.stat()
        size_mb = stat.st_size / (1024 * 1024)
        
        return {
            "path": path,
            "name": path.name,
            "suffix": path.suffix.lower(),
            "size_mb": size_mb,
            "is_binary": is_binary_file(path)
        }
    except Exception:
        return None

def is_binary_file(file_path):
    """Check if file is binary"""
    try:
        with open(file_path, 'rb') as f:
            chunk = f.read(8192)
            return b'\x00' in chunk
    except:
        return True

class CodeFormatter:
    """Multi-language code formatter"""
    
    FORMATTERS = {
        # Go
        '.go': [
            {'tool': 'goimports', 'cmd': 'goimports -w "{file}"', 'desc': 'Go imports & format'},
            {'tool': 'gofmt', 'cmd': 'gofmt -w "{file}"', 'desc': 'Go format'}
        ],
        
        # Python
        '.py': [
            {'tool': 'black', 'cmd': 'black -q "{file}"', 'desc': 'Python Black formatter'},
            {'tool': 'autopep8', 'cmd': 'autopep8 -i "{file}"', 'desc': 'Python autopep8'},
        ],
        
        # JavaScript/TypeScript
        '.js': [
            {'tool': 'prettier', 'cmd': 'prettier --write "{file}"', 'desc': 'Prettier JS'},
        ],
        '.ts': [
            {'tool': 'prettier', 'cmd': 'prettier --write "{file}"', 'desc': 'Prettier TS'},
        ],
        '.jsx': [
            {'tool': 'prettier', 'cmd': 'prettier --write "{file}"', 'desc': 'Prettier JSX'},
        ],
        '.tsx': [
            {'tool': 'prettier', 'cmd': 'prettier --write "{file}"', 'desc': 'Prettier TSX'},
        ],
        
        # JSON
        '.json': [
            {'tool': 'jq', 'cmd': 'jq . "{file}" > "{file}.tmp" && mv "{file}.tmp" "{file}"', 'desc': 'JSON format'},
        ],
        
        # Rust
        '.rs': [
            {'tool': 'rustfmt', 'cmd': 'rustfmt "{file}"', 'desc': 'Rust format'},
        ],
        
        # Shell
        '.sh': [
            {'tool': 'shfmt', 'cmd': 'shfmt -w "{file}"', 'desc': 'Shell format'},
        ],
        
        # CSS
        '.css': [
            {'tool': 'prettier', 'cmd': 'prettier --write "{file}"', 'desc': 'Prettier CSS'},
        ],
        
        # HTML
        '.html': [
            {'tool': 'prettier', 'cmd': 'prettier --write "{file}"', 'desc': 'Prettier HTML'},
        ],
        
        # YAML
        '.yaml': [
            {'tool': 'prettier', 'cmd': 'prettier --write "{file}"', 'desc': 'Prettier YAML'},
        ],
        '.yml': [
            {'tool': 'prettier', 'cmd': 'prettier --write "{file}"', 'desc': 'Prettier YAML'},
        ],
        
        # Markdown
        '.md': [
            {'tool': 'prettier', 'cmd': 'prettier --write "{file}"', 'desc': 'Prettier Markdown'},
        ],
    }
    
    def __init__(self, file_info):
        self.file_info = file_info
        self.available_tools = self._check_available_tools()
    
    def _check_available_tools(self):
        """Check which formatting tools are available"""
        available = set()
        common_tools = ['gofmt', 'goimports', 'black', 'autopep8', 'prettier', 'jq', 'rustfmt', 'shfmt']
        
        for tool in common_tools:
            if check_tool_available(tool):
                available.add(tool)
        
        return available
    
    def can_format(self):
        """Check if this file can be formatted"""
        if not self.file_info or self.file_info['is_binary']:
            return False
        
        if self.file_info['size_mb'] > 1:  # Skip files larger than 1MB
            return False
        
        suffix = self.file_info['suffix']
        if suffix not in self.FORMATTERS:
            return False
        
        # Check if any formatter is available for this file type
        formatters = self.FORMATTERS[suffix]
        return any(f['tool'] in self.available_tools for f in formatters)
    
    def get_best_formatter(self):
        """Get the best available formatter for this file"""
        if not self.can_format():
            return None
        
        suffix = self.file_info['suffix']
        formatters = self.FORMATTERS[suffix]
        
        for formatter in formatters:
            if formatter['tool'] in self.available_tools:
                return formatter
        
        return None
    
    def format_file(self):
        """Format the file using the best available formatter"""
        formatter = self.get_best_formatter()
        if not formatter:
            return False, "No suitable formatter available"
        
        file_path = self.file_info['path']
        cmd = formatter['cmd'].format(file=file_path)
        
        log(f"Formatting {file_path.name} with {formatter['desc']}...")
        
        # Create backup
        backup_path = f"{file_path}.backup"
        try:
            import shutil
            shutil.copy2(file_path, backup_path)
        except Exception as e:
            return False, f"Could not create backup: {e}"
        
        # Run formatter
        success, stdout, stderr = run_command(cmd, cwd=file_path.parent)
        
        if success:
            # Remove backup on success
            try:
                os.unlink(backup_path)
            except:
                pass
            log(f"Successfully formatted {file_path.name}", "SUCCESS")
            return True, formatter['desc']
        else:
            # Restore from backup on failure
            try:
                import shutil
                shutil.move(backup_path, file_path)
            except:
                pass
            error_msg = stderr or stdout or "Unknown error"
            log(f"Failed to format {file_path.name}: {error_msg}", "ERROR")
            return False, error_msg

def suggest_formatter_installation(suffix):
    """Suggest how to install formatters for a file type"""
    suggestions = {
        '.go': "Install Go tools: go install golang.org/x/tools/cmd/goimports@latest",
        '.py': "Install Python formatter: pip install black autopep8",
        '.js': "Install Prettier: npm install -g prettier",
        '.ts': "Install Prettier: npm install -g prettier", 
        '.json': "Install jq: brew install jq (macOS) or apt-get install jq (Ubuntu)",
        '.rs': "Install rustfmt: rustup component add rustfmt",
        '.sh': "Install shfmt: brew install shfmt (macOS) or go install mvdan.cc/sh/v3/cmd/shfmt@latest",
        '.css': "Install Prettier: npm install -g prettier",
        '.html': "Install Prettier: npm install -g prettier",
        '.yaml': "Install Prettier: npm install -g prettier",
        '.yml': "Install Prettier: npm install -g prettier",
        '.md': "Install Prettier: npm install -g prettier",
    }
    
    return suggestions.get(suffix, "Check language-specific formatter documentation")

def main():
    """Main execution function"""
    if len(sys.argv) < 2:
        log("No file path provided", "ERROR")
        return 1
    
    file_path = sys.argv[1]
    
    # Skip if file path looks like a directory or special path
    if not file_path or file_path.endswith('/') or '..' in file_path:
        return 0
    
    file_info = get_file_info(file_path)
    if not file_info:
        log(f"Could not analyze file: {file_path}", "WARN")
        return 0
    
    # Skip binary files or very large files
    if file_info['is_binary']:
        return 0
    
    if file_info['size_mb'] > 1:
        log(f"Skipping large file ({file_info['size_mb']:.1f}MB): {file_info['name']}", "WARN")
        return 0
    
    # Try to format the file
    formatter = CodeFormatter(file_info)
    
    if not formatter.can_format():
        suffix = file_info['suffix']
        if suffix in CodeFormatter.FORMATTERS:
            log(f"No formatter available for {suffix} files", "WARN")
            print(f"💡 Suggestion: {suggest_formatter_installation(suffix)}")
        return 0
    
    success, message = formatter.format_file()
    
    if success:
        return 0
    else:
        log(f"Formatting failed: {message}", "ERROR")
        return 0  # Don't fail the hook, just log the error

if __name__ == "__main__":
    exit(main())