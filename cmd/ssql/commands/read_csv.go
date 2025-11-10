package commands

import (
	"fmt"
	"iter"
	"os"

	cf "github.com/rosscartlidge/autocli/v3"
	"github.com/rosscartlidge/ssql/v2"
	"github.com/rosscartlidge/ssql/v2/cmd/ssql/lib"
)

// RegisterReadCSV registers the read-csv subcommand
func RegisterReadCSV(cmd *cf.CommandBuilder) *cf.CommandBuilder {
	cmd.Subcommand("read-csv").
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
		Done()
	return cmd
}
