package commands

import (
	"context"
	"fmt"
	"iter"
	"os"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// readCSVCommand implements the read-csv command
type readCSVCommand struct {
	cmd *cf.Command
}

func init() {
	RegisterCommand(newReadCSVCommand())
}

func newReadCSVCommand() *readCSVCommand {
	var inputFile string
	var generate bool

	cmd := cf.NewCommand("read-csv").
		Description("Read CSV file and output JSONL stream").

		// -generate flag for code generation
		Flag("-generate", "-g").
			Bool().
			Bind(&generate).
			Global().
			Help("Generate Go code instead of executing").
			Done().

		// Positional argument for input file
		Flag("FILE").
			String().
			Completer(&cf.FileCompleter{Pattern: "*.csv"}).
			Bind(&inputFile).
			Global().
			Default("").
			Help("Input CSV file (or stdin if not specified)").
			Done().

		Handler(func(ctx *cf.Context) error {
			// If -generate flag is set, generate Go code instead of executing
			if generate {
				return generateReadCSVCode(inputFile)
			}

			// Normal execution: Read CSV from file or stdin
			var records iter.Seq[streamv3.Record]
			if inputFile == "" {
				// Read from stdin
				records = streamv3.ReadCSVFromReader(os.Stdin)
			} else {
				// Read from file
				var err error
				records, err = streamv3.ReadCSV(inputFile)
				if err != nil {
					return fmt.Errorf("reading CSV: %w", err)
				}
			}

			// Write as JSONL to stdout
			if err := lib.WriteJSONL(os.Stdout, records); err != nil {
				return fmt.Errorf("writing JSONL: %w", err)
			}

			return nil
		}).

		Build()

	return &readCSVCommand{
		cmd: cmd,
	}
}

func (c *readCSVCommand) Name() string {
	return "read-csv"
}

func (c *readCSVCommand) Description() string {
	return "Read CSV file and output JSONL stream"
}


func (c *readCSVCommand) GetCFCommand() *cf.Command {
	return c.cmd
}

func (c *readCSVCommand) Execute(ctx context.Context, args []string) error {
	// Delegate to completionflags command
	return c.cmd.Execute(args)
}

// generateReadCSVCode generates Go code for the read-csv command
func generateReadCSVCode(filename string) error {
	// Generate ReadCSV call with error handling
	var code string
	var imports []string

	if filename == "" {
		// Reading from stdin - use ReadCSVFromReader
		code = `records := streamv3.ReadCSVFromReader(os.Stdin)`
		imports = []string{"os"}
	} else {
		// Reading from file - use ReadCSV with error handling
		code = fmt.Sprintf(`records, err := streamv3.ReadCSV(%q)
	if err != nil {
		return fmt.Errorf("reading CSV: %%w", err)
	}`, filename)
		imports = []string{"fmt"}
	}

	// Create init fragment (first in pipeline)
	frag := lib.NewInitFragment("records", code, imports)

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}
