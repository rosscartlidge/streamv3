package commands

import (
	"context"
	"fmt"
	"os"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/gogstools/gs"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// selectCommand implements the select command
type selectCommand struct {
	cmd *cf.Command
}

func init() {
	RegisterCommand(newSelectCommand())
}

func newSelectCommand() *selectCommand {
	var inputFile string

	cmd := cf.NewCommand("select").
		Description("Select and optionally rename fields").
		Flag("-field", "-f").
			String().
			Local().
			Help("Field to select").
			Done().
		Flag("-as", "-a").
			String().
			Local().
			Help("Rename field to (optional)").
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
			// Build field mapping from clauses
			fieldMap := make(map[string]string) // original -> new name

			for _, clause := range ctx.Clauses {
				field, _ := clause.Flags["-field"].(string)
				if field == "" {
					continue
				}

				// Check for rename
				asName, _ := clause.Flags["-as"].(string)
				if asName != "" {
					fieldMap[field] = asName
				} else {
					fieldMap[field] = field // Keep original name
				}
			}

			if len(fieldMap) == 0 {
				return fmt.Errorf("no fields specified")
			}

			// Read JSONL from stdin or file
			input, err := lib.OpenInput(inputFile)
			if err != nil {
				return err
			}
			defer input.Close()

			records := lib.ReadJSONL(input)

			// Build selector function
			selector := func(r streamv3.Record) streamv3.Record {
				result := make(streamv3.Record)
				for origField, newField := range fieldMap {
					if val, exists := r[origField]; exists {
						result[newField] = val
					}
				}
				return result
			}

			// Apply selection
			selected := streamv3.Select(selector)(records)

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, selected); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Build()

	return &selectCommand{cmd: cmd}
}

func (c *selectCommand) Name() string {
	return "select"
}

func (c *selectCommand) Description() string {
	return "Select and optionally rename fields"
}

func (c *selectCommand) GetCFCommand() *cf.Command {
	return c.cmd
}

func (c *selectCommand) GetGSCommand() *gs.GSCommand {
	return nil // No longer using gs
}

func (c *selectCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before completionflags framework takes over
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("select - Select and optionally rename fields")
		fmt.Println()
		fmt.Println("Usage: streamv3 select -field <name> [-as <newname>]")
		fmt.Println()
		fmt.Println("Note: Use + to separate multiple field selections")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Select specific fields")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 select -field name + -field age")
		fmt.Println()
		fmt.Println("  # Select and rename")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 select -field name -as fullname + -field age")
		fmt.Println()
		fmt.Println("  # Select three fields")
		fmt.Println("  streamv3 select -field name + -field age + -field department")
		fmt.Println()
		fmt.Println("Debugging with jq:")
		fmt.Println("  # Inspect selected fields")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 select -field name + -field age | jq '.'")
		fmt.Println()
		fmt.Println("  # Extract single field values")
		fmt.Println("  streamv3 read-csv data.csv | jq -r '.name'")
		return nil
	}

	return c.cmd.Execute(args)
}
