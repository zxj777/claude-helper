#!/usr/bin/env python3
"""
Security Check Hook
Scans code for common security vulnerabilities and sensitive information leaks.
"""

import os
import sys
import re
import json
from pathlib import Path
from datetime import datetime
from typing import List, Dict, Tuple

def log(message, level="INFO"):
    """Simple logging"""
    symbols = {"INFO": "🛡️", "WARN": "⚠️", "ERROR": "❌", "VULN": "🚨", "OK": "✅"}
    symbol = symbols.get(level, "ℹ️")
    print(f"{symbol} {message}")

class SecurityRule:
    """Represents a security check rule"""
    def __init__(self, name, pattern, languages, severity, description, suggestion):
        self.name = name
        self.pattern = re.compile(pattern, re.IGNORECASE | re.MULTILINE)
        self.languages = languages  # List of file extensions
        self.severity = severity    # HIGH, MEDIUM, LOW
        self.description = description
        self.suggestion = suggestion
    
    def check(self, content, file_path):
        """Check content against this rule"""
        matches = []
        for match in self.pattern.finditer(content):
            line_num = content[:match.start()].count('\\n') + 1
            matched_text = match.group(0)
            matches.append({
                'line': line_num,
                'text': matched_text.strip()[:100],  # Limit length
                'rule': self.name,
                'severity': self.severity,
                'description': self.description,
                'suggestion': self.suggestion
            })
        return matches

class SecurityScanner:
    """Main security scanner with predefined rules"""
    
    def __init__(self):
        self.rules = self._load_security_rules()
    
    def _load_security_rules(self):
        """Load all security checking rules"""
        rules = []
        
        # 1. Hardcoded Secrets
        rules.extend([
            SecurityRule(
                "hardcoded-password",
                r'(password|pwd|passwd)\\s*[=:]\\s*["\'][^"\'\\s]{6,}["\']',
                ["*"],
                "HIGH",
                "Hardcoded password detected",
                "Use environment variables or secure config files"
            ),
            SecurityRule(
                "api-key-pattern",
                r'(api[_-]?key|apikey|access[_-]?token|secret[_-]?key)\\s*[=:]\\s*["\'][A-Za-z0-9+/]{20,}["\']',
                ["*"],
                "HIGH", 
                "Potential API key or token exposed",
                "Store sensitive keys in environment variables"
            ),
            SecurityRule(
                "aws-credentials",
                r'(AKIA[0-9A-Z]{16}|aws_secret_access_key)',
                ["*"],
                "HIGH",
                "AWS credentials detected in code",
                "Use AWS IAM roles or environment variables"
            ),
            SecurityRule(
                "private-key",
                r'-----BEGIN\\s+(RSA\\s+)?PRIVATE\\s+KEY-----',
                ["*"],
                "HIGH",
                "Private key found in code",
                "Store private keys securely, never in source code"
            )
        ])
        
        # 2. Dangerous Function Calls
        rules.extend([
            SecurityRule(
                "eval-usage",
                r'\\beval\\s*\\(',
                [".js", ".py", ".php"],
                "HIGH",
                "Use of eval() function - code injection risk",
                "Avoid eval(). Use safer alternatives for dynamic code execution"
            ),
            SecurityRule(
                "exec-usage", 
                r'\\bexec\\s*\\(',
                [".py"],
                "HIGH",
                "Use of exec() function - code injection risk",
                "Avoid exec(). Consider ast.literal_eval() for safe evaluation"
            ),
            SecurityRule(
                "system-calls",
                r'(system|shell_exec|exec|passthru)\\s*\\(',
                [".php", ".py", ".js"],
                "MEDIUM",
                "System command execution detected",
                "Validate and sanitize all inputs before system calls"
            ),
            SecurityRule(
                "unsafe-deserialization",
                r'(pickle\\.loads|yaml\\.load|unserialize)\\s*\\(',
                [".py", ".php"],
                "HIGH",
                "Unsafe deserialization detected",
                "Use safe deserialization methods (yaml.safe_load, etc.)"
            )
        ])
        
        # 3. SQL Injection Risks
        rules.extend([
            SecurityRule(
                "sql-injection",
                r'(SELECT|INSERT|UPDATE|DELETE).*\\+.*["\'][^"\']*["\']',
                ["*"],
                "HIGH",
                "Potential SQL injection via string concatenation",
                "Use parameterized queries or prepared statements"
            ),
            SecurityRule(
                "format-sql",
                r'(SELECT|INSERT|UPDATE|DELETE).*\\.format\\s*\\(',
                [".py"],
                "HIGH",
                "SQL query with format() - injection risk",
                "Use parameterized queries instead of string formatting"
            )
        ])
        
        # 4. XSS and Template Injection
        rules.extend([
            SecurityRule(
                "html-concatenation",
                r'["\']<[^>]*>["\']\\s*\\+|\\+\\s*["\']<[^>]*>["\']',
                [".js", ".py", ".php"],
                "MEDIUM", 
                "HTML string concatenation - potential XSS",
                "Use template engines with auto-escaping"
            ),
            SecurityRule(
                "direct-html-write",
                r'(innerHTML|outerHTML|document\\.write)\\s*[=+]',
                [".js"],
                "MEDIUM",
                "Direct HTML manipulation - XSS risk",
                "Use textContent or properly escape HTML content"
            )
        ])
        
        # 5. Path Traversal
        rules.extend([
            SecurityRule(
                "path-traversal",
                r'(open|fopen|readfile|include|require)\\s*\\([^)]*\\.\\./[^)]*\\)',
                ["*"],
                "MEDIUM",
                "Potential path traversal with '../'",
                "Validate and sanitize file paths, use absolute paths"
            ),
            SecurityRule(
                "user-controlled-path",
                r'(open|fopen|readfile)\\s*\\([^)]*request\\.|\\$_GET|\\$_POST',
                [".php", ".py"],
                "HIGH",
                "File operation with user input - path traversal risk", 
                "Validate file paths and use whitelist approach"
            )
        ])
        
        # 6. Weak Cryptography
        rules.extend([
            SecurityRule(
                "weak-hash",
                r'\\b(md5|sha1)\\s*\\(',
                ["*"],
                "MEDIUM",
                "Weak cryptographic hash function",
                "Use SHA-256 or stronger hashing algorithms"
            ),
            SecurityRule(
                "weak-random",
                r'\\b(rand|srand|mt_rand)\\s*\\(',
                [".php", ".c", ".cpp"],
                "LOW",
                "Weak random number generation",
                "Use cryptographically secure random functions"
            )
        ])
        
        # 7. Debug/Development Code
        rules.extend([
            SecurityRule(
                "debug-prints",
                r'(console\\.log|print|echo|var_dump)\\s*\\([^)]*\\$_|request\\.|password|token|key',
                ["*"],
                "MEDIUM",
                "Sensitive data in debug output",
                "Remove debug statements or avoid logging sensitive data"
            ),
            SecurityRule(
                "todo-security",
                r'(TODO|FIXME|HACK).*\\b(security|password|key|token|auth)',
                ["*"],
                "LOW",
                "Security-related TODO found",
                "Address security TODOs before production"
            )
        ])
        
        return rules
    
    def scan_file(self, file_path):
        """Scan a single file for security issues"""
        path = Path(file_path)
        
        if not path.exists() or path.is_dir():
            return []
        
        try:
            # Skip binary files
            with open(path, 'rb') as f:
                if b'\\x00' in f.read(8192):
                    return []
            
            # Read file content
            with open(path, 'r', encoding='utf-8', errors='ignore') as f:
                content = f.read()
            
            # Skip very large files (>1MB)
            if len(content) > 1024 * 1024:
                return []
            
        except Exception:
            return []
        
        # Run security checks
        all_issues = []
        file_ext = path.suffix.lower()
        
        for rule in self.rules:
            # Check if rule applies to this file type
            if "*" not in rule.languages and file_ext not in rule.languages:
                continue
            
            issues = rule.check(content, str(path))
            all_issues.extend(issues)
        
        return all_issues

def format_security_report(issues, file_path):
    """Format security issues into a readable report"""
    if not issues:
        log("No security issues detected", "OK")
        return
    
    # Group by severity
    by_severity = {"HIGH": [], "MEDIUM": [], "LOW": []}
    for issue in issues:
        by_severity[issue['severity']].append(issue)
    
    file_name = Path(file_path).name
    print(f"\\n🚨 Security Issues Found in {file_name}")
    print("=" * 50)
    
    for severity in ["HIGH", "MEDIUM", "LOW"]:
        if by_severity[severity]:
            icon = {"HIGH": "🔴", "MEDIUM": "🟡", "LOW": "🟠"}[severity]
            print(f"\\n{icon} {severity} SEVERITY ({len(by_severity[severity])} issues)")
            
            for i, issue in enumerate(by_severity[severity], 1):
                print(f"\\n  {i}. Line {issue['line']}: {issue['description']}")
                print(f"     Code: {issue['text']}")
                print(f"     💡 {issue['suggestion']}")
    
    # Summary
    total = len(issues)
    high_count = len(by_severity["HIGH"])
    
    print(f"\\n📊 Summary: {total} issues found")
    if high_count > 0:
        print(f"⚠️  {high_count} HIGH severity issues require immediate attention!")

def save_security_context(issues, file_path):
    """Save security scan results to context file"""
    context_dir = Path(".claude/context")
    context_dir.mkdir(parents=True, exist_ok=True)
    
    context = {
        "trigger": "security-check",
        "timestamp": datetime.now().isoformat(),
        "file_path": str(file_path),
        "total_issues": len(issues),
        "issues_by_severity": {
            "HIGH": [i for i in issues if i['severity'] == 'HIGH'],
            "MEDIUM": [i for i in issues if i['severity'] == 'MEDIUM'], 
            "LOW": [i for i in issues if i['severity'] == 'LOW']
        }
    }
    
    context_file = context_dir / "security-context.json"
    try:
        with open(context_file, 'w', encoding='utf-8') as f:
            json.dump(context, f, indent=2, ensure_ascii=False)
    except Exception as e:
        log(f"Could not save security context: {e}", "WARN")

def main():
    """Main execution function"""
    if len(sys.argv) < 2:
        log("No file path provided", "ERROR")
        return 1
    
    file_path = sys.argv[1]
    
    # Skip if file path looks invalid
    if not file_path or '..' in file_path:
        return 0
    
    path = Path(file_path)
    if not path.exists():
        return 0
    
    # Skip certain file types
    skip_extensions = {'.png', '.jpg', '.jpeg', '.gif', '.pdf', '.zip', '.tar', '.gz', '.exe', '.dll', '.so'}
    if path.suffix.lower() in skip_extensions:
        return 0
    
    log(f"Scanning {path.name} for security issues...")
    
    scanner = SecurityScanner()
    issues = scanner.scan_file(file_path)
    
    # Save context for potential follow-up
    save_security_context(issues, file_path)
    
    # Display results
    format_security_report(issues, file_path)
    
    return 0

if __name__ == "__main__":
    exit(main())