#!/usr/bin/env python3
import json
import sys
import os

# Debug log
with open('/tmp/claude-hook-debug.log', 'a') as f:
    f.write(f"Python hook called\n")

try:
    # Read JSON input from stdin
    input_data = json.load(sys.stdin)
    
    with open('/tmp/claude-hook-debug.log', 'a') as f:
        f.write(f"Input: {json.dumps(input_data)}\n")
    
    # Extract prompt
    prompt = input_data.get('prompt', '')
    if not prompt:
        sys.exit(0)
    
    # Load text expansion config
    config_file = os.path.expanduser('~/.claude-helper/text-expander-config.json')
    if not os.path.exists(config_file):
        sys.exit(0)
    
    with open(config_file, 'r') as f:
        config = json.load(f)
    
    # Apply text expansions
    expanded_prompt = prompt
    for marker, replacement in config.get('mappings', {}).items():
        expanded_prompt = expanded_prompt.replace(marker, replacement)
    
    with open('/tmp/claude-hook-debug.log', 'a') as f:
        f.write(f"Original: {prompt}, Expanded: {expanded_prompt}\n")
    
    # If prompt changed, output expanded prompt as context and allow through
    if prompt != expanded_prompt:
        with open('/tmp/claude-hook-debug.log', 'a') as f:
            f.write(f"Adding expanded context: {expanded_prompt}\n")
        
        # Output expanded prompt as additional context for Claude
        print(f"用户的意思是: {expanded_prompt}")
        # Exit with code 0 to allow the original prompt through with added context
        sys.exit(0)
    
    # No change needed, allow original prompt through
    sys.exit(0)
        
except Exception as e:
    with open('/tmp/claude-hook-debug.log', 'a') as f:
        f.write(f"Error: {str(e)}\n")
    sys.exit(0)

