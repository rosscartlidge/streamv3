package commands

import (
	"fmt"
	"os"

	cf "github.com/rosscartlidge/autocli/v3"
	"github.com/rosscartlidge/ssql/v2/cmd/ssql/lib"
)

// RegisterReadJSON registers the read-json subcommand
func RegisterReadJSON(cmd *cf.CommandBuilder) *cf.CommandBuilder {
	cmd.Subcommand("read-json").
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
		Done()
	return cmd
}
