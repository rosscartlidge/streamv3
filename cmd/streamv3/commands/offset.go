package commands

import (
	"context"
	"fmt"
	"os"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// offsetCommand implements the offset command
type offsetCommand struct {
	cmd *cf.Command
}

func init() {
	RegisterCommand(newOffsetCommand())
}

func newOffsetCommand() *offsetCommand {
	var n int
	var inputFile string
	var generate bool

	cmd := cf.NewCommand("offset").
		Description("Skip first N records (SQL OFFSET)").
		Flag("-n").
			Int().
			Bind(&n).
			Global().
			Help("Number of records to skip").
			Done().
		Flag("-generate", "-g").
			Bool().
			Bind(&generate).
			Global().
			Help("Generate Go code instead of executing").
			Done().
		Flag("-file", "-f").
			String().
			Completer(&cf.FileCompleter{Pattern: "*.jsonl"}).
			Bind(&inputFile).
			Global().
			Default("").
			Help("Input JSONL file (or stdin if not specified)").
			Done().
		Handler(func(ctx *cf.Context) error {
			// If -generate flag is set, generate Go code instead of executing
			if generate {
				return generateOffsetCode(n, inputFile)
			}

			// Validate n
			if n < 0 {
				return fmt.Errorf("offset must be non-negative, got %d", n)
			}

			// Read JSONL from stdin or file
			input, err := lib.OpenInput(inputFile)
			if err != nil {
				return err
			}
			defer input.Close()

			records := lib.ReadJSONL(input)

			// Apply offset
			skipped := streamv3.Offset[streamv3.Record](n)(records)

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, skipped); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Build()

	return &offsetCommand{cmd: cmd}
}

func (c *offsetCommand) Name() string {
	return "offset"
}

func (c *offsetCommand) Description() string {
	return "Skip first N records (SQL OFFSET)"
}

func (c *offsetCommand) GetCFCommand() *cf.Command {
	return c.cmd
}

func (c *offsetCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before completionflags framework takes over
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("offset - Skip first N records (SQL OFFSET)")
		fmt.Println()
		fmt.Println("Usage: streamv3 offset -n <count>")
		fmt.Println()
		fmt.Println("Skips the first N records from the input stream.")
		fmt.Println("Commonly used with LIMIT for pagination.")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Skip first 20 records")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 offset -n 20")
		fmt.Println()
		fmt.Println("  # Pagination: skip 20, take 10 (records 21-30)")
		fmt.Println("  streamv3 read-csv data.csv | \\")
		fmt.Println("    streamv3 offset -n 20 | \\")
		fmt.Println("    streamv3 limit -n 10")
		fmt.Println()
		fmt.Println("  # SQL equivalent: LIMIT 10 OFFSET 20")
		fmt.Println("  streamv3 read-csv data.csv | \\")
		fmt.Println("    streamv3 offset -n 20 | \\")
		fmt.Println("    streamv3 limit -n 10")
		return nil
	}

	return c.cmd.Execute(args)
}

// generateOffsetCode generates Go code for the offset command
func generateOffsetCode(n int, inputFile string) error {
	// Read all previous code fragments from stdin (if any)
	fragments, err := lib.ReadAllCodeFragments()
	if err != nil {
		return fmt.Errorf("reading code fragments: %w", err)
	}

	// Pass through all previous fragments
	for _, frag := range fragments {
		if err := lib.WriteCodeFragment(frag); err != nil {
			return fmt.Errorf("writing previous fragment: %w", err)
		}
	}

	// Get input variable from last fragment
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}

	// Generate offset code
	outputVar := "skipped"
	code := fmt.Sprintf("%s := streamv3.Offset[streamv3.Record](%d)(%s)", outputVar, n, inputVar)

	// Create code fragment
	frag := lib.NewStmtFragment(outputVar, inputVar, code, nil)

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}
