package commands

import (
	"fmt"
	"os"
	"strings"
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
							var parsedValue any

							// Check if value is an expression or a literal
							if isExpression(upd.value) {
								// Evaluate as expression
								result, err := evaluateExpression(upd.value, frozen)
								if err != nil {
									// If expression evaluation fails, fall back to literal
									// This allows backward compatibility
									parsedValue = parseValue(upd.value)
								} else {
									parsedValue = result
								}
							} else {
								// Parse as literal
								parsedValue = parseValue(upd.value)
							}

							// Apply the value to the record
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
							case int:
								// expr might return int instead of int64
								mut = mut.Int(upd.field, int64(v))
							case float32:
								// expr might return float32 instead of float64
								mut = mut.Float(upd.field, float64(v))
							default:
								// For unknown types, convert to string
								mut = mut.String(upd.field, fmt.Sprintf("%v", v))
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

// generateUpdateCode generates Go code for the update command with conditional clauses
func generateUpdateCode(ctx *cf.Context, inputFile string) error {
	// Read all previous code fragments from stdin
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

	// Get input variable from last fragment (or default to "records")
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
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

	// Generate Update code with conditional clauses
	var codeBody strings.Builder
	needsTime := false
	needsStrings := false
	needsRegexp := false

	// Check if we need frozen (for reading in conditions)
	hasConditions := false
	for _, clause := range clauses {
		if len(clause.conditions) > 0 {
			hasConditions = true
			break
		}
	}

	if hasConditions {
		codeBody.WriteString("\t\tfrozen := mut.Freeze()\n\n")
	}

	// Generate clause evaluation (first-match-wins)
	for i, clause := range clauses {
		indent := "\t\t"

		// Generate condition check for this clause
		if len(clause.conditions) > 0 {
			if i == 0 {
				codeBody.WriteString(indent + "if ")
			} else {
				codeBody.WriteString(indent + "} else if ")
			}

			// Generate all conditions with AND logic
			for j, cond := range clause.conditions {
				if j > 0 {
					codeBody.WriteString(" && ")
				}
				codeBody.WriteString(generateConditionCode(cond.field, cond.op, cond.value))

				// Track which imports are needed
				switch cond.op {
				case "contains", "startswith", "endswith":
					needsStrings = true
				case "pattern", "regexp", "regex":
					needsRegexp = true
				}
			}
			codeBody.WriteString(" {\n")
			indent = "\t\t\t"
		} else if i > 0 {
			// Default case (no conditions) - use else
			codeBody.WriteString("\t\t} else {\n")
			indent = "\t\t\t"
		}

		// Generate update statements
		for _, upd := range clause.updates {
			parsedValue := parseValue(upd.value)

			var stmt string
			switch v := parsedValue.(type) {
			case int64:
				stmt = fmt.Sprintf("%smut = mut.Int(%q, int64(%d))", indent, upd.field, v)
			case float64:
				stmt = fmt.Sprintf("%smut = mut.Float(%q, %f)", indent, upd.field, v)
			case bool:
				stmt = fmt.Sprintf("%smut = mut.Bool(%q, %t)", indent, upd.field, v)
			case time.Time:
				timeExpr := fmt.Sprintf("time.Date(%d, %d, %d, %d, %d, %d, %d, time.UTC)",
					v.Year(), v.Month(), v.Day(), v.Hour(), v.Minute(), v.Second(), v.Nanosecond())
				stmt = fmt.Sprintf("%smut = ssql.Set(mut, %q, %s)", indent, upd.field, timeExpr)
				needsTime = true
			case string:
				stmt = fmt.Sprintf("%smut = mut.String(%q, %q)", indent, upd.field, v)
			default:
				stmt = fmt.Sprintf("%smut = mut.String(%q, %q)", indent, upd.field, upd.value)
			}

			codeBody.WriteString(stmt + "\n")
		}
	}

	// Close the if-else chain if we had conditions
	if len(clauses) > 0 && (len(clauses[0].conditions) > 0 || len(clauses) > 1) {
		codeBody.WriteString("\t\t}")
	}

	outputVar := "updated"
	code := fmt.Sprintf(`%s := ssql.Update(func(mut ssql.MutableRecord) ssql.MutableRecord {
%s
		return mut
	})(%s)`, outputVar, codeBody.String(), inputVar)

	// Determine imports needed
	imports := []string{}
	if needsTime {
		imports = append(imports, "time")
	}
	if needsStrings {
		imports = append(imports, "strings")
	}
	if needsRegexp {
		imports = append(imports, "regexp")
	}

	// Create and write fragment
	frag := lib.NewStmtFragment(outputVar, inputVar, code, imports, getCommandString())
	return lib.WriteCodeFragment(frag)
}

// generateConditionCode generates the Go code for a single condition check
func generateConditionCode(field, op, value string) string {
	switch op {
	case "eq":
		return fmt.Sprintf("ssql.GetOr(frozen, %q, %s) == %s",
			field, getDefaultValueForComparison(value), getComparisonValue(value))
	case "ne":
		return fmt.Sprintf("ssql.GetOr(frozen, %q, %s) != %s",
			field, getDefaultValueForComparison(value), getComparisonValue(value))
	case "gt", "ge", "lt", "le":
		// Numeric comparisons
		return fmt.Sprintf("ssql.GetOr(frozen, %q, float64(0)) %s %s",
			field, getOperatorCode(op), getComparisonValue(value))
	case "contains":
		return fmt.Sprintf("strings.Contains(ssql.GetOr(frozen, %q, \"\"), %s)",
			field, getComparisonValue(value))
	case "startswith":
		return fmt.Sprintf("strings.HasPrefix(ssql.GetOr(frozen, %q, \"\"), %s)",
			field, getComparisonValue(value))
	case "endswith":
		return fmt.Sprintf("strings.HasSuffix(ssql.GetOr(frozen, %q, \"\"), %s)",
			field, getComparisonValue(value))
	case "pattern", "regexp", "regex":
		// For regexp, we need to compile the pattern
		return fmt.Sprintf("regexp.MustCompile(%s).MatchString(ssql.GetOr(frozen, %q, \"\"))",
			getComparisonValue(value), field)
	default:
		// Fallback to equality
		return fmt.Sprintf("ssql.GetOr(frozen, %q, %s) == %s",
			field, getDefaultValueForComparison(value), getComparisonValue(value))
	}
}

// getDefaultValueForComparison returns the default value for GetOr based on the comparison value's type
func getDefaultValueForComparison(value string) string {
	parsedValue := parseValue(value)
	switch parsedValue.(type) {
	case int64, float64:
		return "float64(0)"
	case bool:
		return "false"
	case time.Time:
		return "time.Time{}"
	default:
		return `""`
	}
}

// getOperatorCode converts operator string to Go comparison operator
func getOperatorCode(op string) string {
	switch op {
	case "eq":
		return "=="
	case "ne":
		return "!="
	case "gt":
		return ">"
	case "ge":
		return ">="
	case "lt":
		return "<"
	case "le":
		return "<="
	default:
		return "=="
	}
}

// getComparisonValue formats a value for comparison in generated code
func getComparisonValue(value string) string {
	parsedValue := parseValue(value)
	switch v := parsedValue.(type) {
	case int64:
		return fmt.Sprintf("float64(%d)", v)
	case float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case string:
		return fmt.Sprintf("%q", v)
	default:
		return fmt.Sprintf("%q", value)
	}
}
