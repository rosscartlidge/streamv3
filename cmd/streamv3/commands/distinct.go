package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// distinctCommand implements the distinct command
type distinctCommand struct {
	cmd *cf.Command
}

func init() {
	RegisterCommand(newDistinctCommand())
}

func newDistinctCommand() *distinctCommand {
	var inputFile string
	var generate bool

	cmd := cf.NewCommand("distinct").
		Description("Remove duplicate records (SQL DISTINCT)").
		Flag("-by", "-b").
			String().
			Completer(cf.NoCompleter{Hint: "<field-name>"}).
			Accumulate().
			Local().
			Help("Field to use for uniqueness (can specify multiple)").
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
			if shouldGenerate(generate) {
				return generateDistinctCode(ctx, inputFile)
			}

			// Get -by fields from first clause
			var byFields []string
			if len(ctx.Clauses) > 0 {
				if byRaw, ok := ctx.Clauses[0].Flags["-by"]; ok {
					if bySlice, ok := byRaw.([]any); ok {
						for _, v := range bySlice {
							if field, ok := v.(string); ok && field != "" {
								byFields = append(byFields, field)
							}
						}
					}
				}
			}

			// Read JSONL from stdin or file
			input, err := lib.OpenInput(inputFile)
			if err != nil {
				return err
			}
			defer input.Close()

			records := lib.ReadJSONL(input)

			// Apply distinct
			var result streamv3.Filter[streamv3.Record, streamv3.Record]
			if len(byFields) == 0 {
				// Distinct on all fields - requires Record to be comparable
				// Since Record is map[string]any, we need to use DistinctBy with a string key
				result = streamv3.DistinctBy(func(r streamv3.Record) string {
					return recordToKey(r)
				})
			} else if len(byFields) == 1 {
				// Distinct by single field
				field := byFields[0]
				result = streamv3.DistinctBy(func(r streamv3.Record) string {
					return fmt.Sprintf("%v", streamv3.GetOr(r, field, ""))
				})
			} else {
				// Distinct by multiple fields (composite key)
				result = streamv3.DistinctBy(func(r streamv3.Record) string {
					var parts []string
					for _, field := range byFields {
						parts = append(parts, fmt.Sprintf("%v", streamv3.GetOr(r, field, "")))
					}
					return strings.Join(parts, "|")
				})
			}

			distinct := result(records)

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, distinct); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Build()

	return &distinctCommand{cmd: cmd}
}

func (c *distinctCommand) Name() string {
	return "distinct"
}

func (c *distinctCommand) Description() string {
	return "Remove duplicate records (SQL DISTINCT)"
}

func (c *distinctCommand) GetCFCommand() *cf.Command {
	return c.cmd
}

func (c *distinctCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before completionflags framework takes over
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("distinct - Remove duplicate records (SQL DISTINCT)")
		fmt.Println()
		fmt.Println("Usage: streamv3 distinct [-by <field>]...")
		fmt.Println()
		fmt.Println("Removes duplicate records from the stream. By default, compares")
		fmt.Println("entire records. Use -by to specify fields for uniqueness.")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Remove duplicate records (compare all fields)")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 distinct")
		fmt.Println()
		fmt.Println("  # Distinct by single field")
		fmt.Println("  streamv3 read-csv employees.csv | streamv3 distinct -by department")
		fmt.Println()
		fmt.Println("  # Distinct by multiple fields (composite key)")
		fmt.Println("  streamv3 read-csv data.csv | \\")
		fmt.Println("    streamv3 distinct -by department -by location")
		fmt.Println()
		fmt.Println("  # SQL equivalent: SELECT DISTINCT department, location FROM data")
		fmt.Println("  streamv3 read-csv data.csv | \\")
		fmt.Println("    streamv3 select -field department + -field location | \\")
		fmt.Println("    streamv3 distinct")
		return nil
	}

	return c.cmd.Execute(args)
}

// recordToKey converts a record to a string key for deduplication
func recordToKey(r streamv3.Record) string {
	// Create a stable string representation of the record
	// Sort keys to ensure consistency and exclude _row_number
	var keys []string
	for k := range r.KeysIter() {
		if k != "_row_number" {
			keys = append(keys, k)
		}
	}

	// Sort keys for deterministic output
	sortStrings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, streamv3.GetOr(r, k, "")))
	}
	return strings.Join(parts, "|")
}

// sortStrings sorts a slice of strings in place (simple bubble sort for small slices)
func sortStrings(s []string) {
	n := len(s)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if s[j] > s[j+1] {
				s[j], s[j+1] = s[j+1], s[j]
			}
		}
	}
}

// generateDistinctCode generates Go code for the distinct command
func generateDistinctCode(ctx *cf.Context, inputFile string) error {
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

	// Get -by fields
	var byFields []string
	if len(ctx.Clauses) > 0 {
		if byRaw, ok := ctx.Clauses[0].Flags["-by"]; ok {
			if bySlice, ok := byRaw.([]any); ok {
				for _, v := range bySlice {
					if field, ok := v.(string); ok && field != "" {
						byFields = append(byFields, field)
					}
				}
			}
		}
	}

	// Generate distinct code
	var code string
	outputVar := "distinct"

	if len(byFields) == 0 {
		// Distinct on all fields
		code = fmt.Sprintf(`%s := streamv3.DistinctBy(func(r streamv3.Record) string {
		var parts []string
		for k, v := range r.All() {
			parts = append(parts, fmt.Sprintf("%%s=%%v", k, v))
		}
		return strings.Join(parts, "|")
	})(%s)`, outputVar, inputVar)
	} else if len(byFields) == 1 {
		// Single field
		field := byFields[0]
		code = fmt.Sprintf(`%s := streamv3.DistinctBy(func(r streamv3.Record) string {
		return fmt.Sprintf("%%v", streamv3.GetOr(r, %q, ""))
	})(%s)`, outputVar, field, inputVar)
	} else {
		// Multiple fields
		var fieldRefs []string
		for _, field := range byFields {
			fieldRefs = append(fieldRefs, fmt.Sprintf(`fmt.Sprintf("%%v", streamv3.GetOr(r, %q, ""))`, field))
		}
		code = fmt.Sprintf(`%s := streamv3.DistinctBy(func(r streamv3.Record) string {
		parts := []string{%s}
		return strings.Join(parts, "|")
	})(%s)`, outputVar, strings.Join(fieldRefs, ", "), inputVar)
	}

	// Create code fragment with imports
	imports := []string{"fmt", "strings"}
	frag := lib.NewStmtFragment(outputVar, inputVar, code, imports, getCommandString())

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}
