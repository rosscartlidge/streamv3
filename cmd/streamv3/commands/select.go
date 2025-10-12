package commands

import (
	"os"
	"context"
	"fmt"

	"github.com/rosscartlidge/gogstools/gs"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// SelectConfig holds configuration for select command
type SelectConfig struct {
	Field string `gs:"field,local,last,help=Field to select"`
	As    string `gs:"string,local,last,help=Rename field to (optional)"`
	Argv  string `gs:"file,global,last,help=Input JSONL file (or stdin if not specified),suffix=.jsonl"`
}

// selectCommand implements the select command
type selectCommand struct {
	config *SelectConfig
	cmd    *gs.GSCommand
}

func init() {
	RegisterCommand(newSelectCommand())
}

func newSelectCommand() *selectCommand {
	config := &SelectConfig{}
	cmd, err := gs.NewCommand(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create select command: %v", err))
	}

	return &selectCommand{
		config: config,
		cmd:    cmd,
	}
}

func (c *selectCommand) Name() string {
	return "select"
}

func (c *selectCommand) Description() string {
	return "Select and optionally rename fields"
}

func (c *selectCommand) GetGSCommand() *gs.GSCommand {
	return c.cmd
}

func (c *selectCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("select - Select and optionally rename fields")
		fmt.Println()
		fmt.Println("Usage: streamv3 select - field <name> [- as <newname>]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Select specific fields")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 select - field name - field age")
		fmt.Println()
		fmt.Println("  # Select and rename")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 select - field name - as fullname - field age")
		return nil
	}

	// Parse arguments using gs framework
	clauses, err := c.cmd.Parse(args)
	if err != nil {
		return fmt.Errorf("parsing arguments: %w", err)
	}

	// Build field mapping from clauses
	fieldMap := make(map[string]string) // original -> new name
	for _, clause := range clauses {
		field, _ := clause.Fields["Field"].(string)
		if field == "" {
			continue
		}

		// Check for rename
		if asName, ok := clause.Fields["As"].(string); ok && asName != "" {
			fieldMap[field] = asName
		} else {
			fieldMap[field] = field // Keep original name
		}
	}

	if len(fieldMap) == 0 {
		return fmt.Errorf("no fields specified")
	}

	// Read JSONL from stdin
	input, err := lib.OpenInput(c.config.Argv)
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
}

// Validate implements gs.Commander interface
func (c *SelectConfig) Validate() error {
	return nil
}

// Execute implements gs.Commander interface (not used, we use command.Execute)
func (c *SelectConfig) Execute(ctx context.Context, clauses []gs.ClauseSet) error {
	return nil
}
