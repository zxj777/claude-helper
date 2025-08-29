#!/usr/bin/env python3
import shutil
import os

# 复制系统声音文件
source = '/System/Library/Sounds/Glass.aiff'
dest = '/Users/zhuxiaojiang/toys/claude-helper/.claude/sounds/notification.aiff'

try:
    shutil.copy(source, dest)
    print(f"复制成功: {dest}")
except Exception as e:
    print(f"复制失败: {e}")

# 删除旧的wav文件
sounds_dir = '/Users/zhuxiaojiang/toys/claude-helper/.claude/sounds'
wav_files = ['attention.wav', 'bell.wav', 'chime.wav', 'complete.wav', 'error.wav', 'subtle.wav', 'success.wav']

for wav_file in wav_files:
    try:
        file_path = os.path.join(sounds_dir, wav_file)
        if os.path.exists(file_path):
            os.remove(file_path)
            print(f"删除文件: {wav_file}")
    except Exception as e:
        print(f"删除文件失败 {wav_file}: {e}")

print("音频文件处理完成")