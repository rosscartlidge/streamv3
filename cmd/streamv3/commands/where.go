package commands

import (
	"os"
	"context"
	"fmt"
	"strconv"

	"github.com/rosscartlidge/gogstools/gs"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// WhereConfig holds configuration for where command
type WhereConfig struct {
	// Match uses multi-argument pattern: -match field op value
	// Multiple -match in same clause are ANDed, separate clauses (+) are ORed
	Match []map[string]interface{} `gs:"multi,local,list,args=field:op:value,help=Filter condition: field operator value"`
	Argv  string                   `gs:"file,global,last,help=Input JSONL file (or stdin if not specified),suffix=.jsonl"`
}

// whereCommand implements the where command
type whereCommand struct {
	config *WhereConfig
	cmd    *gs.GSCommand
}

func init() {
	RegisterCommand(newWhereCommand())
}

func newWhereCommand() *whereCommand {
	config := &WhereConfig{}
	cmd, err := gs.NewCommand(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create where command: %v", err))
	}

	return &whereCommand{
		config: config,
		cmd:    cmd,
	}
}

func (c *whereCommand) Name() string {
	return "where"
}

func (c *whereCommand) Description() string {
	return "Filter records based on field conditions"
}

func (c *whereCommand) GetGSCommand() *gs.GSCommand {
	return c.cmd
}

func (c *whereCommand) Execute(ctx context.Context, args []string) error {
	// This method is called by our main router for custom help handling
	// Handle -help flag before gs framework takes over
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
		return nil
	}

	// For actual execution, delegate to gs framework which will call Config.Execute
	return c.cmd.Execute(ctx, args)
}

// buildWhereFilter builds a filter function from parsed clauses
// Each clause is evaluated independently, results are ORed together
// Multiple conditions within a clause are ANDed together
func buildWhereFilter(clauses []gs.ClauseSet) func(streamv3.Record) bool {
	return func(r streamv3.Record) bool {
		if len(clauses) == 0 {
			return true // No filter conditions
		}

		// OR logic between clauses (any clause can match)
		for _, clause := range clauses {
			// AND logic within clause (all conditions must match)
			clauseMatches := evaluateClause(r, clause)
			if clauseMatches {
				return true // This clause matched, record passes
			}
		}

		return false // No clause matched
	}
}

// evaluateClause evaluates all conditions within a single clause (AND logic)
func evaluateClause(r streamv3.Record, clause gs.ClauseSet) bool {
	// Get the Match list from the clause
	matchListInterface, ok := clause.Fields["Match"]
	if !ok {
		return true // No match conditions means everything passes
	}

	// Convert []interface{} to []map[string]interface{}
	matchListSlice, ok := matchListInterface.([]interface{})
	if !ok || len(matchListSlice) == 0 {
		return true // Empty clause matches everything
	}

	// All match conditions must pass (AND logic within clause)
	for _, matchInterface := range matchListSlice {
		match, ok := matchInterface.(map[string]interface{})
		if !ok {
			continue
		}

		field, _ := match["field"].(string)
		op, _ := match["op"].(string)
		value, _ := match["value"].(string)

		if field == "" || op == "" {
			continue // Skip incomplete match
		}

		// Get field value from record
		fieldValue, exists := r[field]
		if !exists {
			return false // Field doesn't exist
		}

		// Apply operator - if any condition fails, clause fails
		if !applyOperator(fieldValue, op, value) {
			return false
		}
	}

	return true // All conditions in clause matched
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

// contains checks if string contains substring (simple implementation)
func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Validate implements gs.Commander interface
func (c *WhereConfig) Validate() error {
	return nil
}

// Execute implements gs.Commander interface
// This is called by the gs framework after parsing arguments into clauses
func (c *WhereConfig) Execute(ctx context.Context, clauses []gs.ClauseSet) error {
	// Read JSONL from stdin or file
	input, err := lib.OpenInput(c.Argv)
	if err != nil {
		return err
	}
	defer input.Close()

	records := lib.ReadJSONL(input)

	// Build filter from clauses
	filter := buildWhereFilter(clauses)

	// Apply filter
	filtered := streamv3.Where(filter)(records)

	// Write output as JSONL
	if err := lib.WriteJSONL(os.Stdout, filtered); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}
