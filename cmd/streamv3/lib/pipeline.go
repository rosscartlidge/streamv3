package lib

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Command represents a single CLI command in a pipeline
type Command struct {
	Name   string            // "read-csv", "where", "limit"
	Args   map[string]string // Flag values: {"field": "age", "op": "gt", "value": "18"}
	Files  []string          // Positional file arguments
	Stdin  bool              // True if reads from stdin
	Stdout bool              // True if writes to stdout
}

// Pipeline represents a sequence of commands connected by pipes
type Pipeline struct {
	Commands []Command
}

// ParsePipeline parses a shell pipeline from input (one pipeline per line)
func ParsePipeline(r io.Reader) (*Pipeline, error) {
	scanner := bufio.NewScanner(r)
	var lines []string

	// Read all lines and concatenate
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("no pipeline found in input")
	}

	// Join lines (handle backslash continuation)
	pipelineStr := strings.Join(lines, " ")
	pipelineStr = strings.ReplaceAll(pipelineStr, "\\", "")

	// Split by | and parse each command
	parts := strings.Split(pipelineStr, "|")
	commands := make([]Command, 0, len(parts))

	for i, part := range parts {
		cmd, err := parseCommand(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("parsing command %d: %w", i+1, err)
		}

		// First command reads from stdin (unless it has a file arg)
		if i == 0 && len(cmd.Files) == 0 {
			cmd.Stdin = true
		} else if i > 0 {
			cmd.Stdin = true
		}

		// Last command writes to stdout
		if i == len(parts)-1 {
			cmd.Stdout = true
		}

		commands = append(commands, cmd)
	}

	return &Pipeline{Commands: commands}, nil
}

// parseCommand parses a single command like:
// "streamv3 where -field age -op gt -value 18"
// "streamv3 read-csv data.csv"
func parseCommand(cmdLine string) (Command, error) {
	fields := tokenize(cmdLine)
	if len(fields) < 2 {
		return Command{}, fmt.Errorf("invalid command: %s", cmdLine)
	}

	// Check if first token is "streamv3"
	if fields[0] != "streamv3" {
		return Command{}, fmt.Errorf("command must start with 'streamv3': %s", cmdLine)
	}

	cmd := Command{
		Name: fields[1],
		Args: make(map[string]string),
	}

	// Parse flags and positional arguments
	i := 2
	for i < len(fields) {
		token := fields[i]

		if strings.HasPrefix(token, "-") {
			// It's a flag
			flagName := strings.TrimPrefix(token, "-")

			// Check if next token is the value
			if i+1 < len(fields) && !strings.HasPrefix(fields[i+1], "-") {
				cmd.Args[flagName] = fields[i+1]
				i += 2
			} else {
				// Boolean flag
				cmd.Args[flagName] = "true"
				i++
			}
		} else {
			// Positional argument (file)
			cmd.Files = append(cmd.Files, token)
			i++
		}
	}

	return cmd, nil
}

// tokenize splits command line respecting quotes
func tokenize(s string) []string {
	var tokens []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, ch := range s {
		switch {
		case ch == '"' || ch == '\'':
			if inQuote {
				if ch == quoteChar {
					// End quote
					inQuote = false
					quoteChar = 0
				} else {
					current.WriteRune(ch)
				}
			} else {
				// Start quote
				inQuote = true
				quoteChar = ch
			}

		case ch == ' ' || ch == '\t':
			if inQuote {
				current.WriteRune(ch)
			} else if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}

		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// GetInputFile returns the input file for a command, or empty string for stdin
func (c *Command) GetInputFile() string {
	if len(c.Files) > 0 {
		return c.Files[0]
	}
	return ""
}

// GetOutputFile returns the output file for a command, or empty string for stdout
func (c *Command) GetOutputFile() string {
	if len(c.Files) > 0 {
		// For write commands, the file is usually the first/only arg
		return c.Files[0]
	}
	return ""
}

// String returns a string representation for debugging
func (c *Command) String() string {
	var parts []string
	parts = append(parts, "streamv3", c.Name)

	for k, v := range c.Args {
		parts = append(parts, fmt.Sprintf("-%s %s", k, v))
	}

	for _, f := range c.Files {
		parts = append(parts, f)
	}

	return strings.Join(parts, " ")
}

// String returns a string representation of the pipeline
func (p *Pipeline) String() string {
	var parts []string
	for _, cmd := range p.Commands {
		parts = append(parts, cmd.String())
	}
	return strings.Join(parts, " | ")
}
