#!/bin/bash

# Try different Python commands
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
