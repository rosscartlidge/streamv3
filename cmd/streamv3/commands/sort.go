package commands

import (
	"context"
	"fmt"
	"iter"
	"os"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/gogstools/gs"
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

	cmd := cf.NewCommand("sort").
		Description("Sort records by field").
		Flag("-field", "-f").
			String().
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
		Flag("FILE").
			String().
			Bind(&inputFile).
			Global().
			Default("").
			FilePattern("*.jsonl").
			Help("Input JSONL file (or stdin if not specified)").
			Done().
		Handler(func(ctx *cf.Context) error {
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

func (c *sortCommand) GetGSCommand() *gs.GSCommand {
	return nil // No longer using gs
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
