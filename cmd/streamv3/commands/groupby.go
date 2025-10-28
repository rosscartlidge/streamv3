package commands

import (
	"context"
	"fmt"
	"os"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/gogstools/gs"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// groupByCommand implements the group-by command
type groupByCommand struct {
	cmd *cf.Command
}

func init() {
	RegisterCommand(newGroupByCommand())
}

func newGroupByCommand() *groupByCommand {
	var byField, inputFile string
	var generate bool

	cmd := cf.NewCommand("group-by").
		Description("Group records by fields and apply aggregations").
		Flag("-by", "-b").
			String().
			Bind(&byField).
			Global().
			Help("Field to group by").
			Done().
		Flag("-function", "-func").
			String().
			Local().
			Help("Aggregation function (count, sum, avg, min, max)").
			Done().
		Flag("-field", "-f").
			String().
			Local().
			Help("Field to aggregate (not needed for count)").
			Done().
		Flag("-result", "-r").
			String().
			Local().
			Help("Output field name").
			Done().
		Flag("-generate", "-g").
			Bool().
			Bind(&generate).
			Global().
			Help("Generate Go code instead of executing").
			Done().
		Flag("FILE").
			String().
			Bind(&inputFile).
			Global().
			Default("").
			FilePattern("*.jsonl").
			Help("Input JSONL file (or stdin if not specified)").
			Done().
		Handler(func(ctx *cf.Context) error {
			// If -generate flag is set, generate Go code instead of executing
			if generate {
				return generateGroupByCode(ctx, byField, inputFile)
			}

			// Normal execution: apply group-by and aggregations
			if byField == "" {
				return fmt.Errorf("no group-by field specified (use -by)")
			}
			groupByFields := []string{byField}

			// Parse aggregation specifications from clauses
			var aggSpecs []aggregationSpec
			for _, clause := range ctx.Clauses {
				spec, err := parseAggregationSpecCF(clause)
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
			input, err := lib.OpenInput(inputFile)
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
		}).
		Build()

	return &groupByCommand{cmd: cmd}
}

func (c *groupByCommand) Name() string {
	return "group-by"
}

func (c *groupByCommand) Description() string {
	return "Group records by fields and apply aggregations"
}

func (c *groupByCommand) GetCFCommand() *cf.Command {
	return c.cmd
}

func (c *groupByCommand) GetGSCommand() *gs.GSCommand {
	return nil // No longer using gs
}

func (c *groupByCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before completionflags framework takes over
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
		fmt.Println()
		fmt.Println("Debugging with jq:")
		fmt.Println("  # Inspect GROUP BY results")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 group-by -by dept -function count -result n | jq '.'")
		fmt.Println()
		fmt.Println("  # Verify grouping keys")
		fmt.Println("  streamv3 read-csv data.csv | jq -r '.department' | sort | uniq -c")
		fmt.Println()
		fmt.Println("  # Check specific group")
		fmt.Println("  streamv3 ... | streamv3 group-by ... | jq 'select(.department == \"Engineering\")'")
		return nil
	}

	return c.cmd.Execute(args)
}

// aggregationSpec holds parsed aggregation specification
type aggregationSpec struct {
	function string // count, sum, avg, min, max
	field    string // input field (empty for count)
	result   string // output field name
}

// parseAggregationSpecCF parses an aggregation spec from a completionflags clause
func parseAggregationSpecCF(clause cf.Clause) (*aggregationSpec, error) {
	function, _ := clause.Flags["-function"].(string)
	field, _ := clause.Flags["-field"].(string)
	result, _ := clause.Flags["-result"].(string)

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

// generateGroupByCode generates Go code for the group-by command
func generateGroupByCode(ctx *cf.Context, byField, inputFile string) error {
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
	if byField == "" {
		return fmt.Errorf("no group-by field specified (use -by)")
	}
	groupByFields := []string{byField}

	// Parse aggregation specifications from clauses
	var aggSpecs []aggregationSpec
	for _, clause := range ctx.Clauses {
		spec, err := parseAggregationSpecCF(clause)
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
