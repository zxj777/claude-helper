#!/bin/bash

echo "🔧 修复 task-notification hook..."

# 1. 确保脚本有执行权限
echo "1. 设置执行权限..."
chmod +x .claude/hooks/run-python.sh
chmod +x .claude/hooks/task-notification.py

# 2. 测试 Python 脚本
echo "2. 测试 Python 脚本..."
echo '{"test": "data"}' | .claude/hooks/run-python.sh .claude/hooks/task-notification.py
if [ $? -eq 0 ]; then
    echo "✅ Python 脚本测试通过"
else
    echo "❌ Python 脚本测试失败"
fi

# 3. 检查通知配置
echo "3. 检查通知配置..."
if [ -f ".claude/config/notification.json" ]; then
    echo "✅ 通知配置文件存在"
    cat .claude/config/notification.json | jq . 2>/dev/null || echo "配置文件内容："
    cat .claude/config/notification.json
else
    echo "❌ 通知配置文件不存在"
fi

# 4. 测试完整的 hook 命令
echo "4. 测试完整的 hook 命令..."
echo '{"tool_name": "test", "result": "success"}' | bash .claude/hooks/run-python.sh .claude/hooks/task-notification.py
if [ $? -eq 0 ]; then
    echo "✅ Hook 命令测试通过"
else
    echo "❌ Hook 命令测试失败"
fi

echo "🎉 修复脚本执行完成！"