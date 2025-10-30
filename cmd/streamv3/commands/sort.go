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

// sortCommand implements the sort command
type sortCommand struct {
	cmd *cf.Command
}

func init() {
	RegisterCommand(newSortCommand())
}

func newSortCommand() *sortCommand {
	var field string
	var desc bool
	var inputFile string
	var generate bool

	cmd := cf.NewCommand("sort").
		Description("Sort records by field").
		Flag("-field", "-f").
			String().
			Completer(cf.NoCompleter{Hint: "<field-name>"}).
			Bind(&field).
			Local().
			Help("Field to sort by").
			Done().
		Flag("-desc", "-d").
			Bool().
			Bind(&desc).
			Local().
			Help("Sort descending").
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
				return generateSortCode(field, desc, inputFile)
			}

			if field == "" {
				return fmt.Errorf("no sort field specified")
			}

			// Read JSONL from stdin or file
			input, err := lib.OpenInput(inputFile)
			if err != nil {
				return err
			}
			defer input.Close()

			records := lib.ReadJSONL(input)

			// Build sort key extractor and apply sort
			var result iter.Seq[streamv3.Record]
			if desc {
				// Descending: negate numeric values
				sorter := streamv3.SortBy(func(r streamv3.Record) float64 {
					val, _ := r[field]
					return -extractNumeric(val)
				})
				result = sorter(records)
			} else {
				// Ascending
				sorter := streamv3.SortBy(func(r streamv3.Record) float64 {
					val, _ := r[field]
					return extractNumeric(val)
				})
				result = sorter(records)
			}

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, result); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Build()

	return &sortCommand{cmd: cmd}
}

func (c *sortCommand) Name() string {
	return "sort"
}

func (c *sortCommand) Description() string {
	return "Sort records by field"
}

func (c *sortCommand) GetCFCommand() *cf.Command {
	return c.cmd
}


func (c *sortCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before completionflags framework takes over
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("sort - Sort records by field")
		fmt.Println()
		fmt.Println("Usage: streamv3 sort -field <name> [-desc]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Sort ascending")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 sort -field age")
		fmt.Println()
		fmt.Println("  # Sort descending")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 sort -field age -desc")
		return nil
	}

	return c.cmd.Execute(args)
}

// extractNumeric extracts a numeric value for sorting
func extractNumeric(val any) float64 {
	switch v := val.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	case string:
		// For strings, use 0 (they'll maintain relative order)
		return 0
	default:
		return 0
	}
}

// generateSortCode generates Go code for the sort command
func generateSortCode(field string, desc bool, inputFile string) error {
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

	// Generate sort code
	outputVar := "sorted"
	var code string
	if desc {
		// Descending sort
		code = fmt.Sprintf("%s := streamv3.SortBy(func(r streamv3.Record) float64 {\n", outputVar)
		code += fmt.Sprintf("\t\tval, _ := r[%q]\n", field)
		code += "\t\tswitch v := val.(type) {\n"
		code += "\t\tcase int64:\n"
		code += "\t\t\treturn -float64(v)\n"
		code += "\t\tcase float64:\n"
		code += "\t\t\treturn -v\n"
		code += "\t\tdefault:\n"
		code += "\t\t\treturn 0\n"
		code += "\t\t}\n"
		code += fmt.Sprintf("\t})(%s)", inputVar)
	} else {
		// Ascending sort
		code = fmt.Sprintf("%s := streamv3.SortBy(func(r streamv3.Record) float64 {\n", outputVar)
		code += fmt.Sprintf("\t\tval, _ := r[%q]\n", field)
		code += "\t\tswitch v := val.(type) {\n"
		code += "\t\tcase int64:\n"
		code += "\t\t\treturn float64(v)\n"
		code += "\t\tcase float64:\n"
		code += "\t\t\treturn v\n"
		code += "\t\tdefault:\n"
		code += "\t\t\treturn 0\n"
		code += "\t\t}\n"
		code += fmt.Sprintf("\t})(%s)", inputVar)
	}

	// Create code fragment
	frag := lib.NewStmtFragment(outputVar, inputVar, code, nil)

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}
