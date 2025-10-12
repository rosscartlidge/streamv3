package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rosscartlidge/streamv3/cmd/streamv3/commands"
	_ "github.com/rosscartlidge/streamv3/cmd/streamv3/commands" // Import for init() functions
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
		fmt.Println("streamv3 version 0.1.0")
		os.Exit(0)
	}

	// Find and execute subcommand
	subcommand := os.Args[1]
	args := os.Args[2:]

	allCommands := commands.GetCommands()
	for _, cmd := range allCommands {
		if cmd.Name() == subcommand {
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
	fmt.Println("  -help        Show this help message")
	fmt.Println("  -version     Show version information")
	fmt.Println()
	fmt.Println("Use 'streamv3 <command> -help' for more information about a command.")
	fmt.Println()
	fmt.Println("Example pipelines:")
	fmt.Println("  cat data.csv | streamv3 read-csv | streamv3 where - field age - op gt - value 18 | streamv3 write-csv")
	fmt.Println("  streamv3 read-csv data.csv | streamv3 group-by - fields dept - agg 'count=count()' | streamv3 write-csv")
}

// FormatPipelineExample formats a multi-line pipeline for display
func FormatPipelineExample(lines []string) string {
	return strings.Join(lines, " \\\n    ")
}
