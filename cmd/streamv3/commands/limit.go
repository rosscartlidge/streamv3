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
	var generate bool

	cmd := cf.NewCommand("limit").
		Description("Take first N records").
		Flag("-n").
			Int().
			Bind(&n).
			Global().
			Help("Number of records to take").
			Done().
		Flag("-generate", "-g").
			Bool().
			Bind(&generate).
			Global().
			Help("Generate Go code instead of executing").
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
			// If -generate flag is set, generate Go code instead of executing
			if generate {
				return generateLimitCode(n, inputFile)
			}

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

// generateLimitCode generates Go code for the limit command
func generateLimitCode(n int, inputFile string) error {
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

	// Generate limit code
	outputVar := "limited"
	code := fmt.Sprintf("%s := streamv3.Limit[streamv3.Record](%d)(%s)", outputVar, n, inputVar)

	// Create code fragment
	frag := lib.NewStmtFragment(outputVar, inputVar, code, nil)

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}
