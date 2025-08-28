# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Claude Helper is a CLI tool that manages Claude Code hooks and agents. It allows users to quickly install pre-built components (hooks, agents) into their Claude Code configuration and manage them efficiently.

## Common Commands

### Build and Development
- `make build` - Build the binary to bin/claude-helper
- `make dev-build` - Build with race detector for development
- `make test` - Run all tests
- `make fmt` - Format code with gofmt
- `make vet` - Run go vet for static analysis
- `make deps` - Install and tidy dependencies
- `make clean` - Clean build files
- `make install` - Install binary to GOPATH/bin

### Using the Tool
- `./claude-helper help` - Show available commands
- `./claude-helper install <component>` - Install hooks or agents
- `./claude-helper list` - List available components
- `./claude-helper enable <hook>` - Enable a hook
- `./claude-helper disable <hook>` - Disable a hook

## Architecture

### Core Structure
- `cmd/claude-helper/` - Main entry point
- `internal/cli/` - CLI command implementations (cobra-based)
- `internal/config/` - Claude configuration management
- `pkg/types/` - Core data structures (Hook, Agent, TextExpanderConfig)
- `assets/templates/` - Built-in hook and agent templates
- `internal/assets/` - Embedded template assets

### Key Components
1. **CLI Layer**: Cobra-based commands for installing, listing, enabling/disabling components
2. **Configuration Management**: Handles Claude settings.json manipulation and cross-platform paths
3. **Template System**: Pre-built hooks and agents that can be installed into projects
4. **Type System**: Defines Hook, Agent, and configuration structures

### Claude Integration
The tool integrates with Claude Code by:
- Writing to `.claude/settings.json` (project-local) or global Claude settings
- Installing agents to `.claude/agents/` directory
- Managing hook scripts in `.claude/hooks/` directory
- Supporting both local project configuration and global Claude settings

### Available Templates
Built-in hooks include:
- `auto-format` - Automatically formats code after editing
- `commit-helper` - Suggests commit messages based on changes
- `auto-review` - Code review assistance
- `security-check` - Security analysis
- `text-expander` - Text expansion functionality

### Configuration Paths
- macOS: `~/Library/Application Support/Claude/`
- Linux: `~/.config/Claude/`
- Windows: `%APPDATA%/Claude/`
- Project-local: `./.claude/` (preferred for version-controlled settings)

## Recent Refactoring (Template System)

**Issue Fixed**: Resolved YAML parsing error in text-expander hook caused by indented EOF markers in heredoc blocks.

**Template System Refactored**: 
- **Before**: Maintained duplicate templates in both `assets/templates/` and `internal/assets/templates/`
- **After**: Single source of truth in `internal/assets/templates/`, used for both development and embed
- **Architecture**: 
  - Development mode: Uses `internal/assets/templates/` directly
  - Production mode: Extracts from embedded filesystem (same source)
  - Embed directive: `//go:embed templates/*` (relative to internal/assets/)
- **Benefits**: Eliminates synchronization issues, reduces maintenance overhead, no duplicate files

**Manual Cleanup Required**: 
Run `bash cleanup.sh` to remove duplicate outer templates directory.