package commands

import (
	"fmt"
	"os"
	"time"

	cf "github.com/rosscartlidge/autocli/v3"
	"github.com/rosscartlidge/ssql/v2"
	"github.com/rosscartlidge/ssql/v2/cmd/ssql/lib"
)

// RegisterUpdate registers the update subcommand
func RegisterUpdate(cmd *cf.CommandBuilder) *cf.CommandBuilder {
	cmd.Subcommand("update").
		Description("Conditionally update record fields with new values").
		Example("ssql read-csv users.csv | ssql update -match status eq pending -set status approved", "Update status from pending to approved").
		Example("ssql read-csv sales.csv | ssql update -match region eq US -set tax_rate 0.08 -set currency USD", "Set multiple fields for US region").
		Example("ssql read-csv data.csv | ssql update -match age lt 18 -set category minor + -match age ge 18 -set category adult", "Categorize by age using if-else logic").
		ClauseDescription("Clauses are evaluated in order using if-then-else logic.\nSeparators: +, -\nThe FIRST matching clause applies its updates, then processing stops (first-match-wins).\nThis is different from 'where' which uses OR logic - all clauses are evaluated.").
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
			Help("Condition to check: -match <field> <operator> <value>").
		Done().
		Flag("-set", "-s").
			Arg("field").Completer(cf.NoCompleter{Hint: "<field-name>"}).Done().
			Arg("value").Completer(cf.NoCompleter{Hint: "<value>"}).Done().
			Accumulate().
			Local().
			Help("Set field to value: -set <field> <value>").
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
				return generateUpdateCode(ctx, inputFile)
			}

			// Parse clauses - each clause has optional -match conditions and required -set operations
			type updateClause struct {
				conditions []struct {
					field string
					op    string
					value string
				}
				updates []struct {
					field string
					value string
				}
			}

			var clauses []updateClause

			for _, clause := range ctx.Clauses {
				uc := updateClause{}

				// Parse -match conditions (optional)
				if matchesRaw, ok := clause.Flags["-match"]; ok && matchesRaw != nil {
					matches, ok := matchesRaw.([]any)
					if ok {
						for _, matchRaw := range matches {
							matchMap, ok := matchRaw.(map[string]any)
							if !ok {
								continue
							}

							field, _ := matchMap["field"].(string)
							op, _ := matchMap["operator"].(string)
							value, _ := matchMap["value"].(string)

							if field != "" && op != "" {
								uc.conditions = append(uc.conditions, struct {
									field string
									op    string
									value string
								}{field, op, value})
							}
						}
					}
				}

				// Parse -set operations (required)
				if setOpsRaw, ok := clause.Flags["-set"]; ok && setOpsRaw != nil {
					setList, ok := setOpsRaw.([]any)
					if ok {
						for _, setRaw := range setList {
							setMap, ok := setRaw.(map[string]any)
							if !ok {
								continue
							}

							field, _ := setMap["field"].(string)
							value, _ := setMap["value"].(string)

							if field != "" {
								uc.updates = append(uc.updates, struct {
									field string
									value string
								}{field, value})
							}
						}
					}
				}

				if len(uc.updates) > 0 {
					clauses = append(clauses, uc)
				}
			}

			if len(clauses) == 0 {
				return fmt.Errorf("no -set operations specified")
			}

			// Read JSONL from stdin or file
			input, err := lib.OpenInput(inputFile)
			if err != nil {
				return err
			}
			defer input.Close()

			records := lib.ReadJSONL(input)

			// Build update filter with first-match-wins clause evaluation
			updateFilter := ssql.Update(func(mut ssql.MutableRecord) ssql.MutableRecord {
				frozen := mut.Freeze()

				// Evaluate clauses in order - first match wins
				for _, clause := range clauses {
					// Check all conditions in this clause (AND logic)
					allMatch := true
					for _, cond := range clause.conditions {
						fieldValue, exists := ssql.Get[any](frozen, cond.field)
						if !exists || !applyOperator(fieldValue, cond.op, cond.value) {
							allMatch = false
							break
						}
					}

					// If clause matches (or has no conditions), apply updates and stop
					if allMatch {
						for _, upd := range clause.updates {
							parsedValue := parseValue(upd.value)
							switch v := parsedValue.(type) {
							case int64:
								mut = mut.Int(upd.field, v)
							case float64:
								mut = mut.Float(upd.field, v)
							case bool:
								mut = mut.Bool(upd.field, v)
							case time.Time:
								mut = ssql.Set(mut, upd.field, v)
							case string:
								mut = mut.String(upd.field, v)
							}
						}
						break // First match wins
					}
				}

				return mut
			})

			// Apply update
			updated := updateFilter(records)

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, updated); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Done()
	return cmd
}
