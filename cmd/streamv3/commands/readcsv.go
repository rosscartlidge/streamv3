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
	Generate bool   `gs:"flag,global,last,help=Generate Go code instead of executing"`
	Argv     string `gs:"file,global,last,help=Input CSV file (or stdin if not specified),suffix=.csv"`
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
		fmt.Println("  streamv3 read-csv data.csv | streamv3 where -match age gt 18")
		fmt.Println()
		fmt.Println("Debugging with jq:")
		fmt.Println("  # Inspect parsed CSV data")
		fmt.Println("  streamv3 read-csv data.csv | jq '.' | head -5")
		fmt.Println()
		fmt.Println("  # Check field names")
		fmt.Println("  streamv3 read-csv data.csv | head -1 | jq 'keys'")
		fmt.Println()
		fmt.Println("  # Verify field types")
		fmt.Println("  streamv3 read-csv data.csv | head -1 | jq 'to_entries | map({key, type: .value | type})'")
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

	// If -generate flag is set, generate Go code instead of executing
	if c.Generate {
		return c.generateCode(inputFile)
	}

	// Normal execution: Read CSV file (empty string means stdin)
	records, err := streamv3.ReadCSV(inputFile)
	if err != nil {
		return fmt.Errorf("reading CSV: %w", err)
	}

	// Write as JSONL to stdout
	if err := lib.WriteJSONL(os.Stdout, records); err != nil {
		return fmt.Errorf("writing JSONL: %w", err)
	}

	return nil
}

// generateCode generates Go code for the read-csv command
func (c *ReadCSVConfig) generateCode(filename string) error {
	// Generate ReadCSV call with error handling
	var code string
	if filename == "" {
		code = `records, err := streamv3.ReadCSV("")
	if err != nil {
		return fmt.Errorf("reading CSV: %w", err)
	}`
	} else {
		code = fmt.Sprintf(`records, err := streamv3.ReadCSV(%q)
	if err != nil {
		return fmt.Errorf("reading CSV: %%w", err)
	}`, filename)
	}

	// Create init fragment (first in pipeline)
	frag := lib.NewInitFragment("records", code, []string{"fmt"})

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}
