package commands

import (
	"os"
	"context"
	"fmt"

	"github.com/rosscartlidge/gogstools/gs"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// LimitConfig holds configuration for limit command
type LimitConfig struct {
	N    float64 `gs:"number,global,last,help=Number of records to take"`
	Argv string  `gs:"file,global,last,help=Input JSONL file (or stdin if not specified),suffix=.jsonl"`
}

// limitCommand implements the limit command
type limitCommand struct {
	config *LimitConfig
	cmd    *gs.GSCommand
}

func init() {
	RegisterCommand(newLimitCommand())
}

func newLimitCommand() *limitCommand {
	config := &LimitConfig{}
	cmd, err := gs.NewCommand(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create limit command: %v", err))
	}

	return &limitCommand{
		config: config,
		cmd:    cmd,
	}
}

func (c *limitCommand) Name() string {
	return "limit"
}

func (c *limitCommand) Description() string {
	return "Take first N records"
}

func (c *limitCommand) GetGSCommand() *gs.GSCommand {
	return c.cmd
}

func (c *limitCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before gs framework takes over
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("limit - Take first N records")
		fmt.Println()
		fmt.Println("Usage: streamv3 limit -n <count>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 limit -n 10")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 where -match age gt 18 | streamv3 limit -n 5")
		return nil
	}

	// Delegate to gs framework which will call Config.Execute
	return c.cmd.Execute(ctx, args)
}

// Validate implements gs.Commander interface
func (c *LimitConfig) Validate() error {
	return nil
}

// Execute implements gs.Commander interface
// This is called by the gs framework after parsing arguments into clauses
func (c *LimitConfig) Execute(ctx context.Context, clauses []gs.ClauseSet) error {
	// Get N from config or first clause
	n := int(c.N)
	if n == 0 && len(clauses) > 0 {
		if nVal, ok := clauses[0].Fields["N"].(float64); ok {
			n = int(nVal)
		}
	}

	if n <= 0 {
		return fmt.Errorf("limit must be positive, got %d", n)
	}

	// Get input file from Argv field or from bare arguments in clauses
	inputFile := c.Argv
	if inputFile == "" && len(clauses) > 0 {
		if argv, ok := clauses[0].Fields["Argv"].(string); ok && argv != "" {
			inputFile = argv
		}
		if inputFile == "" {
			if args, ok := clauses[0].Fields["_args"].([]string); ok && len(args) > 0 {
				inputFile = args[0]
			}
		}
	}

	// Read JSONL from stdin
	input, err := lib.OpenInput(inputFile)
	if err != nil {
		return err
	}
	defer input.Close()

	records := lib.ReadJSONL(input)

	// Apply limit
	limited := streamv3.Limit[streamv3.Record](n)(records)

	// Write output as JSONL
	if err := lib.WriteJSONL(os.Stdout, limited); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}
