# Claude Helper Architecture Design & Implementation

## Project Overview
Claude Helper is a CLI tool for managing Claude Code hooks and agents, designed to help users quickly build and manage Claude workflows.

## Core Goals
1. Integrate common Claude hooks and agents
2. Provide convenient CLI interface for installing and managing components  
3. Support extensible architecture with clear layers
4. Support fine-grained configuration control

## Project Structure

```
claude-helper/
├── cmd/claude-helper/          # CLI application entry point
│   └── main.go                # Main program entry
├── internal/                   # Private packages (not exposed)
│   ├── cli/                   # Command line interface implementation
│   │   ├── root.go            # Root command and global config
│   │   ├── list.go            # List components command
│   │   ├── install.go         # Install component command
│   │   ├── remove.go          # Remove component command
│   │   ├── enable.go          # Enable component command
│   │   ├── disable.go         # Disable component command
│   │   └── create.go          # Create custom component command
│   ├── config/                # Configuration management
│   │   ├── manager.go         # Config file read/write
│   │   └── claude.go          # Claude Code integration
│   ├── template/              # Template engine
│   │   ├── parser.go          # Template file parsing
│   │   └── renderer.go        # Template rendering
│   └── installer/             # Component installer
│       ├── agent.go           # Agent installation logic
│       └── hook.go            # Hook installation logic
├── pkg/                       # Public packages (can be referenced externally)
│   ├── types/                 # Type definitions
│   │   └── types.go           # Core data structures
│   └── utils/                 # Utility functions
│       └── file.go            # File operation utilities
└── assets/                    # Static resources
    └── templates/             # Pre-built templates
        ├── agents/            # Agent templates
        │   ├── code-reviewer.md
        │   ├── test-generator.md
        │   └── doc-writer.md
        └── hooks/             # Hook templates
            ├── format-code.json
            ├── git-commit.json
            └── security-scan.json
```

## Core Type Design

### 1. Agent Type
```go
type Agent struct {
    Name        string   `json:"name" yaml:"name"`
    Description string   `json:"description" yaml:"description"`
    Tools       []string `json:"tools,omitempty" yaml:"tools,omitempty"`
    Prompt      string   `json:"prompt" yaml:"prompt"`
    Enabled     bool     `json:"enabled" yaml:"enabled"`
}
```

### 2. Hook Type
```go
type Hook struct {
    Name        string            `json:"name" yaml:"name"`
    Description string            `json:"description" yaml:"description"`
    Event       HookEvent         `json:"event" yaml:"event"`
    Matcher     string            `json:"matcher,omitempty" yaml:"matcher,omitempty"`
    Command     string            `json:"command" yaml:"command"`
    Args        []string          `json:"args,omitempty" yaml:"args,omitempty"`
    Env         map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
    Script      string            `json:"script,omitempty" yaml:"script,omitempty"`
    Enabled     bool              `json:"enabled" yaml:"enabled"`
}
```

### 3. Hook Event Types
Supports all standard Claude Code events:
- `PreToolUse` - Before tool use
- `PostToolUse` - After tool use
- `UserPromptSubmit` - When user submits prompt
- `Notification` - Notification events
- `Stop` - Stop events
- `SubagentStop` - Sub-agent stop
- `SessionStart` - Session start
- `SessionEnd` - Session end

## CLI Command Design

### Basic Commands
```bash
claude-helper list                    # List all available components
claude-helper install <name>         # Install specified component
claude-helper remove <name>          # Remove component
claude-helper enable <name>          # Enable component
claude-helper disable <name>         # Disable component
claude-helper create <type> <name>   # Create custom component
claude-helper sync                   # Sync configuration to Claude
```

### Command Options
- `--config` - Specify config file path
- `--verbose` - Verbose output
- `--agents` - Show only agents
- `--hooks` - Show only hooks
- `--installed` - Show only installed components

## Configuration Management

### Config File Locations
- Global config: `~/.claude-helper.yaml`
- Project config: `./.claude-helper.yaml`

### Config Structure
```yaml
claude_path: "/path/to/claude"
templates:
  code-reviewer:
    name: "code-reviewer"
    type: "agent"
    description: "Expert code review specialist"
    version: "1.0.0"
    author: "claude-helper"
    config: {...}
installed:
  code-reviewer: true
  format-code: true
```

## Template System

### Agent Template Format
Uses Markdown format with YAML frontmatter:
```markdown
---
name: code-reviewer
description: Expert code review specialist
tools: [Read, Grep, Glob, Bash]
version: 1.0.0
---

You are a senior software engineer...
```

### Hook Template Format
Uses JSON format:
```json
{
  "name": "format-code",
  "description": "Auto-format code files",
  "event": "PostToolUse",
  "matcher": "Edit|Write",
  "script": "#!/bin/bash\n..."
}
```

## Claude Code Integration

### Config Path Detection
Auto-detect Claude Code installation path:
- macOS: `~/Library/Application Support/Claude/`
- Windows: `%APPDATA%/Claude/`
- Linux: `~/.config/Claude/`

### Config Synchronization
- Read existing Claude configuration
- Merge claude-helper configuration
- Backup original configuration
- Write updated configuration

## Error Handling & Recovery
- Recovery mechanism for corrupted config files
- Rollback operations for failed installations
- Detailed error messages and suggestions
- Configuration validation and conflict detection

## Extensibility Design
- Plugin-based template system
- Support for remote template repositories
- Version management and update mechanism
- Community-contributed template support

## Hook-Agent Integration Architecture

### Design Philosophy
The claude-helper tool follows a **"Self-contained Hooks + Optional Management Tool"** architecture that ensures maximum usability and minimal deployment friction.

### Core Principles

#### 1. **Zero External Dependencies for End Users**
- Hooks operate independently using only standard Python/Shell scripts
- No requirement to install claude-helper for hook execution
- Projects remain self-contained and portable
- Users can use hooks immediately after cloning the project

#### 2. **Optional Development Tool**
- claude-helper serves as a **development and management tool**
- Helps developers create, test, and organize Hook/Agent templates
- Generates self-contained configurations for distribution
- Not required during runtime or by end users

#### 3. **Flexible Integration Patterns**

**Pattern A: Hook Automatically Installs Agent**
```python
# smart-review.py - Self-contained hook script
def ensure_agent_available():
    agents_dir = Path(".claude/agents") 
    agent_file = agents_dir / "code-reviewer.md"
    
    if not agent_file.exists():
        # Copy from templates without external dependencies
        template_path = Path("assets/templates/agents/code-reviewer.md")
        if template_path.exists():
            shutil.copy2(template_path, agent_file)
            print("✅ Code reviewer agent ready")
```

**Pattern B: Conditional Tool Usage**
```bash
# flexible-review.sh - Adaptive hook script
if command -v claude-helper &> /dev/null; then
    # Use claude-helper if available (development environment)
    claude-helper install code-reviewer --type agent
else
    # Fall back to manual setup (production environment)
    cp assets/templates/agents/code-reviewer.md .claude/agents/
fi
```

#### 4. **Configuration Strategy**

**Development Phase:**
- Use `claude-helper` to create and manage templates
- Test hooks and agents in local environment
- Generate optimized configurations

**Deployment Phase:**
- Hooks become standalone Python/Shell scripts
- Agent templates are static markdown files
- Configuration files (settings.json) are committed to version control

**Usage Phase:**
- End users interact only with Claude Code
- Hooks trigger automatically based on configured events
- Agents are loaded seamlessly into conversations

### Integration Workflow Examples

#### Smart Code Review Workflow
```yaml
# Hook Configuration (auto-review.yaml)
name: auto-review
events: [UserPromptSubmit]
matcher: "*review*"
hooks:
  - type: command
    command: "python3 .claude/hooks/smart-review.py"
```

```python
# Hook Implementation (smart-review.py)
def main():
    # 1. Detect review request
    ensure_code_reviewer_agent()
    
    # 2. Prepare context files
    create_review_context()
    
    # 3. Agent automatically participates in conversation
    print("🎯 Code reviewer ready!")
```

#### Automated Documentation Generation
```yaml
# Hook triggers when code files are modified
name: auto-docs
events: [PostToolUse]
matcher: "Edit|Write"
hooks:
  - type: command
    command: "bash .claude/hooks/doc-generator.sh"
```

### Benefits of This Architecture

#### For End Users:
- **Zero Setup**: Clone project and hooks work immediately
- **No Dependencies**: Only need Claude Code itself
- **Transparent**: Hooks work behind the scenes
- **Reliable**: No external tool version conflicts

#### For Developers:
- **Powerful Tooling**: Rich CLI for management and testing
- **Template System**: Reusable components across projects
- **Easy Distribution**: Generate self-contained packages
- **Development Speed**: Rapid prototyping and testing

#### For Projects:
- **Portable**: Works across different environments
- **Maintainable**: Clear separation of concerns
- **Scalable**: Easy to add new hooks and agents
- **Version Control Friendly**: All configurations tracked in git

### Implementation Best Practices

#### Hook Scripts Should:
```python
# ✅ Good: Self-contained with error handling
def ensure_agent_ready():
    try:
        agents_dir = Path(".claude/agents")
        agents_dir.mkdir(parents=True, exist_ok=True)
        # Handle agent installation
    except Exception as e:
        print(f"⚠️ Could not prepare agent: {e}")
        return False
    return True
```

#### Avoid External Dependencies:
```python
# ❌ Bad: Requires claude-helper installation
subprocess.run(["claude-helper", "install", "code-reviewer"])

# ✅ Good: Direct file operations
shutil.copy2("templates/code-reviewer.md", ".claude/agents/")
```

#### Provide Graceful Fallbacks:
```python
# ✅ Good: Multiple fallback strategies
def get_agent_template():
    # Try project templates first
    if template_exists("assets/templates/agents/code-reviewer.md"):
        return load_template("assets/templates/agents/code-reviewer.md")
    
    # Try embedded fallback
    return get_embedded_template("code-reviewer")
```

This architecture ensures that claude-helper enhances the development experience while keeping the final product simple and dependency-free.

## Technology Stack
- **CLI Framework**: cobra - Powerful command-line application framework
- **Configuration**: viper - Flexible configuration solution
- **Template Engine**: text/template - Go standard library
- **YAML Parsing**: gopkg.in/yaml.v3 - YAML processing
- **File Operations**: Standard library filepath, os

## Development Phase Planning

### Phase 1: Basic Architecture ✅
- [x] Project structure setup
- [x] Basic type definitions
- [x] CLI framework initialization

### Phase 2: Core Data Structures ✅
- [x] **Hook Type Implementation** (`pkg/types/hook.go`)
  - HookEvent constants (PreToolUse, PostToolUse, etc.)
  - Hook struct with YAML/JSON serialization
  - `MergeHooksIntoClaudeConfig()` - Converts claude-helper hooks to Claude's official settings.json format
  - Proper handling of Event + Matcher grouping to avoid duplicate entries
- [x] **Agent Type Implementation** (`pkg/types/agent.go`)  
  - Agent struct with frontmatter parsing support
  - `ToMarkdown()` - Converts Agent to Claude's .md file format
  - `ParseAgentFromMarkdown()` - Parses agent from markdown with YAML frontmatter
  - AgentFrontmatter helper type for YAML parsing
- [x] **Dependencies Updated** (`go.mod`)
  - Added gopkg.in/yaml.v3 for YAML processing

### Phase 3: CLI Implementation ✅
- [x] **Basic CLI Framework** (`internal/cli/root.go`, `cmd/claude-helper/main.go`)
  - Cobra + Viper integration
  - Global configuration management
  - Command registration system
- [x] **List Command** (`internal/cli/list.go`)
  - Real file system scanning
  - Component type detection (.md for agents, .yaml for hooks)
  - Installation status checking
  - Command line filtering (--agents, --hooks, --installed)
- [x] **Configuration Manager** (`internal/config/claude.go`)
  - Cross-platform Claude path detection
  - Agent installation status checking
  - Hook installation status checking (basic)

### Phase 4: Component Management ✅
- [x] **Install Command** (`internal/cli/install.go`)
  - Component template discovery and type detection
  - Agent installation (copy .md files to ~/.claude/agents/)
  - **Hook installation** - Complete YAML parsing and settings.json integration
  - Force install option and duplicate detection
  - Comprehensive error handling and validation
- [x] **Remove Command** (`internal/cli/remove.go`) 
  - Installed component detection with type inference
  - Agent removal (delete from ~/.claude/agents/)
  - **Hook removal** - Complete settings.json modification with filtering
  - User confirmation prompts for safety
  - Graceful handling of non-existent components
- [x] **Enable/Disable Commands** (`internal/cli/enable.go`, `internal/cli/disable.go`)
  - Agent enabling/disabling via file rename (.disabled extension)
  - Hook enabling/disabling framework (basic implementation)
  - Component state management
- [x] **Enhanced Settings.json Operations** (`internal/config/claude.go`)
  - `InstallHookToSettings()` - Merge hooks into existing configuration
  - `RemoveHookFromSettings()` - Remove specific hooks with filtering
  - JSON file creation, parsing, and formatting
  - Backup and recovery mechanisms

### Phase 5: Template Creation ✅
- [x] **Create Command** (`internal/cli/create.go`)
  - Agent template generation with customizable prompts
  - Hook template generation with YAML structure  
  - Component name validation and conflict detection
  - Interactive template customization (description, tools)
  - Automatic directory creation and file management

## Current Feature Status

### ✅ Fully Implemented Commands
```bash
./claude-helper list                    # List all components with status
./claude-helper list --agents           # Filter by agents only
./claude-helper list --hooks            # Filter by hooks only  
./claude-helper list --installed        # Show only installed components

./claude-helper install <name>          # Install agent or hook
./claude-helper install <name> --force  # Force reinstall

./claude-helper remove <name>           # Remove with confirmation
./claude-helper remove <name> -y        # Remove without confirmation

./claude-helper enable <name>           # Enable disabled component
./claude-helper disable <name>          # Disable without removing

./claude-helper create agent <name>     # Create new agent template
./claude-helper create hook <name>      # Create new hook template
./claude-helper create agent <name> -d "Description" -t "tool1,tool2"
```

### 🔧 Core Features
- **Cross-platform Claude detection** (macOS/Linux/Windows)
- **Real-time installation status checking**
- **YAML/JSON template parsing and generation**
- **Settings.json manipulation with backup**
- **File-based agent management**
- **Hook-based configuration merging**
- **Template validation and error handling**
- **Interactive user confirmations**

### 🎯 Ready for Production Use
The claude-helper tool now provides a complete solution for managing Claude Code components with:
- Robust error handling and recovery
- User-friendly command-line interface
- Comprehensive logging and feedback
- Safe file operations with confirmations
- Extensible template system