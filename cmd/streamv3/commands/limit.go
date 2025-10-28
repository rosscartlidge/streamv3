package commands

import (
	"context"
	"fmt"
	"os"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// limitCommand implements the limit command
type limitCommand struct {
	cmd *cf.Command
}

func init() {
	RegisterCommand(newLimitCommand())
}

func newLimitCommand() *limitCommand {
	var n int
	var inputFile string

	cmd := cf.NewCommand("limit").
		Description("Take first N records").
		Flag("-n").
			Int().
			Bind(&n).
			Global().
			Help("Number of records to take").
			Done().
		Flag("FILE").
			String().
			Completer(&cf.FileCompleter{Pattern: "*.jsonl"}).
			Bind(&inputFile).
			Global().
			Default("").
			Help("Input JSONL file (or stdin if not specified)").
			Done().
		Handler(func(ctx *cf.Context) error {
			if n <= 0 {
				return fmt.Errorf("limit must be positive, got %d", n)
			}

			// Read JSONL from stdin or file
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
		}).
		Build()

	return &limitCommand{cmd: cmd}
}

func (c *limitCommand) Name() string {
	return "limit"
}

func (c *limitCommand) Description() string {
	return "Take first N records"
}

func (c *limitCommand) GetCFCommand() *cf.Command {
	return c.cmd
}

func (c *limitCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before completionflags framework takes over
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("limit - Take first N records")
		fmt.Println()
		fmt.Println("Usage: streamv3 limit -n <count> [file.jsonl]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 limit -n 10")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 where -match age gt 18 | streamv3 limit -n 5")
		return nil
	}

	return c.cmd.Execute(args)
}
