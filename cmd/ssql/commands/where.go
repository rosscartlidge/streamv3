package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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

// generateWhereCode generates Go code for the where command
func generateWhereCode(ctx *cf.Context, inputFile string) error {
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

	// Get input variable name from last fragment
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}

	// Generate filter code from clauses
	filterCode, imports := generateWhereCodeFromClauses(ctx.Clauses)

	// Build complete statement
	outputVar := "filtered"
	code := fmt.Sprintf("%s := ssql.Where(%s)(%s)", outputVar, filterCode, inputVar)

	// Create code fragment
	frag := lib.NewStmtFragment(outputVar, inputVar, code, imports, getCommandString())

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}

// generateWhereCodeFromClauses generates the filter function code
func generateWhereCodeFromClauses(clauses []cf.Clause) (string, []string) {
	var imports []string
	var clauseConditions []string

	// Build conditions for each clause (OR logic between clauses)
	for _, clause := range clauses {
		matchesRaw, ok := clause.Flags["-match"]
		if !ok || matchesRaw == nil {
			continue
		}

		matches, ok := matchesRaw.([]any)
		if !ok || len(matches) == 0 {
			continue
		}

		// AND logic within clause: all matches must pass
		var andConditions []string
		for _, matchRaw := range matches {
			matchMap, ok := matchRaw.(map[string]any)
			if !ok {
				continue
			}

			field, _ := matchMap["field"].(string)
			op, _ := matchMap["operator"].(string)
			value, _ := matchMap["value"].(string)

			if field == "" || op == "" {
				continue
			}

			// Generate condition code
			cond, imp := generateCondition(field, op, value)
			andConditions = append(andConditions, cond)
			imports = append(imports, imp...)
		}

		// Combine AND conditions for this clause
		if len(andConditions) > 0 {
			if len(andConditions) == 1 {
				clauseConditions = append(clauseConditions, andConditions[0])
			} else {
				clauseConditions = append(clauseConditions, "("+strings.Join(andConditions, " && ")+")")
			}
		}
	}

	// Combine clauses with OR
	var finalCondition string
	if len(clauseConditions) == 0 {
		finalCondition = "return true"
	} else if len(clauseConditions) == 1 {
		finalCondition = "return " + clauseConditions[0]
	} else {
		finalCondition = "return " + strings.Join(clauseConditions, " || ")
	}

	// Build function
	code := fmt.Sprintf("func(r ssql.Record) bool {\n\t\t%s\n\t}", finalCondition)

	return code, dedupeImports(imports)
}

// generateCondition generates code for a single where condition
func generateCondition(field, op, value string) (string, []string) {
	var imports []string

	// Detect if value is numeric
	_, err := strconv.ParseFloat(value, 64)
	isNum := err == nil

	switch op {
	case "eq":
		if isNum {
			return fmt.Sprintf("ssql.GetOr(r, %q, float64(0)) == %s", field, value), nil
		}
		return fmt.Sprintf("ssql.GetOr(r, %q, \"\") == %q", field, value), nil

	case "ne":
		if isNum {
			return fmt.Sprintf("ssql.GetOr(r, %q, float64(0)) != %s", field, value), nil
		}
		return fmt.Sprintf("ssql.GetOr(r, %q, \"\") != %q", field, value), nil

	case "gt":
		return fmt.Sprintf("ssql.GetOr(r, %q, float64(0)) > %s", field, value), nil

	case "ge":
		return fmt.Sprintf("ssql.GetOr(r, %q, float64(0)) >= %s", field, value), nil

	case "lt":
		return fmt.Sprintf("ssql.GetOr(r, %q, float64(0)) < %s", field, value), nil

	case "le":
		return fmt.Sprintf("ssql.GetOr(r, %q, float64(0)) <= %s", field, value), nil

	case "contains":
		imports = append(imports, "strings")
		return fmt.Sprintf("strings.Contains(ssql.GetOr(r, %q, \"\"), %q)", field, value), imports

	case "startswith":
		imports = append(imports, "strings")
		return fmt.Sprintf("strings.HasPrefix(ssql.GetOr(r, %q, \"\"), %q)", field, value), imports

	case "endswith":
		imports = append(imports, "strings")
		return fmt.Sprintf("strings.HasSuffix(ssql.GetOr(r, %q, \"\"), %q)", field, value), imports

	case "pattern", "regexp", "regex":
		imports = append(imports, "regexp")
		return fmt.Sprintf("regexp.MustCompile(%q).MatchString(ssql.GetOr(r, %q, \"\"))", value, field), imports

	default:
		return "false", nil
	}
}

// dedupeImports removes duplicate imports
func dedupeImports(imports []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, imp := range imports {
		if imp != "" && !seen[imp] {
			seen[imp] = true
			result = append(result, imp)
		}
	}
	return result
}
