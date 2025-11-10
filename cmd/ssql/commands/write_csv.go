package commands

import (
	"os"

	cf "github.com/rosscartlidge/autocli/v3"
	"github.com/rosscartlidge/ssql/v2"
	"github.com/rosscartlidge/ssql/v2/cmd/ssql/lib"
)

// RegisterWriteCSV registers the write-csv subcommand
func RegisterWriteCSV(cmd *cf.CommandBuilder) *cf.CommandBuilder {
	cmd.Subcommand("write-csv").
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
		Done()
	return cmd
}
