package commands

import (
	"context"
	"fmt"

	"github.com/rosscartlidge/gogstools/gs"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// ChartConfig holds configuration for chart command
type ChartConfig struct {
	X        string `gs:"field,global,last,help=X-axis field"`
	Y        string `gs:"field,global,last,help=Y-axis field"`
	Output   string `gs:"file,global,last,help=Output HTML file,suffix=.html"`
	Generate bool   `gs:"flag,global,last,help=Generate Go code instead of executing"`
	Argv     string `gs:"file,global,last,help=Input JSONL file (or stdin if not specified),suffix=.jsonl"`
}

// chartCommand implements the chart command
type chartCommand struct {
	config *ChartConfig
	cmd    *gs.GSCommand
}

func init() {
	RegisterCommand(newChartCommand())
}

func newChartCommand() *chartCommand {
	config := &ChartConfig{}
	cmd, err := gs.NewCommand(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create chart command: %v", err))
	}

	return &chartCommand{
		config: config,
		cmd:    cmd,
	}
}

func (c *chartCommand) Name() string {
	return "chart"
}

func (c *chartCommand) Description() string {
	return "Create interactive HTML chart from data"
}

func (c *chartCommand) GetGSCommand() *gs.GSCommand {
	return c.cmd
}

func (c *chartCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before gs framework takes over
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("chart - Create interactive HTML chart from data")
		fmt.Println()
		fmt.Println("Usage: streamv3 chart -x <field> -y <field> -output <file.html>")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  -x <field>        Field for X-axis")
		fmt.Println("  -y <field>        Field for Y-axis")
		fmt.Println("  -output <file>    Output HTML file (default: chart.html)")
		fmt.Println()
		fmt.Println("Creates an interactive Chart.js visualization with:")
		fmt.Println("  - Field selection dropdown")
		fmt.Println("  - Zoom and pan controls")
		fmt.Println("  - Data table view")
		fmt.Println("  - Export to PNG/CSV")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Simple scatter plot")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 chart -x age -y salary -output viz.html")
		fmt.Println()
		fmt.Println("  # After aggregation")
		fmt.Println("  streamv3 read-csv sales.csv | \\")
		fmt.Println("    streamv3 group-by -by region -function sum -field amount -result total | \\")
		fmt.Println("    streamv3 chart -x region -y total -output sales.html")
		return nil
	}

	// Delegate to gs framework which will call Config.Execute
	return c.cmd.Execute(ctx, args)
}

// Validate implements gs.Commander interface
func (c *ChartConfig) Validate() error {
	return nil
}

// Execute implements gs.Commander interface
func (c *ChartConfig) Execute(ctx context.Context, clauses []gs.ClauseSet) error {
	// If -generate flag is set, generate Go code instead of executing
	if c.Generate {
		return c.generateCode(clauses)
	}

	// Normal execution: create chart
	if c.X == "" {
		return fmt.Errorf("X-axis field required (use -x)")
	}
	if c.Y == "" {
		return fmt.Errorf("Y-axis field required (use -y)")
	}

	outputFile := c.Output
	if outputFile == "" {
		outputFile = "chart.html"
	}

	// Read JSONL from stdin or file
	input, err := lib.OpenInput(c.Argv)
	if err != nil {
		return err
	}
	defer input.Close()

	records := lib.ReadJSONL(input)

	// Create chart using QuickChart
	if err := streamv3.QuickChart(records, c.X, c.Y, outputFile); err != nil {
		return fmt.Errorf("creating chart: %w", err)
	}

	fmt.Printf("Chart created: %s\n", outputFile)
	return nil
}

// generateCode generates Go code for the chart command
func (c *ChartConfig) generateCode(clauses []gs.ClauseSet) error {
	// Read all previous code fragments from stdin (if any)
	fragments, err := lib.ReadAllCodeFragments()
	if err != nil {
		return fmt.Errorf("reading code fragments: %w", err)
	}

	// Pass through all previous fragments
	for _, frag := range fragments {
		if err := lib.WriteCodeFragment(frag); err != nil {
			return fmt.Errorf("writing previous fragment: %w", err)
		}
	}

	// Get input variable from last fragment
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}

	// Validate required fields
	if c.X == "" {
		return fmt.Errorf("X-axis field required (use -x)")
	}
	if c.Y == "" {
		return fmt.Errorf("Y-axis field required (use -y)")
	}

	outputFile := c.Output
	if outputFile == "" {
		outputFile = "chart.html"
	}

	// Generate QuickChart call
	code := fmt.Sprintf(`streamv3.QuickChart(%s, %q, %q, %q)`, inputVar, c.X, c.Y, outputFile)

	// Create final fragment (no output variable, it's a terminal operation)
	frag := lib.NewFinalFragment(inputVar, code, nil)
	return lib.WriteCodeFragment(frag)
}
