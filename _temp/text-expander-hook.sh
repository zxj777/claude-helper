#!/bin/bash

# Debug log
echo "Bash hook called" >> /tmp/claude-hook-debug.log

# Read JSON input from stdin
input_json=$(cat)

# Debug log input
echo "Input: $input_json" >> /tmp/claude-hook-debug.log

# Extract prompt from JSON using jq (fallback to basic parsing if jq not available)
if command -v jq >/dev/null 2>&1; then
    prompt=$(echo "$input_json" | jq -r '.prompt // empty')
else
    # Basic JSON parsing fallback - extract prompt value
    prompt=$(echo "$input_json" | sed -n 's/.*"prompt"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
fi

# Exit if no prompt
if [[ -z "$prompt" ]]; then
    exit 0
fi

# Load text expansion config
config_file="$HOME/.claude-helper/text-expander-config.json"
if [[ ! -f "$config_file" ]]; then
    exit 0
fi

# Read config and apply text expansions
expanded_prompt="$prompt"

if command -v jq >/dev/null 2>&1; then
    # Use jq to process mappings
    while IFS=$'\t' read -r marker replacement; do
        if [[ -n "$marker" && -n "$replacement" ]]; then
            expanded_prompt="${expanded_prompt//$marker/$replacement}"
        fi
    done < <(jq -r '.mappings // {} | to_entries[] | [.key, .value] | @tsv' "$config_file")
else
    # Basic fallback parsing for simple JSON mappings
    # This is a simplified approach and may not handle complex JSON
    while read -r line; do
        if [[ $line =~ \"([^\"]+)\"[[:space:]]*:[[:space:]]*\"([^\"]+)\" ]]; then
            marker="${BASH_REMATCH[1]}"
            replacement="${BASH_REMATCH[2]}"
            expanded_prompt="${expanded_prompt//$marker/$replacement}"
        fi
    done < <(grep -o '"[^"]*"[[:space:]]*:[[:space:]]*"[^"]*"' "$config_file")
fi

# Debug log
echo "Original: $prompt, Expanded: $expanded_prompt" >> /tmp/claude-hook-debug.log

# If prompt changed, output expanded prompt as context and allow through
if [[ "$prompt" != "$expanded_prompt" ]]; then
    echo "Adding expanded context: $expanded_prompt" >> /tmp/claude-hook-debug.log
    
    # Output expanded prompt as additional context for Claude
    echo "用户的意思是: $expanded_prompt"
    # Exit with code 0 to allow the original prompt through with added context
    exit 0
fi

# No change needed, allow original prompt through
exit 0