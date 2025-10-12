package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/rosscartlidge/gogstools/gs"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// ReadCSVConfig holds configuration for read-csv command
type ReadCSVConfig struct {
	Argv string `gs:"file,global,last,help=Input CSV file (or stdin if not specified),suffix=.csv"`
}

// readCSVCommand implements the read-csv command
type readCSVCommand struct {
	config *ReadCSVConfig
	cmd    *gs.GSCommand
}

func init() {
	RegisterCommand(newReadCSVCommand())
}

func newReadCSVCommand() *readCSVCommand {
	config := &ReadCSVConfig{}
	cmd, err := gs.NewCommand(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create read-csv command: %v", err))
	}

	return &readCSVCommand{
		config: config,
		cmd:    cmd,
	}
}

func (c *readCSVCommand) Name() string {
	return "read-csv"
}

func (c *readCSVCommand) Description() string {
	return "Read CSV file and output JSONL stream"
}

func (c *readCSVCommand) GetGSCommand() *gs.GSCommand {
	return c.cmd
}

func (c *readCSVCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before gs framework takes over
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("read-csv - Read CSV file and output JSONL stream")
		fmt.Println()
		fmt.Println("Usage: streamv3 read-csv [file.csv]")
		fmt.Println()
		fmt.Println("Reads a CSV file (or stdin) and outputs JSONL (JSON Lines) format.")
		fmt.Println("The first row is treated as the header with field names.")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  streamv3 read-csv data.csv")
		fmt.Println("  cat data.csv | streamv3 read-csv")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 where - match age gt 18")
		return nil
	}

	// Delegate to gs framework which will call Config.Execute
	return c.cmd.Execute(ctx, args)
}

// Validate implements gs.Commander interface
func (c *ReadCSVConfig) Validate() error {
	return nil
}

// Execute implements gs.Commander interface
// This is called by the gs framework after parsing arguments into clauses
func (c *ReadCSVConfig) Execute(ctx context.Context, clauses []gs.ClauseSet) error {
	// Get input file from Argv field or from bare arguments in clauses
	inputFile := c.Argv

	// If Argv not set, check for bare arguments in _args field
	if inputFile == "" && len(clauses) > 0 {
		// Check for Argv in the clause (might be set by -argv flag)
		if argv, ok := clauses[0].Fields["Argv"].(string); ok && argv != "" {
			inputFile = argv
		}
		// Check for bare arguments in _args
		if inputFile == "" {
			if args, ok := clauses[0].Fields["_args"].([]string); ok && len(args) > 0 {
				inputFile = args[0]
			}
		}
	}

	// Read CSV file (empty string means stdin)
	records := streamv3.ReadCSV(inputFile)

	// Write as JSONL to stdout
	if err := lib.WriteJSONL(os.Stdout, records); err != nil {
		return fmt.Errorf("writing JSONL: %w", err)
	}

	return nil
}
