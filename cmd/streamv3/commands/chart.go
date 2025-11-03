package commands

import (	"fmt"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

func NChartCommand() *chartCommand {
	var xField, yField, outputFile, inputFile string
	var generate bool

	cmd := cf.NewCommand("chart").
		Description("Create interactive HTML chart from data").
		Flag("-x").
			String().
			Completer(cf.NoCompleter{Hint: "<field-name>"}).
			Bind(&xField).
			Global().
			Help("X-axis field").
			Done().
		Flag("-y").
			String().
			Completer(cf.NoCompleter{Hint: "<field-name>"}).
			Bind(&yField).
			Global().
			Help("Y-axis field").
			Done().
		Flag("-output", "-o").
			String().
			Completer(&cf.FileCompleter{Pattern: "*.html"}).
			Bind(&outputFile).
			Global().
			Default("chart.html").
			Help("Output HTML file (default: chart.html)").
			Done().
		Flag("-generate", "-g").
			Bool().
			Bind(&generate).
			Global().
			Help("Generate Go code instead of executing").
			Done().
		Flag("FILE").
			String().
			Completer(&cf.FileCompleter{Pattern: "*.jsonl"}).
			Bind(&inputFile).
			Global().
			Default("").
			Help("Input JSONL file (or stdin if not specified)").
			Done().
		Handler(func(ctx *cf.Context) error {
			// If -generate flag is set, generate Go code instead of executing
			if shouldGenerate(generate) {
				return generateChartCode(xField, yField, outputFile, inputFile)
			}

			// Normal execution: create chart
			if xField == "" {
				return fmt.Errorf("X-axis field required (use -x)")
			}
			if yField == "" {
				return fmt.Errorf("Y-axis field required (use -y)")
			}

			// Read JSONL from stdin or file
			input, err := lib.OpenInput(inputFile)
			if err != nil {
				return err
			}
			defer input.Close()

			records := lib.ReadJSONL(input)

			// Create chart using QuickChart
			if err := streamv3.QuickChart(records, xField, yField, outputFile); err != nil {
				return fmt.Errorf("creating chart: %w", err)
			}

			fmt.Printf("Chart created: %s\n", outputFile)
			return nil
		}).
		Build()

	return &chartCommand{cmd: cmd}
}

	return c.cmd.Execute(args)
}

// generateChartCode generates Go code for the chart command
func generateChartCode(xField, yField, outputFile, inputFile string) error {
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
	if xField == "" {
		return fmt.Errorf("X-axis field required (use -x)")
	}
	if yField == "" {
		return fmt.Errorf("Y-axis field required (use -y)")
	}

	if outputFile == "" {
		outputFile = "chart.html"
	}

	// Generate QuickChart call
	code := fmt.Sprintf(`streamv3.QuickChart(%s, %q, %q, %q)`, inputVar, xField, yField, outputFile)

	// Create final fragment (no output variable, it's a terminal operation)
	frag := lib.NewFinalFragment(inputVar, code, nil, getCommandString())
	return lib.WriteCodeFragment(frag)
}
