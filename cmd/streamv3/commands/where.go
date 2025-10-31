package commands

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// whereCommand implements the where command
type whereCommand struct {
	cmd *cf.Command
}

func init() {
	RegisterCommand(newWhereCommand())
}

func newWhereCommand() *whereCommand {
	var inputFile string
	var generate bool

	operators := []string{"eq", "ne", "gt", "ge", "lt", "le", "contains", "startswith", "endswith", "pattern", "regexp", "regex"}

	cmd := cf.NewCommand("where").
		Description("Filter records based on field conditions").
		Flag("-match", "-m").
			Arg("field").Completer(cf.NoCompleter{Hint: "<field-name>"}).Done().
			Arg("operator").Completer(&cf.StaticCompleter{Options: operators}).Done().
			Arg("value").Completer(cf.NoCompleter{Hint: "<value>"}).Done().
			Accumulate().
			Local().
			Help("Filter condition: -match <field> <operator> <value>").
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
			if generate {
				return generateWhereCodeCF(ctx, inputFile)
			}

			// Normal execution: filter records
			// Build filter from clauses
			filter := buildWhereFilterCF(ctx.Clauses)

			// Read JSONL from stdin or file
			input, err := lib.OpenInput(inputFile)
			if err != nil {
				return err
			}
			defer input.Close()

			records := lib.ReadJSONL(input)

			// Apply filter
			filtered := streamv3.Where(filter)(records)

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, filtered); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Build()

	return &whereCommand{cmd: cmd}
}

func (c *whereCommand) Name() string {
	return "where"
}

func (c *whereCommand) Description() string {
	return "Filter records based on field conditions"
}

func (c *whereCommand) GetCFCommand() *cf.Command {
	return c.cmd
}


func (c *whereCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before completionflags framework takes over
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("where - Filter records based on field conditions")
		fmt.Println()
		fmt.Println("Usage: streamv3 where -match <field> <operator> <value>")
		fmt.Println()
		fmt.Println("Operators:")
		fmt.Println("  eq           Equal to")
		fmt.Println("  ne           Not equal to")
		fmt.Println("  gt           Greater than")
		fmt.Println("  ge           Greater than or equal")
		fmt.Println("  lt           Less than")
		fmt.Println("  le           Less than or equal")
		fmt.Println("  contains     String contains")
		fmt.Println("  startswith   String starts with")
		fmt.Println("  endswith     String ends with")
		fmt.Println("  pattern      Regexp pattern match (aliases: regexp, regex)")
		fmt.Println()
		fmt.Println("Clause Logic:")
		fmt.Println("  Multiple -match in same command: AND (all must match)")
		fmt.Println("  Separate clauses with +: OR (any clause can match)")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Single condition")
		fmt.Println("  streamv3 where -match age gt 18")
		fmt.Println()
		fmt.Println("  # Multiple conditions (AND): age > 18 AND status = active")
		fmt.Println("  streamv3 where -match age gt 18 -match status eq active")
		fmt.Println()
		fmt.Println("  # OR conditions: (age > 65) OR (age < 18)")
		fmt.Println("  streamv3 where -match age gt 65 + -match age lt 18")
		fmt.Println()
		fmt.Println("  # Complex: (age > 18 AND status = active) OR (department = Engineering)")
		fmt.Println("  streamv3 where -match age gt 18 -match status eq active + -match department eq Engineering")
		fmt.Println()
		fmt.Println("  # Regexp matching: department ends with 'ing'")
		fmt.Println("  streamv3 where -match department regexp \".*ing$\"")
		fmt.Println()
		fmt.Println("  # Regexp matching: email pattern")
		fmt.Println("  streamv3 where -match email regex \"^[a-z]+@[a-z]+\\.com$\"")
		fmt.Println()
		fmt.Println("Debugging with jq:")
		fmt.Println("  # Inspect records before/after filtering")
		fmt.Println("  streamv3 read-csv data.csv | jq '.' | head -5")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 where -match age gt 30 | jq '.'")
		fmt.Println()
		fmt.Println("  # Check field types")
		fmt.Println("  streamv3 read-csv data.csv | jq '.age | type' | head -5")
		fmt.Println()
		fmt.Println("  # Count matching records")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 where -match status eq active | jq -s 'length'")
		return nil
	}

	return c.cmd.Execute(args)
}

// matchCondition represents a single match condition
type matchCondition struct {
	field string
	op    string
	value string
}

// buildWhereFilterCF builds a filter function from completionflags clauses
func buildWhereFilterCF(clauses []cf.Clause) func(streamv3.Record) bool {
	return func(r streamv3.Record) bool {
		if len(clauses) == 0 {
			return true // No filter conditions
		}

		// OR logic between clauses (any clause can match)
		for _, clause := range clauses {
			// AND logic within clause (all conditions must match)
			clauseMatches := evaluateClauseCF(r, clause)
			if clauseMatches {
				return true // This clause matched, record passes
			}
		}

		return false // No clause matched
	}
}

// evaluateClauseCF evaluates all conditions within a single clause (AND logic)
func evaluateClauseCF(r streamv3.Record, clause cf.Clause) bool {
	// Get the -match flags from the clause (will be []any where each element is []string{field, op, value})
	matchesRaw, ok := clause.Flags["-match"]
	if !ok || matchesRaw == nil {
		return true // No match conditions means everything passes
	}

	// Convert to slice
	matches, ok := matchesRaw.([]any)
	if !ok || len(matches) == 0 {
		return true // No match conditions
	}

	// AND logic: all matches must pass
	for _, matchRaw := range matches {
		matchMap, ok := matchRaw.(map[string]any)
		if !ok {
			return false // Invalid match specification
		}

		field, _ := matchMap["field"].(string)
		op, _ := matchMap["operator"].(string)
		value, _ := matchMap["value"].(string)

		if field == "" || op == "" {
			return false // Invalid match
		}

		// Get field value from record
		fieldValue, exists := r[field]
		if !exists {
			return false // Field doesn't exist
		}

		// Apply operator - if any condition fails, entire clause fails (AND logic)
		if !applyOperator(fieldValue, op, value) {
			return false
		}
	}

	return true // All conditions passed
}

// applyOperator applies a comparison operator
func applyOperator(fieldValue any, op string, compareValue string) bool {
	switch op {
	case "eq":
		return compareEqual(fieldValue, compareValue)
	case "ne":
		return !compareEqual(fieldValue, compareValue)
	case "gt":
		return compareGreater(fieldValue, compareValue)
	case "ge":
		return compareGreater(fieldValue, compareValue) || compareEqual(fieldValue, compareValue)
	case "lt":
		return compareLess(fieldValue, compareValue)
	case "le":
		return compareLess(fieldValue, compareValue) || compareEqual(fieldValue, compareValue)
	case "contains":
		return compareContains(fieldValue, compareValue)
	case "startswith":
		return compareStartsWith(fieldValue, compareValue)
	case "endswith":
		return compareEndsWith(fieldValue, compareValue)
	case "pattern", "regexp", "regex":
		return comparePattern(fieldValue, compareValue)
	default:
		return false
	}
}

// compareEqual checks equality
func compareEqual(fieldValue any, compareValue string) bool {
	switch v := fieldValue.(type) {
	case string:
		return v == compareValue
	case int64:
		if num, err := strconv.ParseInt(compareValue, 10, 64); err == nil {
			return v == num
		}
	case float64:
		if num, err := strconv.ParseFloat(compareValue, 64); err == nil {
			return v == num
		}
	case bool:
		if b, err := strconv.ParseBool(compareValue); err == nil {
			return v == b
		}
	}
	return fmt.Sprintf("%v", fieldValue) == compareValue
}

// compareGreater checks if field > value
func compareGreater(fieldValue any, compareValue string) bool {
	switch v := fieldValue.(type) {
	case int64:
		if num, err := strconv.ParseInt(compareValue, 10, 64); err == nil {
			return v > num
		}
	case float64:
		if num, err := strconv.ParseFloat(compareValue, 64); err == nil {
			return v > num
		}
	case string:
		return v > compareValue
	}
	return false
}

// compareLess checks if field < value
func compareLess(fieldValue any, compareValue string) bool {
	switch v := fieldValue.(type) {
	case int64:
		if num, err := strconv.ParseInt(compareValue, 10, 64); err == nil {
			return v < num
		}
	case float64:
		if num, err := strconv.ParseFloat(compareValue, 64); err == nil {
			return v < num
		}
	case string:
		return v < compareValue
	}
	return false
}

// compareContains checks if string contains substring
func compareContains(fieldValue any, compareValue string) bool {
	if str, ok := fieldValue.(string); ok {
		return contains(str, compareValue)
	}
	return false
}

// compareStartsWith checks if string starts with prefix
func compareStartsWith(fieldValue any, compareValue string) bool {
	if str, ok := fieldValue.(string); ok {
		return len(str) >= len(compareValue) && str[:len(compareValue)] == compareValue
	}
	return false
}

// compareEndsWith checks if string ends with suffix
func compareEndsWith(fieldValue any, compareValue string) bool {
	if str, ok := fieldValue.(string); ok {
		return len(str) >= len(compareValue) && str[len(str)-len(compareValue):] == compareValue
	}
	return false
}

// comparePattern checks if string matches regexp pattern
func comparePattern(fieldValue any, pattern string) bool {
	if str, ok := fieldValue.(string); ok {
		matched, err := regexp.MatchString(pattern, str)
		if err != nil {
			return false // Invalid pattern
		}
		return matched
	}
	return false
}

// contains checks if string contains substring (simple implementation)
func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// generateWhereCodeCF generates Go code for the where command
func generateWhereCodeCF(ctx *cf.Context, inputFile string) error {
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
	code := fmt.Sprintf("%s := streamv3.Where(%s)(%s)", outputVar, filterCode, inputVar)

	// Create code fragment
	frag := lib.NewStmtFragment(outputVar, inputVar, code, imports)

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
				clauseConditions = append(clauseConditions, "("+joinWithAnd(andConditions)+")")
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
		finalCondition = "return " + joinWithOr(clauseConditions)
	}

	// Build function
	code := fmt.Sprintf("func(r streamv3.Record) bool {\n\t\t%s\n\t}", finalCondition)

	return code, dedupeImports(imports)
}

// generateCondition generates code for a single condition
func generateCondition(field, op, value string) (string, []string) {
	var imports []string

	// Detect if value is numeric
	isNum := isNumeric(value)

	switch op {
	case "eq":
		if isNum {
			return fmt.Sprintf("streamv3.GetOr(r, %q, float64(0)) == %s", field, value), nil
		}
		return fmt.Sprintf("r[%q] == %q", field, value), nil

	case "ne":
		if isNum {
			return fmt.Sprintf("streamv3.GetOr(r, %q, float64(0)) != %s", field, value), nil
		}
		return fmt.Sprintf("r[%q] != %q", field, value), nil

	case "gt":
		return fmt.Sprintf("streamv3.GetOr(r, %q, float64(0)) > %s", field, value), nil

	case "ge":
		return fmt.Sprintf("streamv3.GetOr(r, %q, float64(0)) >= %s", field, value), nil

	case "lt":
		return fmt.Sprintf("streamv3.GetOr(r, %q, float64(0)) < %s", field, value), nil

	case "le":
		return fmt.Sprintf("streamv3.GetOr(r, %q, float64(0)) <= %s", field, value), nil

	case "contains":
		imports = append(imports, "strings")
		return fmt.Sprintf("strings.Contains(streamv3.GetOr(r, %q, \"\"), %q)", field, value), imports

	case "startswith":
		imports = append(imports, "strings")
		return fmt.Sprintf("strings.HasPrefix(streamv3.GetOr(r, %q, \"\"), %q)", field, value), imports

	case "endswith":
		imports = append(imports, "strings")
		return fmt.Sprintf("strings.HasSuffix(streamv3.GetOr(r, %q, \"\"), %q)", field, value), imports

	case "pattern", "regexp", "regex":
		imports = append(imports, "regexp")
		return fmt.Sprintf("regexp.MustCompile(%q).MatchString(streamv3.GetOr(r, %q, \"\"))", value, field), imports

	default:
		return "false", nil
	}
}

// joinWithAnd joins conditions with &&
func joinWithAnd(conditions []string) string {
	return strings.Join(conditions, " && ")
}

// joinWithOr joins conditions with ||
func joinWithOr(conditions []string) string {
	return strings.Join(conditions, " || ")
}

// isNumeric checks if a string is a number
func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
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
