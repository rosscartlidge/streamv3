package commands

import (
	"fmt"
	"os"

	cf "github.com/rosscartlidge/autocli/v3"
	"github.com/rosscartlidge/ssql/v2"
	"github.com/rosscartlidge/ssql/v2/cmd/ssql/lib"
)

// RegisterExclude registers the exclude subcommand
func RegisterExclude(cmd *cf.CommandBuilder) *cf.CommandBuilder {
	cmd.Subcommand("exclude").
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

			// Build exclusion function - delete excluded fields
			excluder := func(r ssql.Record) ssql.Record {
				mut := r.ToMutable()
				for _, field := range fields {
					mut = mut.Delete(field)
				}
				return mut.Freeze()
			}

			// Apply exclusion
			excludedRecords := ssql.Select(excluder)(records)

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, excludedRecords); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Done()
	return cmd
}
