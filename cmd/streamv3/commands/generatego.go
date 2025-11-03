package commands

import (	"fmt"
	"os"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

func NGenerateGoCommand() *generateGoCommand {
	var outputFile string

	cmd := cf.NewCommand("generate-go").
		Description("Generate Go code from StreamV3 CLI pipeline").
		Flag("OUTPUT").
			String().
			Completer(&cf.FileCompleter{Pattern: "*.go"}).
			Bind(&outputFile).
			Global().
			Default("").
			Help("Output Go file (or stdout if not specified)").
			Done().
		Handler(func(ctx *cf.Context) error {
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
		Build()

	return &generateGoCommand{cmd: cmd}
}

	return c.cmd.Execute(args)
}
