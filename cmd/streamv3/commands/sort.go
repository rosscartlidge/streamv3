package commands

import (
	"os"
	"context"
	"fmt"
	"iter"

	"github.com/rosscartlidge/gogstools/gs"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// SortConfig holds configuration for sort command
type SortConfig struct {
	Field string `gs:"field,local,last,help=Field to sort by"`
	Desc  bool   `gs:"flag,local,last,help=Sort descending"`
	Argv  string `gs:"file,global,last,help=Input JSONL file (or stdin if not specified),suffix=.jsonl"`
}

// sortCommand implements the sort command
type sortCommand struct {
	config *SortConfig
	cmd    *gs.GSCommand
}

func init() {
	RegisterCommand(newSortCommand())
}

func newSortCommand() *sortCommand {
	config := &SortConfig{}
	cmd, err := gs.NewCommand(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create sort command: %v", err))
	}

	return &sortCommand{
		config: config,
		cmd:    cmd,
	}
}

func (c *sortCommand) Name() string {
	return "sort"
}

func (c *sortCommand) Description() string {
	return "Sort records by field"
}

func (c *sortCommand) GetGSCommand() *gs.GSCommand {
	return c.cmd
}

func (c *sortCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before gs framework takes over
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

	// Delegate to gs framework which will call Config.Execute
	return c.cmd.Execute(ctx, args)
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

// Validate implements gs.Commander interface
func (c *SortConfig) Validate() error {
	return nil
}

// Execute implements gs.Commander interface
// This is called by the gs framework after parsing arguments into clauses
func (c *SortConfig) Execute(ctx context.Context, clauses []gs.ClauseSet) error {
	// Get sort field from first clause
	var sortField string
	var descending bool
	if len(clauses) > 0 {
		sortField, _ = clauses[0].Fields["Field"].(string)
		descending, _ = clauses[0].Fields["Desc"].(bool)
	}

	if sortField == "" {
		return fmt.Errorf("no sort field specified")
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

	// Build sort key extractor and apply sort
	var result iter.Seq[streamv3.Record]
	if descending {
		// Descending: negate numeric values
		sorter := streamv3.SortBy(func(r streamv3.Record) float64 {
			val, _ := r[sortField]
			return -extractNumeric(val)
		})
		result = sorter(records)
	} else {
		// Ascending
		sorter := streamv3.SortBy(func(r streamv3.Record) float64 {
			val, _ := r[sortField]
			return extractNumeric(val)
		})
		result = sorter(records)
	}

	// Write output as JSONL
	if err := lib.WriteJSONL(os.Stdout, result); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}
