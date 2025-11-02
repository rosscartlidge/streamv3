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

		// Subcommand: version
		Subcommand("version").
			Description("Show version information").
			Handler(func(ctx *cf.Context) error {
				fmt.Printf("streamv3 v%s\n", version.Version)
				return nil
			}).
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

		// Subcommand: join
		Subcommand("join").
			Description("Join records from two data sources (SQL JOIN)").

			Handler(func(ctx *cf.Context) error {
				var inputFile, rightFile, joinType string

				if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
					inputFile = fileVal.(string)
				}
				if rightVal, ok := ctx.GlobalFlags["-right"]; ok {
					rightFile = rightVal.(string)
				}
				if typeVal, ok := ctx.GlobalFlags["-type"]; ok {
					joinType = typeVal.(string)
				} else {
					joinType = "inner" // default
				}

				// Validate required flags
				if rightFile == "" {
					return fmt.Errorf("right-side file required (use -right)")
				}

				// Parse join condition from first clause
				var onFields []string
				var leftField, rightField string

				if len(ctx.Clauses) > 0 {
					clause := ctx.Clauses[0]

					// Get -on fields (simple equality on same field name)
					if onRaw, ok := clause.Flags["-on"]; ok {
						if onSlice, ok := onRaw.([]any); ok {
							for _, v := range onSlice {
								if field, ok := v.(string); ok && field != "" {
									onFields = append(onFields, field)
								}
							}
						}
					}

					// Get -left-field and -right-field
					if lf, ok := clause.Flags["-left-field"].(string); ok {
						leftField = lf
					}
					if rf, ok := clause.Flags["-right-field"].(string); ok {
						rightField = rf
					}
				}

				// Validate join conditions
				if len(onFields) == 0 && (leftField == "" || rightField == "") {
					return fmt.Errorf("join condition required: use -on <field> OR (-left-field <field> -right-field <field>)")
				}
				if len(onFields) > 0 && (leftField != "" || rightField != "") {
					return fmt.Errorf("cannot use both -on and -left-field/-right-field")
				}

				// Read left-side input (stdin or file)
				leftInput, err := lib.OpenInput(inputFile)
				if err != nil {
					return fmt.Errorf("opening left input: %w", err)
				}
				defer leftInput.Close()

				leftRecords := lib.ReadJSONL(leftInput)

				// Read right-side file
				var rightSeq iter.Seq[streamv3.Record]
				if strings.HasSuffix(rightFile, ".csv") {
					csvRecords, err := streamv3.ReadCSV(rightFile)
					if err != nil {
						return fmt.Errorf("reading right CSV: %w", err)
					}
					rightSeq = csvRecords
				} else {
					rightInput, err := os.Open(rightFile)
					if err != nil {
						return fmt.Errorf("opening right file: %w", err)
					}
					defer rightInput.Close()
					rightSeq = lib.ReadJSONL(rightInput)
				}

				// Build join predicate
				var predicate streamv3.JoinPredicate
				if len(onFields) > 0 {
					predicate = streamv3.OnFields(onFields...)
				} else {
					// Use different field names
					predicate = streamv3.OnCondition(func(left, right streamv3.Record) bool {
						leftVal, leftOk := streamv3.Get[any](left, leftField)
						rightVal, rightOk := streamv3.Get[any](right, rightField)
						if !leftOk || !rightOk {
							return false
						}
						return fmt.Sprintf("%v", leftVal) == fmt.Sprintf("%v", rightVal)
					})
				}

				// Apply appropriate join
				var joinFilter streamv3.Filter[streamv3.Record, streamv3.Record]
				switch joinType {
				case "inner":
					joinFilter = streamv3.InnerJoin(rightSeq, predicate)
				case "left":
					joinFilter = streamv3.LeftJoin(rightSeq, predicate)
				case "right":
					joinFilter = streamv3.RightJoin(rightSeq, predicate)
				case "full":
					joinFilter = streamv3.FullJoin(rightSeq, predicate)
				default:
					return fmt.Errorf("unsupported join type: %s", joinType)
				}

				joined := joinFilter(leftRecords)

				// Write output as JSONL
				if err := lib.WriteJSONL(os.Stdout, joined); err != nil {
					return fmt.Errorf("writing output: %w", err)
				}

				return nil
			}).

			Flag("-type", "-t").
				String().
				Completer(&cf.StaticCompleter{Options: []string{"inner", "left", "right", "full"}}).
				Global().
				Default("inner").
				Help("Join type: inner, left, right, full (default: inner)").
				Done().

			Flag("-right", "-r").
				String().
				Completer(&cf.FileCompleter{Pattern: "*.{csv,jsonl}"}).
				Global().
				Help("Right-side file to join with (CSV or JSONL)").
				Done().

			Flag("-on").
				String().
				Completer(cf.NoCompleter{Hint: "<field-name>"}).
				Accumulate().
				Local().
				Help("Field name for equality join (same name in both sides)").
				Done().

			Flag("-left-field").
				String().
				Completer(cf.NoCompleter{Hint: "<left-field>"}).
				Local().
				Help("Field name from left side").
				Done().

			Flag("-right-field").
				String().
				Completer(cf.NoCompleter{Hint: "<right-field>"}).
				Local().
				Help("Field name from right side").
				Done().

			Flag("FILE").
				String().
				Completer(&cf.FileCompleter{Pattern: "*.jsonl"}).
				Global().
				Default("").
				Help("Left-side input JSONL file (or stdin if not specified)").
				Done().

			Done().

		// Subcommand: union
		Subcommand("union").
			Description("Combine records from multiple sources (SQL UNION)").

			Handler(func(ctx *cf.Context) error {
				var inputFile string
				var unionAll bool

				if fileVal, ok := ctx.GlobalFlags["-input"]; ok {
					inputFile = fileVal.(string)
				}
				if allVal, ok := ctx.GlobalFlags["-all"]; ok {
					unionAll = allVal.(bool)
				}

				// Get additional files from -file flags
				var additionalFiles []string
				if len(ctx.Clauses) > 0 {
					clause := ctx.Clauses[0]
					if filesRaw, ok := clause.Flags["-file"]; ok {
						if filesSlice, ok := filesRaw.([]any); ok {
							for _, v := range filesSlice {
								if file, ok := v.(string); ok && file != "" {
									additionalFiles = append(additionalFiles, file)
								}
							}
						}
					}
				}

				if len(additionalFiles) == 0 {
					return fmt.Errorf("at least one file required for union (use -file)")
				}

				// Read first input (stdin or file)
				firstInput, err := lib.OpenInput(inputFile)
				if err != nil {
					return fmt.Errorf("opening first input: %w", err)
				}
				defer firstInput.Close()

				firstRecords := lib.ReadJSONL(firstInput)

				// Chain all iterators together
				combined := chainRecords(firstRecords, additionalFiles)

				// Apply distinct if not UNION ALL
				var result iter.Seq[streamv3.Record]
				if unionAll {
					result = combined
				} else {
					// Apply distinct using DistinctBy with full record key
					distinct := streamv3.DistinctBy(unionRecordToKey)
					result = distinct(combined)
				}

				// Write output as JSONL
				if err := lib.WriteJSONL(os.Stdout, result); err != nil {
					return fmt.Errorf("writing output: %w", err)
				}

				return nil
			}).

			Flag("-file", "-f").
				String().
				Completer(&cf.FileCompleter{Pattern: "*.{csv,jsonl}"}).
				Accumulate().
				Local().
				Help("Additional file to union (CSV or JSONL)").
				Done().

			Flag("-all", "-a").
				Bool().
				Global().
				Help("Keep duplicates (UNION ALL instead of UNION)").
				Done().

			Flag("-input", "-i").
				String().
				Completer(&cf.FileCompleter{Pattern: "*.jsonl"}).
				Global().
				Default("").
				Help("First input JSONL file (or stdin if not specified)").
				Done().

			Done().

		// Subcommand: exec
		Subcommand("exec").
			Description("Execute command and parse output as records").

			Handler(func(ctx *cf.Context) error {
				// Everything after -- is in ctx.RemainingArgs
				if len(ctx.RemainingArgs) == 0 {
					return fmt.Errorf("exec requires command after '--' separator (usage: streamv3 exec -- command args...)")
				}

				command := ctx.RemainingArgs[0]
				args := ctx.RemainingArgs[1:]

				// Execute command and parse output
				records, err := streamv3.ExecCommand(command, args)
				if err != nil {
					return fmt.Errorf("executing command: %w", err)
				}

				// Write as JSONL to stdout
				if err := lib.WriteJSONL(os.Stdout, records); err != nil {
					return fmt.Errorf("writing JSONL: %w", err)
				}

				return nil
			}).

			Done().

		// Subcommand: chart
		Subcommand("chart").
			Description("Create interactive HTML chart from data").

			Handler(func(ctx *cf.Context) error {
				var xField, yField, outputFile, inputFile string

				if xVal, ok := ctx.GlobalFlags["-x"]; ok {
					xField = xVal.(string)
				}
				if yVal, ok := ctx.GlobalFlags["-y"]; ok {
					yField = yVal.(string)
				}
				if outVal, ok := ctx.GlobalFlags["-output"]; ok {
					outputFile = outVal.(string)
				} else {
					outputFile = "chart.html"
				}
				if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
					inputFile = fileVal.(string)
				}

				// Validate required fields
				if xField == "" {
					return fmt.Errorf("X-axis field required (use -x)")
				}
				if yField == "" {
					return fmt.Errorf("Y-axis field required (use -y)")
				}

				// Read JSONL from stdin or file
				input, err := lib.OpenInput(inputFile)
				if err != nil {
					return err
				}
				defer input.Close()

				records := lib.ReadJSONL(input)

				// Create chart
				err = streamv3.QuickChart(records, xField, yField, outputFile)
				if err != nil {
					return fmt.Errorf("creating chart: %w", err)
				}

				fmt.Printf("Chart created: %s\n", outputFile)
				return nil
			}).

			Flag("-x").
				String().
				Completer(cf.NoCompleter{Hint: "<field-name>"}).
				Global().
				Help("X-axis field").
				Done().

			Flag("-y").
				String().
				Completer(cf.NoCompleter{Hint: "<field-name>"}).
				Global().
				Help("Y-axis field").
				Done().

			Flag("-output", "-o").
				String().
				Completer(&cf.FileCompleter{Pattern: "*.html"}).
				Global().
				Default("chart.html").
				Help("Output HTML file (default: chart.html)").
				Done().

			Flag("FILE").
				String().
				Completer(&cf.FileCompleter{Pattern: "*.jsonl"}).
				Global().
				Default("").
				Help("Input JSONL file (or stdin if not specified)").
				Done().

			Done().

		// Subcommand: generate-go
		Subcommand("generate-go").
			Description("Generate Go code from StreamV3 CLI pipeline").

			Handler(func(ctx *cf.Context) error {
				var outputFile string

				if outVal, ok := ctx.GlobalFlags["OUTPUT"]; ok {
					outputFile = outVal.(string)
				}

				// Assemble code fragments from stdin
				code, err := lib.AssembleCodeFragments(os.Stdin)
				if err != nil {
					return fmt.Errorf("assembling code fragments: %w", err)
				}

				// Write to output
				if outputFile != "" {
					if err := os.WriteFile(outputFile, []byte(code), 0644); err != nil {
						return fmt.Errorf("writing output file: %w", err)
					}
					fmt.Fprintf(os.Stderr, "Generated Go code written to %s\n", outputFile)
				} else {
					fmt.Print(code)
				}

				return nil
			}).

			Flag("OUTPUT").
				String().
				Completer(&cf.FileCompleter{Pattern: "*.go"}).
				Global().
				Default("").
				Help("Output Go file (or stdout if not specified)").
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
