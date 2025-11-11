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

// generateReadCSVCode generates Go code for the read-csv command
func generateReadCSVCode(filename string) error {
	// Generate ReadCSV call with error handling
	var code string
	var imports []string

	if filename == "" {
		// Reading from stdin - use ReadCSVFromReader
		code = `records := ssql.ReadCSVFromReader(os.Stdin)`
		imports = []string{"os"}
	} else {
		// Reading from file - use ReadCSV with error handling
		code = fmt.Sprintf(`records, err := ssql.ReadCSV(%q)
	if err != nil {
		return fmt.Errorf("reading CSV: %%w", err)
	}`, filename)
		imports = []string{"fmt"}
	}

	// Create init fragment (first in pipeline)
	frag := lib.NewInitFragment("records", code, imports, getCommandString())

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}
