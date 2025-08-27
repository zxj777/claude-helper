#!/usr/bin/env python3
import json
import sys
import os

try:
    # Read JSON input from stdin
    input_data = json.load(sys.stdin)
    
    # Extract prompt
    prompt = input_data.get('prompt', '')
    if not prompt:
        sys.exit(0)
    
    # Load text expansion config from project .claude/config directory
    config_file = '.claude/config/text-expander.json'
    if not os.path.exists(config_file):
        sys.exit(0)
    
    with open(config_file, 'r') as f:
        config = json.load(f)
    
    # Apply text expansions
    expanded_prompt = prompt
    for marker, replacement in config.get('mappings', {}).items():
        expanded_prompt = expanded_prompt.replace(marker, replacement)
    
    # If prompt changed, output expanded prompt as context and allow through
    if prompt != expanded_prompt:
        # Output expanded prompt as additional context for Claude
        print(f"用户的意思是: {expanded_prompt}")
        # Exit with code 0 to allow the original prompt through with added context
        sys.exit(0)
    
    # No change needed, allow original prompt through
    sys.exit(0)
        
except Exception as e:
    # On any error, allow original prompt through
    sys.exit(0)
