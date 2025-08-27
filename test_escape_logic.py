#!/usr/bin/env python3
import re

def apply_text_expansions_with_escape(text, mappings, escape_char='\\'):
    """Apply text expansions with support for escape characters.
    
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

# Test cases
def test_escape_logic():
    mappings = {
        "-e": "expanded",
        "-d": "debug mode",
        "-v": "verbose"
    }
    
    test_cases = [
        # Basic cases
        ("-e", "expanded"),
        ("hello -e world", "hello expanded world"),
        
        # Single escape cases  
        ("\\-e", "-e"),
        ("hello \\-e world", "hello -e world"),
        
        # Double escape cases
        ("\\\\-e", "\\expanded"),
        ("hello \\\\-e world", "hello \\expanded world"),
        
        # Triple escape cases
        ("\\\\\\-e", "\\-e"),
        ("hello \\\\\\-e world", "hello \\-e world"),
        
        # Quadruple escape cases
        ("\\\\\\\\-e", "\\\\expanded"),
        ("hello \\\\\\\\-e world", "hello \\\\expanded world"),
        
        # Multiple markers
        ("-e and -v", "expanded and verbose"),
        ("\\-e and -v", "-e and verbose"),
        ("\\-e and \\-v", "-e and -v"),
        
        # Edge cases
        ("-e-v", "expanded-v"),  # Only first marker matches
        ("no markers here", "no markers here"),
        ("", ""),
    ]
    
    print("Testing escape logic:")
    print("=" * 50)
    
    for i, (input_text, expected) in enumerate(test_cases):
        result = apply_text_expansions_with_escape(input_text, mappings)
        status = "✅" if result == expected else "❌"
        
        print(f"Test {i+1:2}: {status}")
        print(f"  Input:    '{input_text}'")
        print(f"  Expected: '{expected}'")
        print(f"  Got:      '{result}'")
        if result != expected:
            print(f"  ❌ MISMATCH!")
        print()

if __name__ == "__main__":
    test_escape_logic()