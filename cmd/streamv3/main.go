package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rosscartlidge/streamv3/cmd/streamv3/commands"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/version"
)

func main() {
	ctx := context.Background()

	// Show help if no arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	// Handle global flags
	switch os.Args[1] {
	case "-help", "--help", "help":
		printUsage()
		os.Exit(0)
	case "-version", "--version", "version":
		fmt.Printf("streamv3 version %s\n", version.Version)
		os.Exit(0)
	case "-completion-script", "-bash-completion":
		// Support both -completion-script (completionflags convention) and -bash-completion (legacy)
		printBashCompletion()
		os.Exit(0)
	}

	// Find and execute subcommand
	subcommand := os.Args[1]
	args := os.Args[2:]

	allCommands := commands.GetCommands()
	for _, cmd := range allCommands {
		if cmd.Name() == subcommand {
			// Call our custom Execute which handles -help and delegates to completionflags framework
			if err := cmd.Execute(ctx, args); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	// Command not found
	fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", subcommand)
	printUsage()
	os.Exit(1)
}

func printUsage() {
	fmt.Println("streamv3 - Unix-style data processing tools")
	fmt.Println()
	fmt.Println("Usage: streamv3 <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")

	for _, cmd := range commands.GetCommands() {
		fmt.Printf("  %-12s %s\n", cmd.Name(), cmd.Description())
	}

	fmt.Println()
	fmt.Println("Global Flags:")
	fmt.Println("  -help              Show this help message")
	fmt.Println("  -version           Show version information")
	fmt.Println("  -completion-script Generate bash completion script")
	fmt.Println()
	fmt.Println("Use 'streamv3 <command> -help' for more information about a command.")
	fmt.Println()
	fmt.Println("Bash Completion Setup:")
	fmt.Println("  eval \"$(streamv3 -completion-script)\"  # For current session")
	fmt.Println("  streamv3 -completion-script > ~/.local/share/bash-completion/completions/streamv3  # Persistent")
	fmt.Println()
	fmt.Println("Note: -bash-completion is also supported as an alias for -completion-script")
	fmt.Println()
	fmt.Println("Example pipelines:")
	fmt.Println("  # Filter by age and export")
	fmt.Println("  streamv3 read-csv data.csv | streamv3 where -match age gt 18 | streamv3 write-csv output.csv")
	fmt.Println()
	fmt.Println("  # Select fields, sort, and limit")
	fmt.Println("  streamv3 read-csv data.csv | streamv3 select -field name + -field salary | streamv3 sort -field salary -desc | streamv3 limit -n 10")
	fmt.Println()
	fmt.Println("  # Complex filter with AND/OR")
	fmt.Println("  streamv3 read-csv data.csv | streamv3 where -match age gt 30 -match dept eq Engineering + -match salary gt 100000")
}

// FormatPipelineExample formats a multi-line pipeline for display
func FormatPipelineExample(lines []string) string {
	return strings.Join(lines, " \\\n    ")
}

// printBashCompletion outputs bash completion script
func printBashCompletion() {
	// Get the command name (e.g., "streamv3" or "s" if aliased)
	cmdName := os.Args[0]

	// Generate completion script with dynamic command name
	fmt.Printf(`# Bash completion for %s
_streamv3_completion() {
    local cur prev words cword
    _init_completion || return

    # If we're completing the first argument (command name)
    if [ "$cword" -eq 1 ]; then
        local commands="read-csv write-csv where select limit offset sort distinct join union group chart exec generate-go help -help --help -version --version -completion-script -bash-completion"
        COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
        return 0
    fi

    # Get the subcommand
    local subcommand="${words[1]}"

    # For subcommands, delegate to completionflags framework completion
    # The subcommand itself handles -complete via completionflags framework
    # Pass position and all arguments after the subcommand name
    # Adjust cword to be relative to subcommand (subtract 1 for the command name)
    local completions=$(%s "$subcommand" -complete $((cword - 1)) "${words[@]:2}" 2>/dev/null)

    if [ -n "$completions" ]; then
        COMPREPLY=( $(compgen -W "$completions" -- "$cur") )
    else
        # Fallback to file completion
        _filedir
    fi
}

complete -F _streamv3_completion %s`, cmdName, cmdName, cmdName)
	fmt.Println() // Add trailing newline
}
