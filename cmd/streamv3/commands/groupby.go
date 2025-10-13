package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/rosscartlidge/gogstools/gs"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// GroupByConfig holds configuration for group-by command
type GroupByConfig struct {
	// Group-by field (will be in each clause)
	By string `gs:"field,global,last,help=Field to group by"`

	// Per-clause: aggregation specification
	Function string `gs:"string,local,last,help=Aggregation function,enum=count:sum:avg:min:max"`
	Field    string `gs:"field,local,last,help=Field to aggregate (not needed for count)"`
	Result   string `gs:"string,local,last,help=Output field name"`

	Generate bool   `gs:"flag,global,last,help=Generate Go code instead of executing"`
	Argv     string `gs:"file,global,last,help=Input JSONL file (or stdin if not specified),suffix=.jsonl"`
}

// groupByCommand implements the group-by command
type groupByCommand struct {
	config *GroupByConfig
	cmd    *gs.GSCommand
}

func init() {
	RegisterCommand(newGroupByCommand())
}

func newGroupByCommand() *groupByCommand {
	config := &GroupByConfig{}
	cmd, err := gs.NewCommand(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create group-by command: %v", err))
	}

	return &groupByCommand{
		config: config,
		cmd:    cmd,
	}
}

func (c *groupByCommand) Name() string {
	return "group-by"
}

func (c *groupByCommand) Description() string {
	return "Group records by fields and apply aggregations"
}

func (c *groupByCommand) GetGSCommand() *gs.GSCommand {
	return c.cmd
}

func (c *groupByCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before gs framework takes over
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("group-by - Group records by fields and apply aggregations")
		fmt.Println()
		fmt.Println("Usage: streamv3 group-by -by <field> -function <func> -result <name>")
		fmt.Println()
		fmt.Println("Group-by Fields:")
		fmt.Println("  -by <field>      Field to group by (can specify multiple)")
		fmt.Println()
		fmt.Println("Aggregation Functions:")
		fmt.Println("  count            Count records in each group")
		fmt.Println("  sum              Sum numeric field")
		fmt.Println("  avg              Average numeric field")
		fmt.Println("  min              Minimum value")
		fmt.Println("  max              Maximum value")
		fmt.Println()
		fmt.Println("Aggregation Specification (use + to separate multiple):")
		fmt.Println("  -function <func>  Aggregation function")
		fmt.Println("  -field <field>    Field to aggregate (not needed for count)")
		fmt.Println("  -result <name>    Output field name")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Count by department")
		fmt.Println("  streamv3 group-by -by department -function count -result count")
		fmt.Println()
		fmt.Println("  # Multiple aggregations")
		fmt.Println("  streamv3 group-by -by department \\")
		fmt.Println("    -function count -result count + \\")
		fmt.Println("    -function sum -field salary -result total + \\")
		fmt.Println("    -function avg -field salary -result avg_salary")
		fmt.Println()
		fmt.Println("  # Group by multiple fields")
		fmt.Println("  streamv3 group-by -by department -by state \\")
		fmt.Println("    -function count -result count")
		return nil
	}

	// Delegate to gs framework which will call Config.Execute
	return c.cmd.Execute(ctx, args)
}

// Validate implements gs.Commander interface
func (c *GroupByConfig) Validate() error {
	return nil
}

// Execute implements gs.Commander interface
// This is called by the gs framework after parsing arguments into clauses
func (c *GroupByConfig) Execute(ctx context.Context, clauses []gs.ClauseSet) error {
	// If -generate flag is set, generate Go code instead of executing
	if c.Generate {
		return c.generateCode(clauses)
	}

	// Normal execution: apply group-by and aggregations
	// Get group-by field from config (it's global)
	if c.By == "" {
		return fmt.Errorf("no group-by field specified (use -by)")
	}
	groupByFields := []string{c.By}

	// Parse aggregation specifications from clauses
	var aggSpecs []aggregationSpec
	for _, clause := range clauses {
		spec, err := parseAggregationSpec(clause)
		if err != nil {
			return err
		}
		if spec != nil {
			aggSpecs = append(aggSpecs, *spec)
		}
	}

	if len(aggSpecs) == 0 {
		return fmt.Errorf("no aggregations specified (use -function and -result)")
	}

	// Read JSONL from stdin or file
	input, err := lib.OpenInput(c.Argv)
	if err != nil {
		return err
	}
	defer input.Close()

	records := lib.ReadJSONL(input)

	// Apply GroupByFields (use "_group" as sequence field name)
	grouped := streamv3.GroupByFields("_group", groupByFields...)(records)

	// Build aggregations map
	aggregations := make(map[string]streamv3.AggregateFunc)
	for _, spec := range aggSpecs {
		agg, err := buildAggregator(spec)
		if err != nil {
			return err
		}
		aggregations[spec.result] = agg
	}

	// Apply Aggregate
	aggregated := streamv3.Aggregate("_group", aggregations)(grouped)

	// Write output as JSONL
	if err := lib.WriteJSONL(os.Stdout, aggregated); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}

// aggregationSpec holds parsed aggregation specification
type aggregationSpec struct {
	function string // count, sum, avg, min, max
	field    string // input field (empty for count)
	result   string // output field name
}

// parseAggregationSpec parses an aggregation spec from a clause
func parseAggregationSpec(clause gs.ClauseSet) (*aggregationSpec, error) {
	function, _ := clause.Fields["Function"].(string)
	field, _ := clause.Fields["Field"].(string)
	result, _ := clause.Fields["Result"].(string)

	// Skip empty clauses
	if function == "" && result == "" {
		return nil, nil
	}

	if function == "" {
		return nil, fmt.Errorf("aggregation missing -function")
	}

	if result == "" {
		return nil, fmt.Errorf("aggregation missing -result")
	}

	// Validate function
	switch function {
	case "count", "sum", "avg", "min", "max":
		// Valid
	default:
		return nil, fmt.Errorf("unknown aggregation function: %s", function)
	}

	// For non-count functions, field is required
	if function != "count" && field == "" {
		return nil, fmt.Errorf("aggregation function %s requires -field", function)
	}

	return &aggregationSpec{
		function: function,
		field:    field,
		result:   result,
	}, nil
}

// buildAggregator builds a StreamV3 AggregateFunc from a spec
func buildAggregator(spec aggregationSpec) (streamv3.AggregateFunc, error) {
	switch spec.function {
	case "count":
		return streamv3.Count(), nil
	case "sum":
		return streamv3.Sum(spec.field), nil
	case "avg":
		return streamv3.Avg(spec.field), nil
	case "min":
		return streamv3.Min[float64](spec.field), nil
	case "max":
		return streamv3.Max[float64](spec.field), nil
	default:
		return nil, fmt.Errorf("unknown aggregation function: %s", spec.function)
	}
}

// generateCode generates Go code for the group-by command
func (c *GroupByConfig) generateCode(clauses []gs.ClauseSet) error {
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

	// Get group-by field from config (it's global)
	if c.By == "" {
		return fmt.Errorf("no group-by field specified (use -by)")
	}
	groupByFields := []string{c.By}

	// Parse aggregation specifications from clauses
	var aggSpecs []aggregationSpec
	for _, clause := range clauses {
		spec, err := parseAggregationSpec(clause)
		if err != nil {
			return err
		}
		if spec != nil {
			aggSpecs = append(aggSpecs, *spec)
		}
	}

	if len(aggSpecs) == 0 {
		return fmt.Errorf("no aggregations specified (use -function and -result)")
	}

	// Generate complete group-by + aggregate code as single fragment
	code := generateGroupByAggregateCode(inputVar, groupByFields, aggSpecs)
	outputVar := "aggregated"
	frag := lib.NewStmtFragment(outputVar, inputVar, code, nil)
	return lib.WriteCodeFragment(frag)
}

// generateGroupByAggregateCode generates the complete group-by + aggregate pipeline
func generateGroupByAggregateCode(inputVar string, fields []string, specs []aggregationSpec) string {
	var code string

	// Generate GroupByFields call
	code = "grouped := streamv3.GroupByFields(\"_group\""
	for _, field := range fields {
		code += fmt.Sprintf(", %q", field)
	}
	code += fmt.Sprintf(")(%s)\n\t", inputVar)

	// Generate Aggregate call with map
	code += "aggregated := streamv3.Aggregate(\"_group\", map[string]streamv3.AggregateFunc{\n"
	for i, spec := range specs {
		if i > 0 {
			code += ",\n"
		}
		code += fmt.Sprintf("\t\t%q: %s", spec.result, generateAggregatorCode(spec))
	}
	code += ",\n\t})(grouped)"

	return code
}

// generateAggregatorCode generates code for a single aggregator
func generateAggregatorCode(spec aggregationSpec) string {
	switch spec.function {
	case "count":
		return "streamv3.Count()"
	case "sum":
		return fmt.Sprintf("streamv3.Sum(%q)", spec.field)
	case "avg":
		return fmt.Sprintf("streamv3.Avg(%q)", spec.field)
	case "min":
		return fmt.Sprintf("streamv3.Min[float64](%q)", spec.field)
	case "max":
		return fmt.Sprintf("streamv3.Max[float64](%q)", spec.field)
	default:
		return ""
	}
}
