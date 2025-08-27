#!/bin/bash
echo "[$(date)] Simple hook triggered" >> .claude/simple-hook.log
echo "Hook works!" > /tmp/hook-output.txt 2>/dev/null || echo "Hook works!" > hook-output.txt
exit 0