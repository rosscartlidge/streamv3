package main

import (
	"fmt"
	"iter"
	"os"
	"strings"
	"time"

	cf "github.com/rosscartlidge/autocli/v3"
	"github.com/rosscartlidge/ssql"
	"github.com/rosscartlidge/ssql/cmd/ssql/lib"
	"github.com/rosscartlidge/ssql/cmd/ssql/version"
)

func buildRootCommand() *cf.Command {
	return cf.NewCommand("ssql").
		Version(version.Version).
		Description("Unix-style data processing tools").

		// Root global flags
		Flag("-verbose", "-v").
		Bool().
		Global().
		Help("Enable verbose output").
		Done().

		// Subcommand: version
		Subcommand("version").
		Description("Show version information").
		Example("ssql version", "Display the current ssql version").
		Handler(func(ctx *cf.Context) error {
			fmt.Printf("ssql v%s\n", version.Version)
			return nil
		}).
		Done().

		// Subcommand: limit
		Subcommand("limit").
		Description("Take first N records (SQL LIMIT)").
		Example("ssql read-csv data.csv | ssql limit 10", "Show first 10 records").
		Example("ssql read-csv large.csv | ssql limit 100 | ssql table", "Preview first 100 records").
		Flag("-generate", "-g").
		Bool().
		Global().
		Help("Generate Go code instead of executing").
		Done().
		Flag("N").
		Int().
		Required().
		Global().
		Help("Number of records to take").
		Done().
		Handler(func(ctx *cf.Context) error {
			var n int
			var generate bool

			// Get flags from context
			if nVal, ok := ctx.GlobalFlags["N"]; ok {
				n = nVal.(int)
			} else {
				return fmt.Errorf("N argument is required")
			}

			if genVal, ok := ctx.GlobalFlags["-generate"]; ok {
				generate = genVal.(bool)
			}

			if n <= 0 {
				return fmt.Errorf("limit must be positive, got %d", n)
			}

			// Check if generation is enabled (flag or env var)
			if shouldGenerate(generate) {
				return generateLimitCode(n)
			}

			// Read JSONL from stdin
			records := lib.ReadJSONL(os.Stdin)

			// Apply limit
			limited := ssql.Limit[ssql.Record](n)(records)

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, limited); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Done().

		// Subcommand: offset
		Subcommand("offset").
		Description("Skip first N records (SQL OFFSET)").
		Example("ssql read-csv data.csv | ssql offset 10", "Skip first 10 records").
		Example("ssql read-csv data.csv | ssql offset 100 | ssql limit 10", "Get records 101-110 (pagination)").
		Flag("-generate", "-g").
		Bool().
		Global().
		Help("Generate Go code instead of executing").
		Done().
		Flag("N").
		Int().
		Required().
		Global().
		Help("Number of records to skip").
		Done().
		Handler(func(ctx *cf.Context) error {
			var n int
			var generate bool

			if nVal, ok := ctx.GlobalFlags["N"]; ok {
				n = nVal.(int)
			} else {
				return fmt.Errorf("N argument is required")
			}

			if genVal, ok := ctx.GlobalFlags["-generate"]; ok {
				generate = genVal.(bool)
			}

			if n < 0 {
				return fmt.Errorf("offset must be non-negative, got %d", n)
			}

			// Check if generation is enabled (flag or env var)
			if shouldGenerate(generate) {
				return generateOffsetCode(n)
			}

			// Read JSONL from stdin
			records := lib.ReadJSONL(os.Stdin)

			// Apply offset
			offsetted := ssql.Offset[ssql.Record](n)(records)

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, offsetted); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Done().

		// Subcommand: sort
		Subcommand("sort").
		Description("Sort records by field").
		Example("ssql read-csv data.csv | ssql sort age", "Sort by age ascending").
		Example("ssql read-csv sales.csv | ssql sort amount -desc", "Sort by amount descending").
		Flag("FIELD").
		String().
		Required().
		Completer(cf.NoCompleter{Hint: "<field-name>"}).
		Global().
		Help("Field to sort by").
		Done().
		Flag("-generate", "-g").
		Bool().
		Global().
		Help("Generate Go code instead of executing").
		Done().
		Flag("-desc", "-d").
		Bool().
		Global().
		Help("Sort descending").
		Done().
		Handler(func(ctx *cf.Context) error {
			var field string
			var desc bool
			var generate bool

			if fieldVal, ok := ctx.GlobalFlags["FIELD"]; ok {
				field = fieldVal.(string)
			}

			if descVal, ok := ctx.GlobalFlags["-desc"]; ok {
				desc = descVal.(bool)
			}

			if genVal, ok := ctx.GlobalFlags["-generate"]; ok {
				generate = genVal.(bool)
			}

			if field == "" {
				return fmt.Errorf("no sort field specified")
			}

			// Check if generation is enabled (flag or env var)
			if shouldGenerate(generate) {
				return generateSortCode(field, desc)
			}

			// Read JSONL from stdin
			records := lib.ReadJSONL(os.Stdin)

			// Build sort key extractor and apply sort
			var result iter.Seq[ssql.Record]
			if desc {
				// Descending: negate numeric values
				sorter := ssql.SortBy(func(r ssql.Record) float64 {
					val, _ := ssql.Get[any](r, field)
					return -extractNumeric(val)
				})
				result = sorter(records)
			} else {
				// Ascending
				sorter := ssql.SortBy(func(r ssql.Record) float64 {
					val, _ := ssql.Get[any](r, field)
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
		Done().

		// Subcommand: distinct
		Subcommand("distinct").
		Description("Remove duplicate records").
		Example("ssql read-csv data.csv | ssql distinct", "Remove duplicate records").
		Example("ssql read-csv users.csv | ssql include email | ssql distinct", "Get unique email addresses").
		Flag("-generate", "-g").
		Bool().
		Global().
		Help("Generate Go code instead of executing").
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
				return generateDistinctCode()
			}

			// Read JSONL from stdin or file
			input, err := lib.OpenInput(inputFile)
			if err != nil {
				return err
			}
			defer input.Close()

			records := lib.ReadJSONL(input)

			// Apply distinct using DistinctBy with JSON serialization for comparison
			distinct := ssql.DistinctBy(func(r ssql.Record) string {
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
		Done().

		// Subcommand: read-csv
		Subcommand("read-csv").
		Description("Read CSV file and output JSONL stream").
		Example("ssql read-csv data.csv | ssql table", "Read CSV and display as table").
		Example("cat data.csv | ssql read-csv | ssql limit 10", "Read from stdin and show first 10 records").
		Flag("-generate", "-g").
		Bool().
		Global().
		Help("Generate Go code instead of executing").
		Done().
		Flag("FILE").
		String().
		Completer(&cf.FileCompleter{Pattern: "*.csv"}).
		Global().
		Default("").
		Help("Input CSV file (or stdin if not specified)").
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
				return generateReadCSVCode(inputFile)
			}

			// Read CSV from file or stdin
			var records iter.Seq[ssql.Record]
			if inputFile == "" {
				records = ssql.ReadCSVFromReader(os.Stdin)
			} else {
				var err error
				records, err = ssql.ReadCSV(inputFile)
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
		Done().

		// Subcommand: write-csv
		Subcommand("write-csv").
		Description("Read JSONL stream and write as CSV file").
		Example("ssql read-json data.json | ssql write-csv output.csv", "Convert JSON to CSV").
		Example("ssql read-csv data.csv | ssql where -match status eq active | ssql write-csv active.csv", "Filter and save to CSV").
		Flag("-generate", "-g").
		Bool().
		Global().
		Help("Generate Go code instead of executing").
		Done().
		Flag("FILE").
		String().
		Completer(&cf.FileCompleter{Pattern: "*.csv"}).
		Global().
		Default("").
		Help("Output CSV file (or stdout if not specified)").
		Done().
		Handler(func(ctx *cf.Context) error {
			var outputFile string
			var generate bool

			if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
				outputFile = fileVal.(string)
			}

			if genVal, ok := ctx.GlobalFlags["-generate"]; ok {
				generate = genVal.(bool)
			}

			// Check if generation is enabled (flag or env var)
			if shouldGenerate(generate) {
				return generateWriteCSVCode(outputFile)
			}

			// Read JSONL from stdin
			records := lib.ReadJSONL(os.Stdin)

			// Write as CSV
			if outputFile == "" {
				return ssql.WriteCSVToWriter(records, os.Stdout)
			} else {
				return ssql.WriteCSV(records, outputFile)
			}
		}).
		Done().

		// Subcommand: read-json
		Subcommand("read-json").
		Description("Read JSON array or JSONL file (auto-detects format)").
		Example("ssql read-json data.jsonl | ssql table", "Read JSONL file and display as table").
		Example("ssql read-json array.json | ssql where -match status eq active", "Read JSON array and filter records").
		Flag("-generate", "-g").
		Bool().
		Global().
		Help("Generate Go code instead of executing").
		Done().
		Flag("FILE").
		String().
		Completer(&cf.FileCompleter{Pattern: "*.{json,jsonl}"}).
		Global().
		Required().
		Help("Input JSON/JSONL file").
		Done().
		Handler(func(ctx *cf.Context) error {
			var inputFile string
			var generate bool

			if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
				inputFile = fileVal.(string)
			} else {
				return fmt.Errorf("FILE is required")
			}

			if genVal, ok := ctx.GlobalFlags["-generate"]; ok {
				generate = genVal.(bool)
			}

			// Check if generation is enabled (flag or env var)
			if shouldGenerate(generate) {
				return generateReadJSONCode(inputFile)
			}

			// Open and read JSON file
			input, err := lib.OpenInput(inputFile)
			if err != nil {
				return err
			}
			defer input.Close()

			records := lib.ReadJSON(input)

			// Write as JSONL to stdout
			if err := lib.WriteJSONL(os.Stdout, records); err != nil {
				return fmt.Errorf("writing JSONL: %w", err)
			}

			return nil
		}).
		Done().

		// Subcommand: write-json
		Subcommand("write-json").
		Description("Write as JSONL (default) or pretty JSON array (-pretty)").
		Example("ssql read-csv data.csv | ssql write-json", "Convert CSV to JSONL").
		Example("ssql read-csv data.csv | ssql write-json -pretty > output.json", "Convert CSV to pretty JSON array").
		Flag("-generate", "-g").
		Bool().
		Global().
		Help("Generate Go code instead of executing").
		Done().
		Flag("-pretty", "-p").
		Bool().
		Global().
		Help("Pretty-print as JSON array (default: JSONL)").
		Done().
		Flag("FILE").
		String().
		Completer(&cf.FileCompleter{Pattern: "*.{json,jsonl}"}).
		Global().
		Default("").
		Help("Output JSON/JSONL file (or stdout if not specified)").
		Done().
		Handler(func(ctx *cf.Context) error {
			var outputFile string
			var pretty bool
			var generate bool

			if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
				outputFile = fileVal.(string)
			}

			if prettyVal, ok := ctx.GlobalFlags["-pretty"]; ok {
				pretty = prettyVal.(bool)
			}

			if genVal, ok := ctx.GlobalFlags["-generate"]; ok {
				generate = genVal.(bool)
			}

			// Check if generation is enabled (flag or env var)
			if shouldGenerate(generate) {
				return generateWriteJSONCode(outputFile, pretty)
			}

			// Read JSONL from stdin
			records := lib.ReadJSONL(os.Stdin)

			// Write to stdout or file
			if outputFile == "" {
				return lib.WriteJSON(os.Stdout, records, pretty)
			} else {
				output, err := lib.OpenOutput(outputFile)
				if err != nil {
					return err
				}
				defer output.Close()
				return lib.WriteJSON(output, records, pretty)
			}
		}).
		Done().

		// Subcommand: table
		Subcommand("table").
		Description("Display records as a formatted table").
		Example("ssql read-csv data.csv | ssql table", "Display CSV as formatted table").
		Example("ssql read-csv data.csv | ssql where -match age gt 21 | ssql table -max-width 30", "Filter and display with custom column width").
		Flag("-generate", "-g").
		Bool().
		Global().
		Help("Generate Go code instead of executing").
		Done().
		Flag("-max-width").
		Int().
		Global().
		Default(50).
		Help("Maximum column width (truncate longer values)").
		Done().
		Handler(func(ctx *cf.Context) error {
			var generate bool
			var maxWidth int

			if genVal, ok := ctx.GlobalFlags["-generate"]; ok {
				generate = genVal.(bool)
			}

			if widthVal, ok := ctx.GlobalFlags["-max-width"]; ok {
				maxWidth = widthVal.(int)
			}

			// Check if generation is enabled (flag or env var)
			if shouldGenerate(generate) {
				return generateTableCode(maxWidth)
			}

			// Read all records from stdin and display as table
			records := lib.ReadJSONL(os.Stdin)
			ssql.DisplayTable(records, maxWidth)
			return nil
		}).
		Done().

		// Subcommand: where
		Subcommand("where").
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
		Done().

		// Subcommand: update
		Subcommand("update").
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
							parsedValue := parseValue(upd.value)
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
		Done().

		// Subcommand: include
		Subcommand("include").
		Description("Include only specified fields").
		Example("ssql read-csv data.csv | ssql include name age", "Select only name and age columns").
		Example("ssql read-json users.json | ssql include email status | ssql write-csv out.csv", "Extract email and status to CSV").
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
		Help("Fields to include").
		Done().
		Handler(func(ctx *cf.Context) error {
			var generate bool
			var fields []string

			if genVal, ok := ctx.GlobalFlags["-generate"]; ok {
				generate = genVal.(bool)
			}

			if fieldsVal, ok := ctx.GlobalFlags["FIELDS"]; ok {
				switch v := fieldsVal.(type) {
				case []string:
					fields = v
				case []any:
					for _, item := range v {
						if s, ok := item.(string); ok {
							fields = append(fields, s)
						}
					}
				case string:
					fields = []string{v}
				}
			}

			if len(fields) == 0 {
				return fmt.Errorf("no fields specified")
			}

			// Check if generation is enabled (flag or env var)
			if shouldGenerate(generate) {
				return generateIncludeCode(fields)
			}

			// Read JSONL from stdin
			records := lib.ReadJSONL(os.Stdin)

			// Build inclusion function
			includer := func(r ssql.Record) ssql.Record {
				result := ssql.MakeMutableRecord()
				for _, field := range fields {
					if val, exists := ssql.Get[any](r, field); exists {
						result = result.SetAny(field, val)
					}
				}
				return result.Freeze()
			}

			// Apply inclusion
			included := ssql.Select(includer)(records)

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, included); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Done().

		// Subcommand: exclude
		Subcommand("exclude").
		Description("Exclude specified fields").
		Example("ssql read-csv data.csv | ssql exclude id created_at updated_at", "Remove metadata fields").
		Example("ssql read-json api.json | ssql exclude password token secret_key", "Remove sensitive fields").
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
		Help("Fields to exclude").
		Done().
		Handler(func(ctx *cf.Context) error {
			var generate bool
			var fields []string

			if genVal, ok := ctx.GlobalFlags["-generate"]; ok {
				generate = genVal.(bool)
			}

			if fieldsVal, ok := ctx.GlobalFlags["FIELDS"]; ok {
				switch v := fieldsVal.(type) {
				case []string:
					fields = v
				case []any:
					for _, item := range v {
						if s, ok := item.(string); ok {
							fields = append(fields, s)
						}
					}
				case string:
					fields = []string{v}
				}
			}

			if len(fields) == 0 {
				return fmt.Errorf("no fields specified")
			}

			// Check if generation is enabled (flag or env var)
			if shouldGenerate(generate) {
				return generateExcludeCode(fields)
			}

			// Read JSONL from stdin
			records := lib.ReadJSONL(os.Stdin)

			// Build exclusion map
			excluded := make(map[string]bool)
			for _, field := range fields {
				excluded[field] = true
			}

			// Build exclusion function
			excluder := func(r ssql.Record) ssql.Record {
				result := ssql.MakeMutableRecord()
				for k, v := range r.All() {
					if !excluded[k] {
						result = result.SetAny(k, v)
					}
				}
				return result.Freeze()
			}

			// Apply exclusion
			excludedRecords := ssql.Select(excluder)(records)

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, excludedRecords); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Done().

		// Subcommand: rename
		Subcommand("rename").
		Description("Rename fields").
		Example("ssql read-csv data.csv | ssql rename -as oldname newname", "Rename a single field").
		Example("ssql read-csv users.csv | ssql rename -as first_name firstName -as last_name lastName", "Rename multiple fields to camelCase").
		Flag("-generate", "-g").
		Bool().
		Global().
		Help("Generate Go code instead of executing").
		Done().
		Flag("-as").
		Arg("old-field").Completer(cf.NoCompleter{Hint: "<field-name>"}).Done().
		Arg("new-field").Completer(cf.NoCompleter{Hint: "<new-name>"}).Done().
		Accumulate().
		Global().
		Help("Rename old-field to new-field").
		Done().
		Handler(func(ctx *cf.Context) error {
			var generate bool

			if genVal, ok := ctx.GlobalFlags["-generate"]; ok {
				generate = genVal.(bool)
			}

			// Extract rename mappings from -as flags
			var renames []struct{ oldField, newField string }
			if asVal, ok := ctx.GlobalFlags["-as"]; ok {
				asSlice, ok := asVal.([]any)
				if !ok {
					return fmt.Errorf("invalid -as flag format")
				}
				for _, item := range asSlice {
					asMap, ok := item.(map[string]any)
					if !ok {
						return fmt.Errorf("invalid -as flag: expected map format")
					}
					oldField, _ := asMap["old-field"].(string)
					newField, _ := asMap["new-field"].(string)
					if oldField == "" || newField == "" {
						return fmt.Errorf("invalid -as flag: both old-field and new-field are required")
					}
					renames = append(renames, struct{ oldField, newField string }{oldField, newField})
				}
			}

			if len(renames) == 0 {
				return fmt.Errorf("no renames specified")
			}

			// Check if generation is enabled (flag or env var)
			if shouldGenerate(generate) {
				return generateRenameCode(renames)
			}

			// Read JSONL from stdin
			records := lib.ReadJSONL(os.Stdin)

			// Build rename map
			renameMap := make(map[string]string)
			for _, r := range renames {
				renameMap[r.oldField] = r.newField
			}

			// Build renamer function
			renamer := func(r ssql.Record) ssql.Record {
				result := ssql.MakeMutableRecord()
				for k, v := range r.All() {
					if newName, ok := renameMap[k]; ok {
						result = result.SetAny(newName, v)
					} else {
						result = result.SetAny(k, v)
					}
				}
				return result.Freeze()
			}

			// Apply rename
			renamed := ssql.Select(renamer)(records)

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, renamed); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Done().

		// Subcommand: group-by
		Subcommand("group-by").
		Description("Group records by fields and apply aggregations").
		Example("ssql read-csv sales.csv | ssql group-by region -func count -result total", "Count records by region").
		Example("ssql read-csv sales.csv | ssql group-by region -func sum -field amount -result total_sales", "Sum sales amount by region").
		Example("ssql read-csv data.csv | ssql group-by dept status -func count -result count + -func avg -field salary -result avg_salary", "Group by dept and status with multiple aggregations").
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
		Done().

		// Subcommand: join
		Subcommand("join").
		Description("Join records from two data sources (SQL JOIN)").
		Example("ssql read-csv users.csv | ssql join -right orders.csv -on user_id", "Inner join users and orders on user_id").
		Example("ssql read-csv employees.csv | ssql join -type left -right departments.csv -on dept_id", "Left join employees with departments").
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
			var rightSeq iter.Seq[ssql.Record]
			if strings.HasSuffix(rightFile, ".csv") {
				csvRecords, err := ssql.ReadCSV(rightFile)
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
			var predicate ssql.JoinPredicate
			if len(onFields) > 0 {
				predicate = ssql.OnFields(onFields...)
			} else {
				// Use different field names
				predicate = ssql.OnCondition(func(left, right ssql.Record) bool {
					leftVal, leftOk := ssql.Get[any](left, leftField)
					rightVal, rightOk := ssql.Get[any](right, rightField)
					if !leftOk || !rightOk {
						return false
					}
					return fmt.Sprintf("%v", leftVal) == fmt.Sprintf("%v", rightVal)
				})
			}

			// Apply appropriate join
			var joinFilter ssql.Filter[ssql.Record, ssql.Record]
			switch joinType {
			case "inner":
				joinFilter = ssql.InnerJoin(rightSeq, predicate)
			case "left":
				joinFilter = ssql.LeftJoin(rightSeq, predicate)
			case "right":
				joinFilter = ssql.RightJoin(rightSeq, predicate)
			case "full":
				joinFilter = ssql.FullJoin(rightSeq, predicate)
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
		Done().

		// Subcommand: union
		Subcommand("union").
		Description("Combine records from multiple sources (SQL UNION)").
		Example("ssql read-csv 2023.csv | ssql union -file 2024.csv", "Combine records from two CSV files (removes duplicates)").
		Example("ssql read-csv east.csv | ssql union -all -file west.csv -file south.csv", "Combine three files keeping all records (UNION ALL)").
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
			var result iter.Seq[ssql.Record]
			if unionAll {
				result = combined
			} else {
				// Apply distinct using DistinctBy with full record key
				distinct := ssql.DistinctBy(unionRecordToKey)
				result = distinct(combined)
			}

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, result); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Done().

		// Subcommand: exec
		Subcommand("exec").
		Description("Execute command and parse output as records").
		Example("ssql exec -- ps aux | ssql where -match USER eq root", "Parse ps output and filter for root processes").
		Example("ssql exec -- ls -la | ssql include FILE SIZE", "Parse ls output and select specific fields").
		Handler(func(ctx *cf.Context) error {
			// Everything after -- is in ctx.RemainingArgs
			if len(ctx.RemainingArgs) == 0 {
				return fmt.Errorf("exec requires command after '--' separator (usage: ssql exec -- command args...)")
			}

			command := ctx.RemainingArgs[0]
			args := ctx.RemainingArgs[1:]

			// Execute command and parse output
			records, err := ssql.ExecCommand(command, args)
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
		Example("ssql read-csv data.csv | ssql chart -x date -y revenue", "Create line chart of revenue over time").
		Example("ssql read-csv sales.csv | ssql chart -x product -y sales -output sales.html", "Create chart with custom output file").
		Flag("-generate", "-g").
		Bool().
		Global().
		Help("Generate Go code instead of executing").
		Done().
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
		Handler(func(ctx *cf.Context) error {
			var xField, yField, outputFile, inputFile string
			var generate bool

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
			if genVal, ok := ctx.GlobalFlags["-generate"]; ok {
				generate = genVal.(bool)
			}

			// Check if generation is enabled (flag or env var)
			if shouldGenerate(generate) {
				return generateChartCode(xField, yField, outputFile)
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
			err = ssql.QuickChart(records, xField, yField, outputFile)
			if err != nil {
				return fmt.Errorf("creating chart: %w", err)
			}

			fmt.Printf("Chart created: %s\n", outputFile)
			return nil
		}).
		Done().

		// Subcommand: generate-go
		Subcommand("generate-go").
		Description("Generate Go code from StreamV3 CLI pipeline").
		Example("ssql read-csv -g data.csv | ssql where -g -match age gt 18 | ssql generate-go", "Generate Go code from pipeline").
		Example("(export STREAMV3_GENERATE_GO=1 && ssql read-csv data.csv | ssql limit 10 | ssql generate-go) > prog.go", "Generate using environment variable").
		Flag("OUTPUT").
		String().
		Completer(&cf.FileCompleter{Pattern: "*.go"}).
		Global().
		Default("").
		Help("Output Go file (or stdout if not specified)").
		Done().
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
		Done().

		// Root handler (when no subcommand specified)
		Handler(func(ctx *cf.Context) error {
			fmt.Println("ssql - Unix-style data processing tools")
			fmt.Println()
			fmt.Println("Use -help to see available subcommands")
			fmt.Println()
			fmt.Println("To enable tab completion, add to your ~/.bashrc:")
			fmt.Println("  eval \"$(ssql -completion-script)\"")
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
