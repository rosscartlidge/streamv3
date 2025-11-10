package commands

import (
	"fmt"
	"os"

	cf "github.com/rosscartlidge/autocli/v3"
	"github.com/rosscartlidge/ssql/v2"
	"github.com/rosscartlidge/ssql/v2/cmd/ssql/lib"
)

// RegisterGroupBy registers the group-by subcommand
func RegisterGroupBy(cmd *cf.CommandBuilder) *cf.CommandBuilder {
	cmd.Subcommand("group-by").
		Description("Group records by fields and apply aggregations").
		Example("ssql read-csv sales.csv | ssql group-by region -count total", "Count records by region").
		Example("ssql read-csv sales.csv | ssql group-by region -sum amount total_sales", "Sum sales amount by region").
		Example("ssql read-csv data.csv | ssql group-by dept -count num_employees -avg salary avg_salary -sum hours total_hours", "Multiple aggregations in one command").
		Flag("-generate", "-g").
			Bool().
			Global().
			Help("Generate Go code instead of executing").
		Done().
		Flag("FIELDS").
			String().
			Variadic().
			Completer(cf.NoCompleter{Hint: "<field-name>"}).
			Global().
			Help("Fields to group by").
		Done().
		Flag("-count").
			Arg("result-name").Completer(cf.NoCompleter{Hint: "<name>"}).Done().
			Accumulate().
			Global().
			Help("Count records (result field name)").
		Done().
		Flag("-sum").
			Arg("field").Completer(cf.NoCompleter{Hint: "<field>"}).Done().
			Arg("result-name").Completer(cf.NoCompleter{Hint: "<name>"}).Done().
			Accumulate().
			Global().
			Help("Sum field values (field name, result name)").
		Done().
		Flag("-avg").
			Arg("field").Completer(cf.NoCompleter{Hint: "<field>"}).Done().
			Arg("result-name").Completer(cf.NoCompleter{Hint: "<name>"}).Done().
			Accumulate().
			Global().
			Help("Average field values (field name, result name)").
		Done().
		Flag("-min").
			Arg("field").Completer(cf.NoCompleter{Hint: "<field>"}).Done().
			Arg("result-name").Completer(cf.NoCompleter{Hint: "<name>"}).Done().
			Accumulate().
			Global().
			Help("Minimum field value (field name, result name)").
		Done().
		Flag("-max").
			Arg("field").Completer(cf.NoCompleter{Hint: "<field>"}).Done().
			Arg("result-name").Completer(cf.NoCompleter{Hint: "<name>"}).Done().
			Accumulate().
			Global().
			Help("Maximum field value (field name, result name)").
		Done().
		Handler(func(ctx *cf.Context) error {
			var groupByFields []string
			var generate bool

			// Extract group-by fields from variadic positional
			if fieldsVal, ok := ctx.GlobalFlags["FIELDS"]; ok {
				switch v := fieldsVal.(type) {
				case []string:
					groupByFields = v
				case []any:
					for _, item := range v {
						if s, ok := item.(string); ok {
							groupByFields = append(groupByFields, s)
						}
					}
				case string:
					groupByFields = []string{v}
				}
			}

			if genVal, ok := ctx.GlobalFlags["-generate"]; ok {
				generate = genVal.(bool)
			}

			if len(groupByFields) == 0 {
				return fmt.Errorf("no group-by fields specified")
			}

			// Check if generation is enabled (flag or env var)
			if shouldGenerate(generate) {
				return generateGroupByCode(ctx, groupByFields)
			}

			// Parse aggregation specifications from new flag format
			type aggSpec struct {
				function string
				field    string
				result   string
			}

			var aggSpecs []aggSpec

			// Parse -count flags (only result name)
			if countVals, ok := ctx.GlobalFlags["-count"]; ok {
				counts, _ := countVals.([]any)
				for _, countVal := range counts {
					// When there's only 1 Arg(), autocli doesn't wrap in a slice
					if resultName, ok := countVal.(string); ok {
						aggSpecs = append(aggSpecs, aggSpec{
							function: "count",
							field:    "",
							result:   resultName,
						})
					}
				}
			}

			// Parse -sum flags (field and result name)
			if sumVals, ok := ctx.GlobalFlags["-sum"]; ok {
				sums, _ := sumVals.([]any)
				for _, sumVal := range sums {
					// When there are 2+ Args(), autocli returns a map with arg names as keys
					if argsMap, ok := sumVal.(map[string]any); ok {
						field, _ := argsMap["field"].(string)
						result, _ := argsMap["result-name"].(string)
						if field != "" && result != "" {
							aggSpecs = append(aggSpecs, aggSpec{
								function: "sum",
								field:    field,
								result:   result,
							})
						}
					}
				}
			}

			// Parse -avg flags (field and result name)
			if avgVals, ok := ctx.GlobalFlags["-avg"]; ok {
				avgs, _ := avgVals.([]any)
				for _, avgVal := range avgs {
					if argsMap, ok := avgVal.(map[string]any); ok {
						field, _ := argsMap["field"].(string)
						result, _ := argsMap["result-name"].(string)
						if field != "" && result != "" {
							aggSpecs = append(aggSpecs, aggSpec{
								function: "avg",
								field:    field,
								result:   result,
							})
						}
					}
				}
			}

			// Parse -min flags (field and result name)
			if minVals, ok := ctx.GlobalFlags["-min"]; ok {
				mins, _ := minVals.([]any)
				for _, minVal := range mins {
					if argsMap, ok := minVal.(map[string]any); ok {
						field, _ := argsMap["field"].(string)
						result, _ := argsMap["result-name"].(string)
						if field != "" && result != "" {
							aggSpecs = append(aggSpecs, aggSpec{
								function: "min",
								field:    field,
								result:   result,
							})
						}
					}
				}
			}

			// Parse -max flags (field and result name)
			if maxVals, ok := ctx.GlobalFlags["-max"]; ok {
				maxs, _ := maxVals.([]any)
				for _, maxVal := range maxs {
					if argsMap, ok := maxVal.(map[string]any); ok {
						field, _ := argsMap["field"].(string)
						result, _ := argsMap["result-name"].(string)
						if field != "" && result != "" {
							aggSpecs = append(aggSpecs, aggSpec{
								function: "max",
								field:    field,
								result:   result,
							})
						}
					}
				}
			}

			if len(aggSpecs) == 0 {
				return fmt.Errorf("no aggregations specified (use -count, -sum, -avg, -min, or -max)")
			}

			// Read JSONL from stdin
			records := lib.ReadJSONL(os.Stdin)

			// Apply GroupByFields
			grouped := ssql.GroupByFields("_group", groupByFields...)(records)

			// Build aggregations map
			aggregations := make(map[string]ssql.AggregateFunc)
			for _, spec := range aggSpecs {
				agg, err := buildAggregator(spec.function, spec.field)
				if err != nil {
					return err
				}
				aggregations[spec.result] = agg
			}

			// Apply Aggregate
			aggregated := ssql.Aggregate("_group", aggregations)(grouped)

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, aggregated); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Done()
	return cmd
}
