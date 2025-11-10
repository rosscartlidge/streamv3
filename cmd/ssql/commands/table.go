package commands

import (
	"os"

	cf "github.com/rosscartlidge/autocli/v3"
	"github.com/rosscartlidge/ssql/v2"
	"github.com/rosscartlidge/ssql/v2/cmd/ssql/lib"
)

// RegisterTable registers the table subcommand
func RegisterTable(cmd *cf.CommandBuilder) *cf.CommandBuilder {
	cmd.Subcommand("table").
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
		Done()
	return cmd
}
