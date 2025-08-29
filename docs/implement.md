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

## 工作流系统扩展设计

### 🎯 产品愿景
打造交互式前端开发工作流CLI工具，从需求分析到测试交付的端到端流程支持，实现人机协作的智能开发工作流。

### 📋 完整前端开发工作流程（可交互）

#### 1. 需求分析阶段
- **需求理解确认**：Claude 解析需求 → 人工确认理解正确性
- **任务拆解建议**：Claude 提供拆解方案 → 人工调整和完善
- **复杂度评估**：Claude 初步评估 → 人工基于经验修正
- **技术方案讨论**：Claude 提供多个技术方案 → 人工选择和定制
- **风险点识别**：Claude 列出潜在风险 → 人工补充遗漏的风险

#### 2. 任务规划阶段
- **工时估算协作**：Claude 给出预估时间 → 人工基于项目经验调整
- **依赖关系梳理**：Claude 分析任务依赖 → 人工验证和补充
- **里程碑制定**：Claude 建议关键节点 → 人工结合业务需求调整
- **并行开发规划**：Claude 识别可并行任务 → 人工考虑资源分配

#### 3. 设计转换阶段
- **组件设计评审**：Claude 分析设计稿提取组件 → 人工确认组件粒度
- **状态管理建议**：Claude 推荐状态管理方案 → 人工选择适合的方案
- **API 接口设计**：Claude 生成接口定义 → 人工review和完善

#### 4. 编码开发阶段
- **代码生成 + 人工review**：Claude 生成代码框架 → 人工检查和优化
- **逻辑实现协作**：Claude 提供实现思路 → 人工编写具体逻辑
- **重构建议**：Claude 发现代码问题 → 人工判断是否采纳

#### 5. 测试阶段
- **测试用例生成**：Claude 生成测试用例 → 人工补充边界情况
- **测试结果分析**：Claude 分析测试报告 → 人工判断问题优先级
- **Bug 定位协助**：Claude 提供调试建议 → 人工验证和修复

#### 6. 代码审查阶段
- **自动检查 + 人工判断**：Claude 标记问题 → 人工决定是否修改
- **性能优化建议**：Claude 提供优化方案 → 人工评估投入产出比
- **安全漏洞分析**：Claude 识别安全问题 → 人工评估风险等级

#### 7. 构建发布阶段
- **构建配置优化**：Claude 建议配置调整 → 人工验证效果
- **发布风险评估**：Claude 分析发布影响 → 人工制定回滚预案
- **监控指标设计**：Claude 推荐监控点 → 人工选择关键指标

### 🏗️ 扩展架构设计

#### 目录结构扩展
```
claude-helper/
├── cmd/
│   ├── claude-helper/          # 现有CLI入口
│   └── workflow-manager/       # 新增：工作流管理器
├── internal/
│   ├── cli/                   # 现有CLI命令
│   ├── config/                # 现有配置管理
│   ├── assets/                # 现有模板资源
│   ├── workflow/              # 新增：工作流引擎
│   │   ├── engine/           # 工作流执行引擎
│   │   ├── stages/           # 各阶段处理器
│   │   ├── interactive/      # 交互处理
│   │   └── state/           # 状态管理
│   ├── ui/                   # 新增：用户界面
│   │   └── tui/             # 终端UI (BubbleTea)
│   └── storage/              # 新增：数据存储
│       ├── project/         # 项目数据
│       ├── session/         # 会话状态
│       └── history/         # 操作历史
├── pkg/
│   ├── types/                # 扩展现有类型
│   ├── workflow/             # 工作流类型定义
│   ├── claude/               # Claude API封装
│   └── frontend/             # 前端特定类型
└── assets/
    ├── templates/            # 现有模板
    └── workflows/            # 新增：工作流模板
        ├── frontend/         # 前端工作流
        ├── backend/          # 后端工作流
        └── fullstack/        # 全栈工作流
```

#### 轻量化技术栈
```go
// 保持现有Go技术栈，扩展新库
- Go 1.21+                    // 主语言
- Cobra                       // CLI框架 (已有)
- Viper                       // 配置管理 (已有)

// 新增核心库
- BubbleTea                   // 终端UI框架
- Lipgloss                    // 终端样式
- Claude API SDK              // Claude官方SDK
- WebSocket                   // 实时交互 (可选)

// 无需的组件
❌ 数据库 (SQLite/GORM)
❌ Web服务器 (Gin/Fiber) 
❌ 复杂的持久化层
```

### 📁 文件系统存储设计

#### 项目配置结构
```
项目根目录/
├── .claude/
│   ├── settings.json         # 现有Claude配置
│   ├── hooks/               # 现有hooks目录
│   ├── agents/              # 现有agents目录
│   ├── workflows/           # 新增：工作流配置
│   │   ├── frontend.yaml    # 前端工作流定义
│   │   └── custom.yaml      # 自定义工作流
│   ├── sessions/            # 新增：会话状态
│   │   ├── current.json     # 当前会话状态
│   │   └── history/         # 历史会话记录
│   └── cache/               # 临时缓存
│       ├── claude-context.json  # Claude对话上下文
│       └── user-preferences.json # 用户偏好
```

#### 核心数据结构
```go
// 工作流定义
type Workflow struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Type        string                 `json:"type"`        // frontend, backend, fullstack
    Stages      []Stage                `json:"stages"`
    Context     WorkflowContext        `json:"context"`
    State       WorkflowState          `json:"state"`
}

// 工作流阶段
type Stage struct {
    ID           string                 `json:"id"`
    Name         string                 `json:"name"`
    Type         StageType              `json:"type"`        // analysis, planning, development, etc.
    Actions      []Action               `json:"actions"`
    Dependencies []string               `json:"dependencies"`
    Interactive  bool                   `json:"interactive"`
    Status       StageStatus            `json:"status"`      // pending, in_progress, completed
}

// 交互会话
type InteractiveSession struct {
    WorkflowID     string             `json:"workflow_id"`
    StageID        string             `json:"stage_id"`
    UserInputs     []UserInput        `json:"user_inputs"`
    ClaudeOutputs  []ClaudeOutput     `json:"claude_outputs"`
    Confirmations  []Confirmation     `json:"confirmations"`
    Timestamp      time.Time          `json:"timestamp"`
}

// 工作流状态
type WorkflowState struct {
    CurrentStage    string             `json:"current_stage"`
    StageHistory    []StageRecord      `json:"stage_history"`
    UserInputs      map[string]any     `json:"user_inputs"`
    ClaudeResponses []ClaudeResponse   `json:"claude_responses"`
    LastUpdated     time.Time          `json:"last_updated"`
}
```

### 🚀 CLI命令扩展设计

#### 新增工作流命令
```bash
# 工作流管理命令
claude-helper workflow list                    # 列出可用工作流模板
claude-helper workflow start frontend         # 启动前端工作流
claude-helper workflow resume                 # 恢复中断的工作流
claude-helper workflow status                 # 查看当前工作流状态
claude-helper workflow goto <stage>           # 跳转到指定阶段
claude-helper workflow reset                  # 重置工作流状态

# 交互式会话命令
claude-helper interactive start               # 启动交互式会话
claude-helper interactive history             # 查看交互历史
claude-helper interactive export              # 导出会话记录

# 模板管理扩展
claude-helper template create workflow <name> # 创建自定义工作流模板
claude-helper template validate <file>        # 验证工作流模板
```

### 🎯 交互机制设计

#### 确认机制
- 每个阶段都有"确认/修改/重新生成"选项
- 支持批量确认和单项调整  
- 保存人工修正的历史，供后续学习
- 智能推荐基于历史选择

#### 反馈循环
- 记录人工修正的模式，改进Claude的建议质量
- 支持自定义规则和偏好设置
- 团队间共享修正经验
- 持续优化工作流模板

#### 状态管理
- 支持工作流中断和恢复
- 自动保存用户输入和Claude响应
- 版本控制友好的状态文件格式
- 支持多项目并行工作流

### 📈 实施策略

#### Phase 1: 核心工作流引擎 (4-6周)
1. 扩展现有CLI命令支持工作流
2. 实现基础的阶段管理和状态跟踪
3. 添加简单的交互机制
4. 创建前端工作流基础模板

#### Phase 2: 交互式界面 (3-4周)
1. 实现TUI界面用于工作流管理
2. 添加实时的Claude对话功能
3. 实现确认/修改/重新生成机制
4. 完善状态持久化和恢复

#### Phase 3: 前端工作流特化 (2-3周)
1. 实现完整的前端开发工作流模板
2. 添加项目类型检测和推荐
3. 集成常用前端工具链
4. 优化用户体验和性能

### 💡 关键设计原则

#### 保持轻量化
- 基于现有架构扩展，不引入重型依赖
- 使用文件系统而非数据库存储
- 保持CLI工具的简洁性和可移植性

#### 人机协作优先
- 避免Claude幻觉导致的错误决策
- 每个关键步骤都需要人工确认
- 提供丰富的交互选项和反馈机制

#### 可扩展性
- 支持自定义工作流模板
- 可插拔的阶段处理器
- 支持不同技术栈和项目类型
- 便于社区贡献和扩展

这个扩展设计将claude-helper从单纯的组件管理工具升级为完整的Claude工作流产品，为前端开发者提供从需求分析到项目交付的全流程智能化支持。