# Claude Helper 音频通知功能设计文档

## 功能概述

为 Claude Helper 添加音频通知功能，当 Claude 完成任务时播放提示音，提升用户体验。用户可以在安装时选择提示音，并支持后续修改配置。

## 设计目标

1. **任务完成提醒**：Claude 完成工具操作后播放音频通知
2. **个性化配置**：用户可选择不同的提示音
3. **跨平台兼容**：支持 macOS、Linux 和 Windows Git Bash 环境
4. **智能过滤**：避免频繁通知，只在有意义的任务完成时播放

## Hook 事件选择

### 推荐事件：`PostToolUse`
- **触发时机**：Claude 完成任何工具操作后
- **适用场景**：文件编辑、命令执行、搜索等操作完成
- **优势**：能够捕获大部分用户关心的任务完成事件

### 备选事件：
- `Stop`：Claude 停止响应时（适合长时间操作结束）
- `SessionEnd`：会话结束时（适合工作完成提醒）

## 技术实现方案

### 1. 跨平台音频播放

#### macOS
```bash
afplay /path/to/sound.wav
```

#### Linux
```bash
# 优先级顺序尝试
aplay /path/to/sound.wav          # ALSA
paplay /path/to/sound.wav         # PulseAudio
play /path/to/sound.wav           # SoX
```

#### Windows (Git Bash)
```bash
# 通过 PowerShell 播放音频
powershell -c "(New-Object Media.SoundPlayer '/path/to/sound.wav').PlaySync()"
```

### 2. 音频文件格式
- **首选格式**：`.wav` (无损，跨平台兼容性最佳)
- **备选格式**：`.mp3` (文件较小，但需要解码器支持)
- **文件大小**：控制在 100KB 以下，确保快速播放

### 3. 预设音效类别

| 音效名称 | 用途 | 描述 |
|---------|------|------|
| `success.wav` | 成功完成 | 愉悦的铃声，表示任务成功 |
| `error.wav` | 出现错误 | 低沉的提示音，表示操作失败 |
| `complete.wav` | 一般完成 | 中性的提示音，任务完成 |
| `attention.wav` | 需要注意 | 较为明显的提醒音 |
| `subtle.wav` | 轻微提醒 | 轻柔的提示音，不打扰工作 |
| `chime.wav` | 清脆铃声 | 清脆悦耳的铃声 |
| `bell.wav` | 传统铃声 | 经典的铃铛声音 |

## 配置结构设计

### AudioConfig 结构体
```go
type AudioConfig struct {
    Enabled      bool   `json:"enabled"`           // 是否启用音频通知
    SuccessSound string `json:"success_sound"`     // 成功时播放的音频文件
    ErrorSound   string `json:"error_sound"`       // 错误时播放的音频文件  
    DefaultSound string `json:"default_sound"`     // 默认音频文件
    Volume       int    `json:"volume"`            // 音量 (0-100)
    Cooldown     int    `json:"cooldown_seconds"`  // 冷却时间，避免频繁播放
}
```

### 配置文件示例
```json
{
  "enabled": true,
  "success_sound": "success.wav",
  "error_sound": "error.wav", 
  "default_sound": "complete.wav",
  "volume": 70,
  "cooldown_seconds": 2
}
```

## Hook 实现逻辑

### Python 脚本逻辑流程
1. **接收 PostToolUse 事件数据**
2. **解析工具执行结果**
   - 检查工具执行是否成功
   - 分析错误信息（如果存在）
   - 判断是否是重要的任务完成事件
3. **选择合适的音效**
   - 成功：播放 success_sound
   - 失败：播放 error_sound  
   - 其他：播放 default_sound
4. **执行音频播放命令**
5. **记录播放时间（避免频繁播放）**

### 智能过滤规则
- **冷却时间**：2秒内不重复播放
- **工具类型过滤**：只对特定工具操作播放通知
- **成功状态判断**：基于工具返回状态和输出内容判断

## 安装流程设计

### 1. 组件选择
用户运行 `./claude-helper install audio-notification` 时触发

### 2. 音效试听界面
```
🔊 选择音频通知设置

请选择任务完成时的提示音：
1. success.wav    - 成功铃声 [试听]
2. complete.wav   - 完成提示 [试听] 
3. subtle.wav     - 轻柔提醒 [试听]
4. chime.wav      - 清脆铃声 [试听]
5. bell.wav       - 传统铃声 [试听]
6. attention.wav  - 注意提醒 [试听]
7. 禁用音频通知

请输入选择 (1-7): 
```

### 3. 音量设置
```
🔊 设置音量 (1-100，默认70): 
```

### 4. 配置保存
- 配置保存到 `.claude/config/audio-notification.json`
- Hook 配置写入 `.claude/settings.json`

## 目录结构

```
claude-helper/
├── internal/assets/sounds/          # 音频文件目录
│   ├── success.wav
│   ├── error.wav
│   ├── complete.wav
│   ├── attention.wav
│   ├── subtle.wav
│   ├── chime.wav
│   └── bell.wav
├── internal/assets/templates/hooks/
│   └── audio-notification.yaml      # Hook 模板
└── pkg/types/
    └── types.go                     # AudioConfig 定义
```

## 使用场景示例

### 场景1：代码编辑完成
1. 用户让 Claude 修改文件
2. Claude 使用 Edit 工具完成修改
3. PostToolUse 事件触发
4. 脚本检测到文件编辑成功
5. 播放 `success.wav`

### 场景2：命令执行失败  
1. 用户让 Claude 运行测试
2. Claude 使用 Bash 工具执行命令
3. 命令返回非零退出码
4. PostToolUse 事件触发
5. 脚本检测到执行失败
6. 播放 `error.wav`

## 扩展功能

### 未来可考虑的增强功能
1. **自定义音频文件**：允许用户上传自己的音频文件
2. **通知策略**：基于任务类型的不同通知策略
3. **视觉通知**：结合系统通知显示任务完成状态
4. **统计功能**：记录任务完成次数和类型

## 兼容性考虑

### Git Bash 特殊处理
- Windows 用户通常使用 Git Bash 作为终端
- PowerShell 命令需要通过 `powershell -c` 调用
- 路径格式需要转换为 Windows 格式

### 权限要求
- macOS：可能需要访问音频设备权限
- Linux：需要音频系统（ALSA/PulseAudio）可用
- Windows：需要 PowerShell 可用

### 错误处理
- 音频文件不存在时的降级处理
- 音频系统不可用时的静默处理
- 权限不足时的友好提示

## 测试策略

### 单元测试
- AudioConfig 结构体的序列化/反序列化
- 跨平台音频播放命令生成
- 配置文件读写功能

### 集成测试  
- 完整的 Hook 安装流程
- 不同平台的音频播放测试
- 配置修改和持久化测试

### 用户测试
- 音效试听功能
- 不同场景下的通知触发
- 配置界面的易用性

## 实现优先级

### P0 (必须实现)
- 基础音频播放功能
- PostToolUse Hook 集成
- 跨平台兼容性
- 基本配置管理

### P1 (重要功能)  
- 音效试听功能
- 智能过滤和冷却
- 错误处理机制

### P2 (增强功能)
- 多种音效选择
- 音量控制
- 通知策略优化

这个设计文档为音频通知功能的实现提供了完整的技术规格和实现指导。