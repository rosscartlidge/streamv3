package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rosscartlidge/gogstools/gs"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// WhereConfig holds configuration for where command
type WhereConfig struct {
	Field string `gs:"field,local,last,help=Field to filter on"`
	Op    string `gs:"string,local,last,help=Comparison operator,enum=eq:ne:gt:ge:lt:le:contains:startswith:endswith"`
	Value string `gs:"string,local,last,help=Value to compare against"`
	Argv  string `gs:"file,global,last,help=Input JSONL file (or stdin if not specified),suffix=.jsonl"`
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

func (c *whereCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("where - Filter records based on field conditions")
		fmt.Println()
		fmt.Println("Usage: streamv3 where - field <name> - op <operator> - value <val>")
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
		fmt.Println("Examples:")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 where - field age - op gt - value 18")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 where - field name - op contains - value Smith")
		fmt.Println()
		fmt.Println("Multiple conditions (AND):")
		fmt.Println("  streamv3 where - field age - op gt - value 18 - field status - op eq - value active")
		return nil
	}

	// Parse arguments using gs framework
	clauses, err := c.cmd.Parse(args)
	if err != nil {
		return fmt.Errorf("parsing arguments: %w", err)
	}

	// Read JSONL from stdin
	input, err := lib.OpenInput(c.config.Argv)
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
	if err := lib.WriteJSONL(lib.Stdout, filtered); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}

// buildWhereFilter builds a filter function from parsed clauses
func buildWhereFilter(clauses []gs.ClauseSet) func(streamv3.Record) bool {
	return func(r streamv3.Record) bool {
		// All clauses must match (AND logic)
		for _, clause := range clauses {
			field, _ := clause.Fields["Field"].(string)
			op, _ := clause.Fields["Op"].(string)
			value, _ := clause.Fields["Value"].(string)

			if field == "" || op == "" {
				continue // Skip incomplete clauses
			}

			// Get field value from record
			fieldValue, exists := r[field]
			if !exists {
				return false // Field doesn't exist
			}

			// Apply operator
			if !applyOperator(fieldValue, op, value) {
				return false
			}
		}
		return true
	}
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

// Execute implements gs.Commander interface (not used, we use command.Execute)
func (c *WhereConfig) Execute(ctx context.Context, clauses []gs.ClauseSet) error {
	return nil
}
