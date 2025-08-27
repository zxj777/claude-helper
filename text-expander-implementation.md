# Text Expander Hook 实现问题总结

## 概述

在实现Claude Code的text-expander hook时遇到了一系列路径和编码问题，本文档总结了这些问题及其解决方案。

## 主要问题分析

### 1. 路径兼容性问题

**问题描述：**
- 在Windows MINGW/Git Bash环境中，绝对路径格式不被正确识别
- 原始配置使用了`/d/work/claude-helper/.claude/hooks/...`格式，导致"系统找不到指定路径"错误

**解决方案：**
- 使用相对路径替代绝对路径：`.claude/hooks/run-python.sh`
- 创建跨平台兼容的Python运行脚本`run-python.sh`

### 2. Unicode编码问题

**问题描述：**
- 用户输入包含无效的UTF-8字符（如`\udc80`, `\udcaf`等代理字符）
- Python脚本在处理这些字符时崩溃：`UnicodeEncodeError: 'utf-8' codec can't encode character '\udc80'`
- 日志写入失败导致脚本无法完成执行

**解决方案：**
```python
# 清理输入中的无效Unicode字符
try:
    prompt = prompt.encode('utf-8', errors='ignore').decode('utf-8')
except:
    prompt = prompt.encode('ascii', errors='ignore').decode('ascii')

# 在文件操作中使用错误处理
with open('.claude/hook-error.log', 'a', encoding='utf-8', errors='replace') as f:
```

### 3. Hook输出格式问题

**问题描述：**
- 最初使用简单的`print()`输出，Claude Code无法正确解析
- 需要特定的JSON格式才能被Claude Code识别和处理

**解决方案：**
```python
# 使用正确的hook输出格式
result = {
    "hookSpecificOutput": {
        "hookEventName": "UserPromptSubmit",
        "additionalContext": f"用户的意思是: {expanded_prompt}"
    }
}
print(json.dumps(result, ensure_ascii=True), flush=True)
```

### 4. 跨平台Python执行问题

**问题描述：**
- 不同系统上Python命令名称不同（`python`, `python3`, `py`）
- 直接使用`python`命令可能在某些系统上失败

**解决方案：**
创建`run-python.sh`脚本：
```bash
#!/bin/bash
if command -v python3 > /dev/null 2>&1; then
    python3 "$@"
elif command -v python > /dev/null 2>&1; then
    python "$@"
elif command -v py > /dev/null 2>&1; then
    py -3 "$@"
else
    echo "Python not found. Please install Python or add it to PATH." >&2
    exit 1
fi
```

## 调试过程

### 调试工具
1. **日志文件**：`.claude/hook-test.log` - 记录hook执行过程
2. **错误日志**：`.claude/hook-error.log` - 记录异常信息
3. **Claude Code调试信息**：通过`[DEBUG]`信息跟踪hook执行状态

### 关键调试发现
1. **环境变量传递**：`$PROMPT`变量未被正确展开，实际传递的是字面量字符串
2. **JSON解析**：Claude Code通过stdin传递JSON数据，而不是命令行参数
3. **输出解析**：Claude Code期望以`{`开头的JSON输出，其他格式被视为普通文本

## 最终修改

### 源文件位置
- `assets/templates/hooks/text-expander.yaml`
- `internal/assets/templates/hooks/text-expander.yaml`

### 主要改进
1. **编码安全**：添加UTF-8字符清理和错误处理
2. **跨平台兼容**：使用相对路径和Python运行脚本
3. **正确的输出格式**：使用Claude Code期望的JSON结构
4. **增强的错误处理**：添加详细的错误日志记录
5. **更好的调试支持**：添加执行过程日志

## 测试验证

修改完成后，text-expander hook能够：
- ✅ 正确处理包含无效UTF-8字符的输入
- ✅ 在Windows/Mac/Linux环境中正常工作
- ✅ 正确识别和展开文本快捷键
- ✅ 向Claude提供正确的上下文信息

## 经验教训

1. **编码问题是跨平台应用的常见陷阱**，需要从一开始就考虑编码安全
2. **路径处理**在不同操作系统间差异很大，相对路径通常更可靠
3. **API文档理解**很重要，Claude Code的hook机制有特定的输入输出格式要求
4. **充分的调试日志**对于诊断问题至关重要，特别是在复杂的执行环境中

## 相关文件

- 源模板文件：`assets/templates/hooks/text-expander.yaml`
- 生成的脚本：`.claude/hooks/text-expander.py`
- Python运行器：`.claude/hooks/run-python.sh`
- 配置文件：`.claude/config/text-expander.json`
- 调试日志：`.claude/hook-test.log`, `.claude/hook-error.log`