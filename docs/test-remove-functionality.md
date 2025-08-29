# Remove 功能修复验证

## 修复的问题

### 1. Settings.json 中移除所有hooks的问题
**原因**: `containsHookName` 函数逻辑错误，总是返回 `true`
**修复**: 改为检查命令中是否包含特定的hook脚本名称

```go
// 修复前 (错误)
func containsHookName(command, hookName string) bool {
    return command != "" && hookName != ""  // 总是返回true!
}

// 修复后 (正确)  
func containsHookName(command, hookName string) bool {
    if command == "" || hookName == "" {
        return false
    }
    
    return strings.Contains(command, hookName+".py") || 
           strings.Contains(command, hookName+".sh") ||
           strings.Contains(command, hookName+".js") ||
           strings.Contains(command, "/"+hookName) ||
           strings.Contains(command, "\\"+hookName)
}
```

### 2. 不清理相关文件的问题
**原因**: `removeHook` 只调用 `RemoveHookFromSettings`，不清理文件
**修复**: 增强 `removeHook` 函数，完整清理所有相关文件

## 现在的清理范围

### Hook脚本文件
- `.claude/hooks/{hookName}.py`
- `.claude/hooks/{hookName}.sh`
- `.claude/hooks/{hookName}.js`
- `.claude/hooks/{hookName}.ts`

### 配置文件
- `.claude/config/{hookName}.json`
- `.claude/config/{hookName}-config.json`
- **特殊hook的特定配置**:
  - `audio-notification`: `audio-notification.json`
  - `task-notification`: `notification.json`
  - `text-expander`: `text-expander.json`

### 音频文件 (可选)
- `.claude/sounds/` (用户确认后删除)

### 临时/状态文件
- `.claude/last-notification-time`
- `.claude/last-audio-notification`
- `.claude/hook-error.log`
- `.claude/notification-error.log`

## 测试步骤

### 1. 安装一个hook进行测试
```bash
./claude-helper install task-notification
```

### 2. 验证安装的文件
```bash
# 检查settings.json
cat .claude/settings.json

# 检查创建的文件
ls -la .claude/hooks/
ls -la .claude/config/
ls -la .claude/sounds/
```

### 3. 移除hook
```bash
./claude-helper remove task-notification
```

### 4. 验证清理结果
```bash
# 检查settings.json (应该只移除对应的hook)
cat .claude/settings.json

# 检查文件是否被清理
ls -la .claude/hooks/     # task-notification.py 应该被删除
ls -la .claude/config/    # notification.json 应该被删除
ls -la .claude/sounds/    # 根据用户选择决定是否删除
```

### 5. 测试多个hook共存
```bash
# 安装多个hook
./claude-helper install text-expander
./claude-helper install task-notification

# 移除其中一个
./claude-helper remove text-expander

# 验证只移除了指定的hook，其他hook保持不变
cat .claude/settings.json
ls -la .claude/hooks/
ls -la .claude/config/
```

## 用户体验改进

### 1. 详细的反馈信息
```
Removing component: task-notification
Found installed hook: task-notification

This will remove the hook 'task-notification' from your Claude Code configuration.
Are you sure you want to continue? (y/N): y

Removed hook script: /path/.claude/hooks/task-notification.py
Removed config file: /path/.claude/config/notification.json
Do you want to remove audio files? (y/N): y
Removed sounds directory: /path/.claude/sounds
Removed temp file: /path/.claude/last-notification-time

✓ Successfully removed hook 'task-notification'
```

### 2. 安全确认
- 默认需要用户确认才删除
- 音频文件单独确认 (因为可能被多个hook共享)
- 可使用 `-y` 参数跳过确认

### 3. 错误处理
- 如果某个文件删除失败，显示警告但继续执行
- 保持目录结构以便其他组件使用

## 向后兼容

### 支持旧版本安装的hook
- 自动检测并处理旧版本的配置文件格式
- 兼容不同的文件命名模式

### 渐进式清理
- 优先清理核心配置 (settings.json)
- 然后清理脚本和配置文件
- 最后清理可选的资源文件

## 测试场景

### 场景1: 单个hook完整测试
1. 安装 `task-notification`
2. 验证所有文件都被创建
3. 移除 `task-notification`
4. 验证所有文件都被清理，settings.json正确

### 场景2: 多hook并存测试
1. 安装 `text-expander`, `task-notification`
2. 移除 `text-expander`
3. 验证只有text-expander相关文件被删除
4. task-notification的hook和文件保持不变

### 场景3: 配置冲突测试
1. 安装 `audio-notification` (旧版)
2. 安装 `task-notification` (新版，共享音频文件)
3. 移除 `audio-notification`
4. 验证音频文件处理正确，新版配置不受影响

这些修复确保了 `claude-helper remove` 功能的完整性和可靠性。