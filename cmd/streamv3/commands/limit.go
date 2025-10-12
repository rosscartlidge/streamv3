package commands

import (
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

func (c *limitCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("limit - Take first N records")
		fmt.Println()
		fmt.Println("Usage: streamv3 limit - n <count>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 limit - n 10")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 where - field age - op gt - value 18 | streamv3 limit - n 5")
		return nil
	}

	// Parse arguments using gs framework
	clauses, err := c.cmd.Parse(args)
	if err != nil {
		return fmt.Errorf("parsing arguments: %w", err)
	}

	// Get N from config or first clause
	n := int(c.config.N)
	if n == 0 && len(clauses) > 0 {
		if nVal, ok := clauses[0].Fields["N"].(float64); ok {
			n = int(nVal)
		}
	}

	if n <= 0 {
		return fmt.Errorf("limit must be positive, got %d", n)
	}

	// Read JSONL from stdin
	input, err := lib.OpenInput(c.config.Argv)
	if err != nil {
		return err
	}
	defer input.Close()

	records := lib.ReadJSONL(input)

	// Apply limit
	limited := streamv3.Limit[streamv3.Record](n)(records)

	// Write output as JSONL
	if err := lib.WriteJSONL(lib.Stdout, limited); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}

// Validate implements gs.Commander interface
func (c *LimitConfig) Validate() error {
	return nil
}

// Execute implements gs.Commander interface (not used, we use command.Execute)
func (c *LimitConfig) Execute(ctx context.Context, clauses []gs.ClauseSet) error {
	return nil
}
