package commands

import (
	"context"
	"fmt"
	"os"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/gogstools/gs"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// generateGoCommand implements the generate-go command
type generateGoCommand struct {
	cmd *cf.Command
}

func init() {
	RegisterCommand(newGenerateGoCommand())
}

func newGenerateGoCommand() *generateGoCommand {
	var outputFile string

	cmd := cf.NewCommand("generate-go").
		Description("Generate Go code from StreamV3 CLI pipeline").
		Flag("OUTPUT").
			String().
			Bind(&outputFile).
			Global().
			Default("").
			FilePattern("*.go").
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

func (c *generateGoCommand) Name() string {
	return "generate-go"
}

func (c *generateGoCommand) Description() string {
	return "Generate Go code from StreamV3 CLI pipeline"
}

func (c *generateGoCommand) GetCFCommand() *cf.Command {
	return c.cmd
}

func (c *generateGoCommand) GetGSCommand() *gs.GSCommand {
	return nil // No longer using gs
}

func (c *generateGoCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before completionflags framework takes over
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("generate-go - Generate Go code from StreamV3 CLI pipeline")
		fmt.Println()
		fmt.Println("Usage: streamv3 generate-go [output.go]")
		fmt.Println("       echo 'pipeline' | streamv3 generate-go")
		fmt.Println("       streamv3 generate-go < pipeline.sh > main.go")
		fmt.Println()
		fmt.Println("Reads a StreamV3 CLI pipeline from stdin and generates equivalent Go code")
		fmt.Println("using the StreamV3 library. The generated code can be compiled into a")
		fmt.Println("production binary that runs 10-100x faster than the CLI version.")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Generate from pipeline script")
		fmt.Println("  streamv3 generate-go < pipeline.sh > main.go")
		fmt.Println()
		fmt.Println("  # Generate inline")
		fmt.Println("  echo 'streamv3 read-csv data.csv | streamv3 where -match age gt 18' | \\")
		fmt.Println("    streamv3 generate-go > main.go")
		fmt.Println()
		fmt.Println("  # Compile and run")
		fmt.Println("  streamv3 generate-go < pipeline.sh > main.go")
		fmt.Println("  go mod init myproject")
		fmt.Println("  go get github.com/rosscartlidge/streamv3@latest")
		fmt.Println("  go build -o myproject main.go")
		fmt.Println("  ./myproject")
		fmt.Println()
		fmt.Println("Workflow:")
		fmt.Println("  1. Prototype: Build pipeline interactively with CLI tools")
		fmt.Println("  2. Test: Verify results with real data")
		fmt.Println("  3. Save: Save working pipeline to shell script")
		fmt.Println("  4. Generate: Convert to Go code with generate-go")
		fmt.Println("  5. Deploy: Compile and deploy production binary")
		return nil
	}

	return c.cmd.Execute(args)
}
