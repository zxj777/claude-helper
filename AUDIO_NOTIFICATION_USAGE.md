# Claude Helper 音频通知功能使用说明

## 功能介绍

音频通知功能为 Claude Helper 添加了任务完成声音提醒。当 Claude 完成工具操作后，系统会根据操作结果播放相应的提示音。

## 安装步骤

### 1. 安装音频通知 Hook

```bash
./claude-helper install audio-notification
```

### 2. 交互式配置

安装过程中会提示选择：

```
🔊 Configuring Audio Notification Settings...
Choose when to play notification sounds after Claude completes tasks.

Available notification sounds:
1. success.wav    - 成功铃声 (愉悦的成功提示)
2. complete.wav   - 完成提示 (中性的任务完成)
3. subtle.wav     - 轻柔提醒 (不打扰工作)
4. chime.wav      - 清脆铃声 (清脆悦耳)
5. bell.wav       - 传统铃声 (经典铃铛声)
6. attention.wav  - 注意提醒 (明显的提醒音)
7. 禁用音频通知

请选择默认提示音 (1-7): 2
✅ 选择了完成提示作为默认提示音

设置音量 (1-100，默认70): 80
✅ 音量设置为: 80

📝 音频通知配置已保存到: .claude/config/audio-notification.json
```

### 3. 生成音频文件

```bash
# 安装 Python 依赖 (如果还没有)
pip install numpy scipy

# 生成音频文件
python3 generate-sounds.py
```

## 配置文件

配置文件位于 `.claude/config/audio-notification.json`：

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

### 配置参数说明

- `enabled`: 是否启用音频通知
- `success_sound`: 操作成功时的音频文件
- `error_sound`: 操作失败时的音频文件  
- `default_sound`: 默认音频文件
- `volume`: 音量级别 (1-100)
- `cooldown_seconds`: 冷却时间，防止频繁播放

## 工作原理

### Hook 事件
- **触发事件**: `PostToolUse`
- **触发时机**: Claude 完成任何工具操作后
- **匹配器**: `"*"` (匹配所有操作)

### 智能判断
Python 脚本会分析工具执行结果：
- 检查输出中是否包含 "error", "failed", "exception" 等关键词
- 成功操作播放 `success_sound`
- 失败操作播放 `error_sound`
- 其他情况播放 `default_sound`

### 跨平台支持
- **macOS**: 使用 `afplay` 命令
- **Linux**: 尝试 `aplay`, `paplay`, `play` 命令
- **Windows**: 通过 PowerShell 调用 `Media.SoundPlayer`

## 文件结构

```
.claude/
├── config/
│   └── audio-notification.json    # 音频配置文件
├── hooks/
│   ├── audio-notification.py      # Hook 脚本
│   ├── run-python.sh              # 跨平台 Python 运行器
│   └── run-python.bat             # Windows Python 运行器
├── sounds/                        # 音频文件目录
│   ├── success.wav                # 成功提示音
│   ├── error.wav                  # 错误提示音
│   ├── complete.wav               # 完成提示音
│   ├── attention.wav              # 注意提示音
│   ├── subtle.wav                 # 轻柔提示音
│   ├── chime.wav                  # 清脆铃声
│   └── bell.wav                   # 传统铃声
└── settings.json                  # Claude 设置文件 (包含 Hook 配置)
```

## 自定义音频

### 1. 使用自定义音频文件
将音频文件放置在 `.claude/sounds/` 目录：

```bash
cp my-custom-sound.wav .claude/sounds/
```

### 2. 修改配置
编辑 `.claude/config/audio-notification.json`：

```json
{
  "enabled": true,
  "default_sound": "my-custom-sound.wav",
  "volume": 80,
  "cooldown_seconds": 1
}
```

### 3. 支持的音频格式
- WAV (推荐)
- MP3 (需要系统支持)
- AIFF (macOS)

## 故障排除

### 1. 无声音播放
检查以下项目：
- 音频文件是否存在：`ls .claude/sounds/`
- 系统音频播放器是否可用：
  - macOS: `which afplay`
  - Linux: `which aplay`
  - Windows: PowerShell 是否可用

### 2. Hook 未触发
检查 Hook 是否正确安装：
```bash
./claude-helper list
# 应该显示 audio-notification hook
```

### 3. Python 脚本错误
检查错误日志：
```bash
cat .claude/hook-error.log
```

### 4. 重新配置
删除配置文件后重新安装：
```bash
rm .claude/config/audio-notification.json
./claude-helper install audio-notification --force
```

## 高级配置

### 1. 禁用音频通知
```json
{
  "enabled": false
}
```

### 2. 调整冷却时间
```json
{
  "cooldown_seconds": 5
}
```

### 3. 不同操作使用不同音效
```json
{
  "success_sound": "success.wav",
  "error_sound": "error.wav", 
  "default_sound": "subtle.wav"
}
```

## 注意事项

1. **隐私**: 音频通知在本地播放，不会传输任何数据
2. **性能**: Hook 脚本执行时间短 (< 1秒)，不影响 Claude 响应速度
3. **兼容性**: 支持所有支持音频播放的操作系统
4. **工作环境**: 建议在个人工作环境使用，避免打扰他人

## 版本历史

- **v1.0.0**: 初始版本，支持基础音频通知功能
- 支持 PostToolUse 事件
- 跨平台音频播放
- 智能成功/失败判断
- 交互式配置界面

## 反馈和贡献

如有问题或建议，请在项目 GitHub 页面提交 Issue。