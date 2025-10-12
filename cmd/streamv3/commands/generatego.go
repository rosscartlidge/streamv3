package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/rosscartlidge/gogstools/gs"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// GenerateGoConfig holds configuration for generate-go command
type GenerateGoConfig struct {
	Output string `gs:"file,global,last,help=Output Go file (or stdout if not specified),suffix=.go"`
}

// generateGoCommand implements the generate-go command
type generateGoCommand struct {
	config *GenerateGoConfig
	cmd    *gs.GSCommand
}

func init() {
	RegisterCommand(newGenerateGoCommand())
}

func newGenerateGoCommand() *generateGoCommand {
	config := &GenerateGoConfig{}
	cmd, err := gs.NewCommand(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create generate-go command: %v", err))
	}

	return &generateGoCommand{
		config: config,
		cmd:    cmd,
	}
}

func (c *generateGoCommand) Name() string {
	return "generate-go"
}

func (c *generateGoCommand) Description() string {
	return "Generate Go code from StreamV3 CLI pipeline"
}

func (c *generateGoCommand) GetGSCommand() *gs.GSCommand {
	return c.cmd
}

func (c *generateGoCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before gs framework takes over
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

	// Delegate to gs framework which will call Config.Execute
	return c.cmd.Execute(ctx, args)
}

// Validate implements gs.Commander interface
func (c *GenerateGoConfig) Validate() error {
	return nil
}

// Execute implements gs.Commander interface
// This is called by the gs framework after parsing arguments into clauses
func (c *GenerateGoConfig) Execute(ctx context.Context, clauses []gs.ClauseSet) error {
	// Get output file from Output field or from bare arguments in clauses
	outputFile := c.Output
	if outputFile == "" && len(clauses) > 0 {
		if output, ok := clauses[0].Fields["Output"].(string); ok && output != "" {
			outputFile = output
		}
		if outputFile == "" {
			if args, ok := clauses[0].Fields["_args"].([]string); ok && len(args) > 0 {
				outputFile = args[0]
			}
		}
	}

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
}
