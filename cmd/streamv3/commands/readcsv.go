package commands

import (
	"context"
	"fmt"

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

func (c *readCSVCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag
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
		fmt.Println("  streamv3 read-csv data.csv | streamv3 where - field age - op gt - value 18")
		return nil
	}

	// Parse arguments using gs framework
	clauses, err := c.cmd.Parse(args)
	if err != nil {
		return fmt.Errorf("parsing arguments: %w", err)
	}

	// Get input file from config or first clause
	inputFile := c.config.Argv
	if inputFile == "" && len(clauses) > 0 {
		if argv, ok := clauses[0].Fields["Argv"].(string); ok {
			inputFile = argv
		}
	}

	// Read CSV file
	records := streamv3.ReadCSV(inputFile)

	// Write as JSONL to stdout
	if err := lib.WriteJSONL(lib.Stdout, records); err != nil {
		return fmt.Errorf("writing JSONL: %w", err)
	}

	// Suppress unused warning for clauses
	_ = clauses

	return nil
}

// Validate implements gs.Commander interface
func (c *ReadCSVConfig) Validate() error {
	return nil
}

// Execute implements gs.Commander interface (not used, we use command.Execute)
func (c *ReadCSVConfig) Execute(ctx context.Context, clauses []gs.ClauseSet) error {
	return nil
}
