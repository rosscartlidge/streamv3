package commands

import (
	"fmt"
	"os"

	cf "github.com/rosscartlidge/autocli/v3"
	"github.com/rosscartlidge/ssql/v2"
	"github.com/rosscartlidge/ssql/v2/cmd/ssql/lib"
)

// RegisterWhere registers the where subcommand
func RegisterWhere(cmd *cf.CommandBuilder) *cf.CommandBuilder {
	cmd.Subcommand("where").
		Description("Filter records based on field conditions").
		Example("ssql read-csv data.csv | ssql where -match age gt 18", "Filter records where age > 18").
		Example("ssql read-csv sales.csv | ssql where -match status eq active -match amount gt 1000", "Active records with amount > 1000 (AND logic)").
		Example("ssql read-csv users.csv | ssql where -match dept eq Sales + -match dept eq Marketing", "Sales OR Marketing departments").
		Flag("-generate", "-g").
			Bool().
			Global().
			Help("Generate Go code instead of executing").
		Done().
		Flag("-match", "-m").
			Arg("field").Completer(cf.NoCompleter{Hint: "<field-name>"}).Done().
			Arg("operator").Completer(&cf.StaticCompleter{Options: []string{"eq", "ne", "gt", "ge", "lt", "le", "contains", "startswith", "endswith", "pattern", "regexp", "regex"}}).Done().
			Arg("value").Completer(cf.NoCompleter{Hint: "<value>"}).Done().
			Accumulate().
			Local().
			Help("Filter condition: -match <field> <operator> <value>").
		Done().
		Flag("FILE").
			String().
			Completer(&cf.FileCompleter{Pattern: "*.jsonl"}).
			Global().
			Default("").
			Help("Input JSONL file (or stdin if not specified)").
		Done().
		Handler(func(ctx *cf.Context) error {
			var inputFile string
			var generate bool

			if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
				inputFile = fileVal.(string)
			}

			if genVal, ok := ctx.GlobalFlags["-generate"]; ok {
				generate = genVal.(bool)
			}

			// Check if generation is enabled (flag or env var)
			if shouldGenerate(generate) {
				return generateWhereCode(ctx, inputFile)
			}

			// Build filter from clauses (OR between clauses, AND within)
			filter := func(r ssql.Record) bool {
				if len(ctx.Clauses) == 0 {
					return true
				}

				// OR logic between clauses
				for _, clause := range ctx.Clauses {
					// Get all -match conditions from this clause
					matchesRaw, ok := clause.Flags["-match"]
					if !ok || matchesRaw == nil {
						continue
					}

					matches, ok := matchesRaw.([]any)
					if !ok || len(matches) == 0 {
						continue
					}

					// AND logic within clause
					clauseMatches := true
					for _, matchRaw := range matches {
						matchMap, ok := matchRaw.(map[string]any)
						if !ok {
							clauseMatches = false
							break
						}

						field, _ := matchMap["field"].(string)
						op, _ := matchMap["operator"].(string)
						value, _ := matchMap["value"].(string)

						if field == "" || op == "" {
							clauseMatches = false
							break
						}

						// Get field value from record
						fieldValue, exists := ssql.Get[any](r, field)
						if !exists {
							clauseMatches = false
							break
						}

						// Apply operator
						if !applyOperator(fieldValue, op, value) {
							clauseMatches = false
							break
						}
					}

					if clauseMatches {
						return true // This clause matched
					}
				}

				return false // No clause matched
			}

			// Read JSONL from stdin or file
			input, err := lib.OpenInput(inputFile)
			if err != nil {
				return err
			}
			defer input.Close()

			records := lib.ReadJSONL(input)

			// Apply filter
			filtered := ssql.Where(filter)(records)

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, filtered); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Done()
	return cmd
}
