#!/usr/bin/env python3
import json
import sys
import os
import re

def apply_text_expansions_with_escape(text, mappings, escape_char='\\'):
    if not mappings:
        return text
    result = text
    for marker, replacement in mappings.items():
        pattern = r'(\\*)' + re.escape(marker)
        def replace_func(match):
            backslashes = match.group(1)
            backslash_count = len(backslashes)
            if backslash_count == 0:
                return replacement
            elif backslash_count % 2 == 1:
                return '\\' * (backslash_count // 2) + marker
            else:
                return '\\' * (backslash_count // 2) + replacement
        result = re.sub(pattern, replace_func, result)
    return result

try:
    input_data = json.load(sys.stdin)
    prompt = input_data.get('prompt', '')
    if not prompt:
        sys.exit(0)
    config_file = '.claude/config/text-expander.json'
    if not os.path.exists(config_file):
        sys.exit(0)
    with open(config_file, 'r') as f:
        config = json.load(f)
    escape_char = config.get('escape_char', '\\')
    mappings = config.get('mappings', {})
    expanded_prompt = apply_text_expansions_with_escape(prompt, mappings, escape_char)
    if prompt != expanded_prompt:
        print(f"鐢ㄦ埛鐨勬剰鎬濇槸: {expanded_prompt}")
        sys.exit(0)
    sys.exit(0)
except Exception as e:
    sys.exit(0)
