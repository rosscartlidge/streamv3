package commands

import (
	"context"
	"fmt"
	"os"

	cf "github.com/rosscartlidge/completionflags"
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
	var generate bool

	cmd := cf.NewCommand("select").
		Description("Select and optionally rename fields").
		Flag("-field", "-f").
			String().
			Completer(cf.NoCompleter{Hint: "<field-name>"}).
			Local().
			Help("Field to select").
			Done().
		Flag("-as", "-a").
			String().
			Completer(cf.NoCompleter{Hint: "<new-name>"}).
			Local().
			Help("Rename field to (optional)").
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
			if shouldGenerate(generate) {
				return generateSelectCode(ctx, inputFile)
			}

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
				result := streamv3.MakeMutableRecord()
				for origField, newField := range fieldMap {
					if val, exists := streamv3.Get[any](r, origField); exists {
						result = result.SetAny(newField, val)
					}
				}
				return result.Freeze()
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

// generateSelectCode generates Go code for the select command
func generateSelectCode(ctx *cf.Context, inputFile string) error {
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

	// Build field mapping from clauses
	fieldMap := make(map[string]string) // original -> new name
	for _, clause := range ctx.Clauses {
		field, _ := clause.Flags["-field"].(string)
		if field == "" {
			continue
		}
		asName, _ := clause.Flags["-as"].(string)
		if asName != "" {
			fieldMap[field] = asName
		} else {
			fieldMap[field] = field // Keep original name
		}
	}

	// Generate select code
	outputVar := "selected"
	var code string
	code = fmt.Sprintf("%s := streamv3.Select(func(r streamv3.Record) streamv3.Record {\n", outputVar)
	code += "\t\tresult := streamv3.MakeMutableRecord()\n"
	for origField, newField := range fieldMap {
		if origField == newField {
			// No rename
			code += fmt.Sprintf("\t\tif val, exists := streamv3.Get[any](r, %q); exists {\n", origField)
			code += fmt.Sprintf("\t\t\tresult = result.SetAny(%q, val)\n", origField)
			code += "\t\t}\n"
		} else {
			// With rename
			code += fmt.Sprintf("\t\tif val, exists := streamv3.Get[any](r, %q); exists {\n", origField)
			code += fmt.Sprintf("\t\t\tresult = result.SetAny(%q, val)\n", newField)
			code += "\t\t}\n"
		}
	}
	code += "\t\treturn result.Freeze()\n"
	code += fmt.Sprintf("\t})(%s)", inputVar)

	// Create code fragment
	frag := lib.NewStmtFragment(outputVar, inputVar, code, nil, getCommandString())

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}
