#!/usr/bin/env python3
"""
Fix indentation of autocli fluent method calls in main.go

This script reformats cmd/ssql/main.go to show the hierarchical structure
of autocli builder calls with proper indentation:

- Subcommand() at base level
- Description/Example/Flag/Handler indented under subcommand
- Flag properties (Bool/Int/Help/etc) indented under Flag
- Done() aligned with its opening call
- Handler function bodies preserve Go indentation

Usage:
    python3 scripts/fix-fluent-indent.py cmd/ssql/main.go

The script modifies the file in-place. Run 'go build' after to verify
the reformatted code still compiles correctly.
"""
import re

def should_indent_after(line):
    """Check if next line should be indented"""
    stripped = line.strip()
    return (stripped.startswith('Subcommand(') or 
            stripped.startswith('Flag(') or
            stripped.startswith('Arg('))

def should_dedent_before(line):
    """Check if this line should dedent"""
    stripped = line.strip()
    return stripped.startswith('Done().')

def is_handler_line(line):
    """Check if this is part of a Handler function"""
    return 'Handler(func' in line

def fix_builder_indentation(lines, base_indent=2):
    """Fix indentation for autocli builder calls"""
    result = []
    indent_level = base_indent
    in_handler = False
    handler_indent = 0
    brace_count = 0
    
    for line in lines:
        stripped = line.lstrip()
        
        # Skip empty lines - preserve exactly
        if not stripped:
            result.append('')
            continue
        
        # Skip comments - apply current indent
        if stripped.startswith('//'):
            result.append('\t' * indent_level + stripped)
            continue
        
        # Detect Handler function start
        if is_handler_line(stripped) and not in_handler:
            in_handler = True
            handler_indent = indent_level
            brace_count = 0
            result.append('\t' * indent_level + stripped)
            brace_count = stripped.count('{') - stripped.count('}')
            continue
        
        # Inside handler function - preserve original indentation relative to handler start
        if in_handler:
            # Track braces
            brace_count += stripped.count('{') - stripped.count('}')
            
            # Check if we're at the end of handler (closing }).
            if stripped == '}).':
                in_handler = False
                result.append('\t' * indent_level + stripped)
                continue
            
            # Preserve the line's indentation for Go code
            result.append(line.rstrip())
            continue
        
        # Handle Done() - dedent before adding
        if should_dedent_before(stripped):
            indent_level = max(base_indent, indent_level - 1)
            result.append('\t' * indent_level + stripped)
            continue
        
        # Add line at current indent
        result.append('\t' * indent_level + stripped)
        
        # Check if we should indent after this line
        if should_indent_after(stripped):
            indent_level += 1
    
    return result

def main():
    import sys
    
    if len(sys.argv) != 2:
        print("Usage: fix_fluent_indent_v2.py <file>")
        sys.exit(1)
    
    filename = sys.argv[1]
    
    with open(filename, 'r') as f:
        lines = [line.rstrip('\n') for line in f]
    
    # Find buildRootCommand function
    in_function = False
    output = []
    func_lines = []
    skip_next = False
    
    for i, line in enumerate(lines):
        if skip_next:
            skip_next = False
            continue
            
        if 'func buildRootCommand()' in line:
            in_function = True
            output.append(line)
            # Next line should be the return statement
            if i + 1 < len(lines):
                output.append(lines[i+1])
                skip_next = True
            continue
        
        if in_function:
            # End of function
            if line.startswith('}') and not '}).' in line:
                # Process accumulated lines
                fixed = fix_builder_indentation(func_lines)
                output.extend(fixed)
                output.append(line)
                in_function = False
                func_lines = []
            else:
                func_lines.append(line)
        else:
            output.append(line)
    
    # Write result
    with open(filename, 'w') as f:
        for line in output:
            f.write(line + '\n')
    
    print(f"Fixed indentation in {filename}")

if __name__ == '__main__':
    main()
