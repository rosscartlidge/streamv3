package main

import (
	"fmt"
	"iter"
	"os"
	"strings"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/version"
)

func buildRootCommand() *cf.Command {
	var verbose bool

	return cf.NewCommand("streamv3").
		Version(version.Version).
		Description("Unix-style data processing tools").

		// Root global flags
		Flag("-verbose", "-v").
			Bool().
			Bind(&verbose).
			Global().
			Help("Enable verbose output").
			Done().

		// Subcommand: limit
		Subcommand("limit").
			Description("Take first N records (SQL LIMIT)").

			Handler(func(ctx *cf.Context) error {
				var n int
				var inputFile string

				// Get flags from context
				if nVal, ok := ctx.GlobalFlags["-n"]; ok {
					n = nVal.(int)
				} else {
					return fmt.Errorf("-n flag is required")
				}

				if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
					inputFile = fileVal.(string)
				}

				if n <= 0 {
					return fmt.Errorf("limit must be positive, got %d", n)
				}

				// Read JSONL from stdin or file
				input, err := lib.OpenInput(inputFile)
				if err != nil {
					return err
				}
				defer input.Close()

				records := lib.ReadJSONL(input)

				// Apply limit
				limited := streamv3.Limit[streamv3.Record](n)(records)

				// Write output as JSONL
				if err := lib.WriteJSONL(os.Stdout, limited); err != nil {
					return fmt.Errorf("writing output: %w", err)
				}

				return nil
			}).

			Flag("-n").
				Int().
				Required().
				Global().
				Help("Number of records to take").
				Done().

			Flag("FILE").
				String().
				Completer(&cf.FileCompleter{Pattern: "*.jsonl"}).
				Global().
				Default("").
				Help("Input JSONL file (or stdin if not specified)").
				Done().

			Done().

		// Subcommand: offset
		Subcommand("offset").
			Description("Skip first N records (SQL OFFSET)").

			Handler(func(ctx *cf.Context) error {
				var n int
				var inputFile string

				if nVal, ok := ctx.GlobalFlags["-n"]; ok {
					n = nVal.(int)
				} else {
					return fmt.Errorf("-n flag is required")
				}

				if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
					inputFile = fileVal.(string)
				}

				if n < 0 {
					return fmt.Errorf("offset must be non-negative, got %d", n)
				}

				// Read JSONL from stdin or file
				input, err := lib.OpenInput(inputFile)
				if err != nil {
					return err
				}
				defer input.Close()

				records := lib.ReadJSONL(input)

				// Apply offset
				offsetted := streamv3.Offset[streamv3.Record](n)(records)

				// Write output as JSONL
				if err := lib.WriteJSONL(os.Stdout, offsetted); err != nil {
					return fmt.Errorf("writing output: %w", err)
				}

				return nil
			}).

			Flag("-n").
				Int().
				Required().
				Global().
				Help("Number of records to skip").
				Done().

			Flag("FILE").
				String().
				Completer(&cf.FileCompleter{Pattern: "*.jsonl"}).
				Global().
				Default("").
				Help("Input JSONL file (or stdin if not specified)").
				Done().

			Done().

		// Subcommand: sort
		Subcommand("sort").
			Description("Sort records by field").

			Handler(func(ctx *cf.Context) error {
				var field string
				var desc bool
				var inputFile string

				if fieldVal, ok := ctx.GlobalFlags["-field"]; ok {
					field = fieldVal.(string)
				}

				if descVal, ok := ctx.GlobalFlags["-desc"]; ok {
					desc = descVal.(bool)
				}

				if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
					inputFile = fileVal.(string)
				}

				if field == "" {
					return fmt.Errorf("no sort field specified")
				}

				// Read JSONL from stdin or file
				input, err := lib.OpenInput(inputFile)
				if err != nil {
					return err
				}
				defer input.Close()

				records := lib.ReadJSONL(input)

				// Build sort key extractor and apply sort
				var result iter.Seq[streamv3.Record]
				if desc {
					// Descending: negate numeric values
					sorter := streamv3.SortBy(func(r streamv3.Record) float64 {
						val, _ := streamv3.Get[any](r, field)
						return -extractNumeric(val)
					})
					result = sorter(records)
				} else {
					// Ascending
					sorter := streamv3.SortBy(func(r streamv3.Record) float64 {
						val, _ := streamv3.Get[any](r, field)
						return extractNumeric(val)
					})
					result = sorter(records)
				}

				// Write output as JSONL
				if err := lib.WriteJSONL(os.Stdout, result); err != nil {
					return fmt.Errorf("writing output: %w", err)
				}

				return nil
			}).

			Flag("-field", "-f").
				String().
				Completer(cf.NoCompleter{Hint: "<field-name>"}).
				Global().
				Help("Field to sort by").
				Done().

			Flag("-desc", "-d").
				Bool().
				Global().
				Help("Sort descending").
				Done().

			Flag("FILE").
				String().
				Completer(&cf.FileCompleter{Pattern: "*.jsonl"}).
				Global().
				Default("").
				Help("Input JSONL file (or stdin if not specified)").
				Done().

			Done().

		// Subcommand: distinct
		Subcommand("distinct").
			Description("Remove duplicate records").

			Handler(func(ctx *cf.Context) error {
				var inputFile string

				if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
					inputFile = fileVal.(string)
				}

				// Read JSONL from stdin or file
				input, err := lib.OpenInput(inputFile)
				if err != nil {
					return err
				}
				defer input.Close()

				records := lib.ReadJSONL(input)

				// Apply distinct using DistinctBy with JSON serialization for comparison
				distinct := streamv3.DistinctBy(func(r streamv3.Record) string {
					// Use JSON representation as unique key
					// This is simpler than making Record comparable
					json := fmt.Sprintf("%v", r)
					return json
				})(records)

				// Write output as JSONL
				if err := lib.WriteJSONL(os.Stdout, distinct); err != nil {
					return fmt.Errorf("writing output: %w", err)
				}

				return nil
			}).

			Flag("FILE").
				String().
				Completer(&cf.FileCompleter{Pattern: "*.jsonl"}).
				Global().
				Default("").
				Help("Input JSONL file (or stdin if not specified)").
				Done().

			Done().

		// Subcommand: read-csv
		Subcommand("read-csv").
			Description("Read CSV file and output JSONL stream").

			Handler(func(ctx *cf.Context) error {
				var inputFile string

				if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
					inputFile = fileVal.(string)
				}

				// Read CSV from file or stdin
				var records iter.Seq[streamv3.Record]
				if inputFile == "" {
					records = streamv3.ReadCSVFromReader(os.Stdin)
				} else {
					var err error
					records, err = streamv3.ReadCSV(inputFile)
					if err != nil {
						return fmt.Errorf("reading CSV: %w", err)
					}
				}

				// Write as JSONL to stdout
				if err := lib.WriteJSONL(os.Stdout, records); err != nil {
					return fmt.Errorf("writing JSONL: %w", err)
				}

				return nil
			}).

			Flag("FILE").
				String().
				Completer(&cf.FileCompleter{Pattern: "*.csv"}).
				Global().
				Default("").
				Help("Input CSV file (or stdin if not specified)").
				Done().

			Done().

		// Subcommand: write-csv
		Subcommand("write-csv").
			Description("Read JSONL stream and write as CSV file").

			Handler(func(ctx *cf.Context) error {
				var outputFile string

				if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
					outputFile = fileVal.(string)
				}

				// Read JSONL from stdin
				records := lib.ReadJSONL(os.Stdin)

				// Write as CSV
				if outputFile == "" {
					return streamv3.WriteCSVToWriter(records, os.Stdout)
				} else {
					return streamv3.WriteCSV(records, outputFile)
				}
			}).

			Flag("FILE").
				String().
				Completer(&cf.FileCompleter{Pattern: "*.csv"}).
				Global().
				Default("").
				Help("Output CSV file (or stdout if not specified)").
				Done().

			Done().

		// Subcommand: where
		Subcommand("where").
			Description("Filter records based on field conditions").

			Handler(func(ctx *cf.Context) error {
				var inputFile string
				if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
					inputFile = fileVal.(string)
				}

				// Build filter from clauses (OR between clauses, AND within)
				filter := func(r streamv3.Record) bool {
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
							fieldValue, exists := streamv3.Get[any](r, field)
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
				filtered := streamv3.Where(filter)(records)

				// Write output as JSONL
				if err := lib.WriteJSONL(os.Stdout, filtered); err != nil {
					return fmt.Errorf("writing output: %w", err)
				}

				return nil
			}).

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

			Done().

		// Subcommand: select
		Subcommand("select").
			Description("Select and optionally rename fields").

			Handler(func(ctx *cf.Context) error {
				var inputFile string
				if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
					inputFile = fileVal.(string)
				}

				// Build field mapping from clauses
				fieldMap := make(map[string]string) // original -> new name

				for _, clause := range ctx.Clauses {
					field, _ := clause.Flags["-field"].(string)
					if field == "" {
						continue
					}

					// Check for rename
					asName, _ := clause.Flags["-as"].(string)
					if asName != "" {
						fieldMap[field] = asName
					} else {
						fieldMap[field] = field
					}
				}

				if len(fieldMap) == 0 {
					return fmt.Errorf("no fields specified")
				}

				// Read JSONL from stdin or file
				input, err := lib.OpenInput(inputFile)
				if err != nil {
					return err
				}
				defer input.Close()

				records := lib.ReadJSONL(input)

				// Build selector function
				selector := func(r streamv3.Record) streamv3.Record {
					result := streamv3.MakeMutableRecord()
					for origField, newField := range fieldMap {
						if val, exists := streamv3.Get[any](r, origField); exists {
							result = result.SetAny(newField, val)
						}
					}
					return result.Freeze()
				}

				// Apply selection
				selected := streamv3.Select(selector)(records)

				// Write output as JSONL
				if err := lib.WriteJSONL(os.Stdout, selected); err != nil {
					return fmt.Errorf("writing output: %w", err)
				}

				return nil
			}).

			Flag("-field", "-f").
				String().
				Completer(cf.NoCompleter{Hint: "<field-name>"}).
				Local().
				Help("Field to select").
				Done().

			Flag("-as", "-a").
				String().
				Completer(cf.NoCompleter{Hint: "<new-name>"}).
				Local().
				Help("Rename field to (optional)").
				Done().

			Flag("FILE").
				String().
				Completer(&cf.FileCompleter{Pattern: "*.jsonl"}).
				Global().
				Default("").
				Help("Input JSONL file (or stdin if not specified)").
				Done().

			Done().

		// Subcommand: group-by
		Subcommand("group-by").
			Description("Group records by fields and apply aggregations").

			Handler(func(ctx *cf.Context) error {
				var inputFile string
				var byField string

				if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
					inputFile = fileVal.(string)
				}

				if byVal, ok := ctx.GlobalFlags["-by"]; ok {
					byField = byVal.(string)
				}

				if byField == "" {
					return fmt.Errorf("no group-by field specified (use -by)")
				}
				groupByFields := []string{byField}

				// Parse aggregation specifications from clauses
				type aggSpec struct {
					function string
					field    string
					result   string
				}

				var aggSpecs []aggSpec
				for _, clause := range ctx.Clauses {
					function, _ := clause.Flags["-function"].(string)
					field, _ := clause.Flags["-field"].(string)
					result, _ := clause.Flags["-result"].(string)

					// Skip empty clauses
					if function == "" && result == "" {
						continue
					}

					if function == "" {
						return fmt.Errorf("aggregation missing -function")
					}

					if result == "" {
						return fmt.Errorf("aggregation missing -result")
					}

					// Validate function
					switch function {
					case "count", "sum", "avg", "min", "max":
						// Valid
					default:
						return fmt.Errorf("unknown aggregation function: %s", function)
					}

					// For non-count functions, field is required
					if function != "count" && field == "" {
						return fmt.Errorf("aggregation function %s requires -field", function)
					}

					aggSpecs = append(aggSpecs, aggSpec{
						function: function,
						field:    field,
						result:   result,
					})
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

				// Apply GroupByFields
				grouped := streamv3.GroupByFields("_group", groupByFields...)(records)

				// Build aggregations map
				aggregations := make(map[string]streamv3.AggregateFunc)
				for _, spec := range aggSpecs {
					agg, err := buildAggregator(spec.function, spec.field)
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

			Flag("-by", "-b").
				String().
				Completer(cf.NoCompleter{Hint: "<field-name>"}).
				Global().
				Help("Field to group by").
				Done().

			Flag("-function", "-func").
				String().
				Completer(&cf.StaticCompleter{Options: []string{"count", "sum", "avg", "min", "max"}}).
				Local().
				Help("Aggregation function (count, sum, avg, min, max)").
				Done().

			Flag("-field", "-f").
				String().
				Completer(cf.NoCompleter{Hint: "<field-name>"}).
				Local().
				Help("Field to aggregate (not needed for count)").
				Done().

			Flag("-result", "-r").
				String().
				Completer(cf.NoCompleter{Hint: "<field-name>"}).
				Local().
				Help("Output field name").
				Done().

			Flag("FILE").
				String().
				Completer(&cf.FileCompleter{Pattern: "*.csv"}).
				Global().
				Default("").
				Help("Input JSONL file (or stdin if not specified)").
				Done().

			Done().

		// Root handler (when no subcommand specified)
		Handler(func(ctx *cf.Context) error {
			fmt.Println("streamv3 - Unix-style data processing tools")
			fmt.Println()
			fmt.Println("Use -help to see available subcommands")
			return nil
		}).

		Build()
}

func main() {
	cmd := buildRootCommand()
	if err := cmd.Execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
