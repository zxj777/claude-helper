#!/usr/bin/env python3
import json
import sys
import os
import re

def apply_text_expansions_with_escape(text, mappings, escape_char='\\'):
    r"""Apply text expansions with support for escape characters.
    
    Rules:
    - \marker -> literal marker (no expansion)
    - \\marker -> literal \ + expand marker  
    - \\\marker -> literal \ + literal marker
    - \\\\marker -> literal \\ + expand marker
    """
    if not mappings:
        return text
    
    result = text
    
    # Process each mapping
    for marker, replacement in mappings.items():
        # Create a pattern that matches the marker with potential escaping
        # We need to handle sequences of backslashes before the marker
        pattern = r'(\\*)' + re.escape(marker)
        
        def replace_func(match):
            backslashes = match.group(1)
            backslash_count = len(backslashes)
            
            if backslash_count == 0:
                # No backslashes, normal expansion
                return replacement
            elif backslash_count % 2 == 1:
                # Odd number of backslashes: last one escapes the marker
                # Return half the backslashes (rounded down) + literal marker
                return '\\' * (backslash_count // 2) + marker
            else:
                # Even number of backslashes: marker is not escaped
                # Return half the backslashes + expanded marker
                return '\\' * (backslash_count // 2) + replacement
        
        result = re.sub(pattern, replace_func, result)
    
    return result

try:
    # Read JSON input from stdin
    input_data = json.load(sys.stdin)
    
    # Extract prompt
    prompt = input_data.get('prompt', '')
    
    # Clean the prompt to remove invalid Unicode characters
    try:
        # Encode and decode to clean up any invalid UTF-8 characters
        prompt = prompt.encode('utf-8', errors='ignore').decode('utf-8')
    except:
        # If that fails, use ascii encoding as fallback
        prompt = prompt.encode('ascii', errors='ignore').decode('ascii')
    
    if not prompt:
        sys.exit(0)
    
    # Load text expansion config from project .claude/config directory
    config_file = '.claude/config/text-expander.json'
    if not os.path.exists(config_file):
        sys.exit(0)
    
    with open(config_file, 'r', encoding='utf-8') as f:
        config = json.load(f)
    
    # Get escape character (default to backslash)
    escape_char = config.get('escape_char', '\\')
    mappings = config.get('mappings', {})
    
    # Apply text expansions with escape support
    expanded_prompt = apply_text_expansions_with_escape(prompt, mappings, escape_char)
    
    # If prompt changed, add expanded text as additional context
    if prompt != expanded_prompt:
        result = {
            "hookSpecificOutput": {
                "hookEventName": "UserPromptSubmit",
                "additionalContext": f"用户的意思是: {expanded_prompt}"
            }
        }
        print(json.dumps(result, ensure_ascii=True), flush=True)
        sys.exit(0)
    
    # No change needed, allow original prompt through
    sys.exit(0)
        
except Exception as e:
    # On any error, allow original prompt through
    # For debugging, could log error to a file
    try:
        with open('.claude/hook-error.log', 'a', encoding='utf-8', errors='replace') as f:
            import traceback
            f.write(f"Text-expander error: {type(e).__name__}: {str(e)}\\n")
            f.write(f"Traceback: {traceback.format_exc()}\\n")
    except:
        pass  # Ignore logging errors
    sys.exit(0)
