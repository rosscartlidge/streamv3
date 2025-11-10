package commands

import (
	"fmt"
	"os"

	cf "github.com/rosscartlidge/autocli/v3"
	"github.com/rosscartlidge/ssql/v2"
	"github.com/rosscartlidge/ssql/v2/cmd/ssql/lib"
)

// RegisterInclude registers the include subcommand
func RegisterInclude(cmd *cf.CommandBuilder) *cf.CommandBuilder {
	cmd.Subcommand("include").
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

			// Build included fields map
			includedMap := make(map[string]bool)
			for _, field := range fields {
				includedMap[field] = true
			}

			// Build inclusion function - delete fields not in the included list
			includer := func(r ssql.Record) ssql.Record {
				mut := r.ToMutable()
				for k := range r.All() {
					if !includedMap[k] {
						mut = mut.Delete(k)
					}
				}
				return mut.Freeze()
			}

			// Apply inclusion
			included := ssql.Select(includer)(records)

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, included); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Done()
	return cmd
}
