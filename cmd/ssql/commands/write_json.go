package commands

import (
	"os"

	cf "github.com/rosscartlidge/autocli/v3"
	"github.com/rosscartlidge/ssql/v2/cmd/ssql/lib"
)

// RegisterWriteJSON registers the write-json subcommand
func RegisterWriteJSON(cmd *cf.CommandBuilder) *cf.CommandBuilder {
	cmd.Subcommand("write-json").
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
		Done()
	return cmd
}
