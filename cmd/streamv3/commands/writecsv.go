package commands

import (
	"context"
	"fmt"
	"sort"

	"github.com/rosscartlidge/gogstools/gs"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// WriteCSVConfig holds configuration for write-csv command
type WriteCSVConfig struct {
	Generate bool   `gs:"flag,global,last,help=Generate Go code instead of executing"`
	Argv     string `gs:"file,global,last,help=Output CSV file (or stdout if not specified),suffix=.csv"`
}

// writeCSVCommand implements the write-csv command
type writeCSVCommand struct {
	config *WriteCSVConfig
	cmd    *gs.GSCommand
}

func init() {
	RegisterCommand(newWriteCSVCommand())
}

func newWriteCSVCommand() *writeCSVCommand {
	config := &WriteCSVConfig{}
	cmd, err := gs.NewCommand(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create write-csv command: %v", err))
	}

	return &writeCSVCommand{
		config: config,
		cmd:    cmd,
	}
}

func (c *writeCSVCommand) Name() string {
	return "write-csv"
}

func (c *writeCSVCommand) Description() string {
	return "Read JSONL stream and write as CSV file"
}

func (c *writeCSVCommand) GetGSCommand() *gs.GSCommand {
	return c.cmd
}

func (c *writeCSVCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before gs framework takes over
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("write-csv - Read JSONL stream and write as CSV file")
		fmt.Println()
		fmt.Println("Usage: streamv3 write-csv [file.csv]")
		fmt.Println()
		fmt.Println("Reads JSONL (JSON Lines) from stdin and writes as CSV.")
		fmt.Println("Field order is determined from the first record.")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 write-csv output.csv")
		fmt.Println("  streamv3 read-csv data.csv | streamv3 where -match age gt 18 | streamv3 write-csv")
		return nil
	}

	// Delegate to gs framework which will call Config.Execute
	return c.cmd.Execute(ctx, args)
}

// formatCSVRow formats a slice of strings as a CSV row
func formatCSVRow(values []string) string {
	escaped := make([]string, len(values))
	for i, val := range values {
		// Simple CSV escaping: quote if contains comma, newline, or quote
		if containsSpecialChar(val) {
			escaped[i] = fmt.Sprintf("\"%s\"", escapeQuotes(val))
		} else {
			escaped[i] = val
		}
	}

	// Simple join with commas
	result := ""
	for i, val := range escaped {
		if i > 0 {
			result += ","
		}
		result += val
	}
	return result
}

// formatCSVValue formats a Record value as a CSV string
func formatCSVValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%g", val)
	case bool:
		return fmt.Sprintf("%v", val)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

// containsSpecialChar checks if a string contains CSV special characters
func containsSpecialChar(s string) bool {
	for _, c := range s {
		if c == ',' || c == '"' || c == '\n' || c == '\r' {
			return true
		}
	}
	return false
}

// escapeQuotes escapes double quotes in a string for CSV
func escapeQuotes(s string) string {
	result := ""
	for _, c := range s {
		if c == '"' {
			result += "\"\""
		} else {
			result += string(c)
		}
	}
	return result
}

// Validate implements gs.Commander interface
func (c *WriteCSVConfig) Validate() error {
	return nil
}

// Execute implements gs.Commander interface
// This is called by the gs framework after parsing arguments into clauses
func (c *WriteCSVConfig) Execute(ctx context.Context, clauses []gs.ClauseSet) error {
	// Get output file from Argv field or from bare arguments in clauses
	outputFile := c.Argv

	// If Argv not set, check for bare arguments in _args field
	if outputFile == "" && len(clauses) > 0 {
		// Check for Argv in the clause (might be set by -argv flag)
		if argv, ok := clauses[0].Fields["Argv"].(string); ok && argv != "" {
			outputFile = argv
		}
		// Check for bare arguments in _args
		if outputFile == "" {
			if args, ok := clauses[0].Fields["_args"].([]string); ok && len(args) > 0 {
				outputFile = args[0]
			}
		}
	}

	// If -generate flag is set, generate Go code instead of executing
	if c.Generate {
		return c.generateCode(outputFile)
	}

	// Normal execution: Read JSONL from stdin
	input, err := lib.OpenInput("-")
	if err != nil {
		return err
	}
	defer input.Close()

	records := lib.ReadJSONL(input)

	// Collect records and determine field order
	var recordSlice []streamv3.Record
	var fieldOrder []string
	fieldSet := make(map[string]bool)

	for record := range records {
		// Collect field names from first record
		if len(fieldOrder) == 0 {
			for k := range record {
				if !fieldSet[k] {
					fieldOrder = append(fieldOrder, k)
					fieldSet[k] = true
				}
			}
			sort.Strings(fieldOrder) // Sort fields alphabetically
		}

		recordSlice = append(recordSlice, record)
	}

	if len(recordSlice) == 0 {
		// No records to write
		return nil
	}

	// Write CSV to output file
	output, err := lib.OpenOutput(outputFile)
	if err != nil {
		return err
	}
	defer output.Close()

	// Write header
	fmt.Fprintf(output, "%s\n", formatCSVRow(fieldOrder))

	// Write data rows
	for _, record := range recordSlice {
		values := make([]string, len(fieldOrder))
		for i, field := range fieldOrder {
			if val, ok := record[field]; ok {
				values[i] = formatCSVValue(val)
			} else {
				values[i] = ""
			}
		}
		fmt.Fprintf(output, "%s\n", formatCSVRow(values))
	}

	return nil
}

// generateCode generates Go code for the write-csv command
func (c *WriteCSVConfig) generateCode(filename string) error {
	// Read all previous code fragments from stdin
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

	// Get input variable name from last fragment
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}

	// Generate WriteCSV call
	var code string
	if filename == "" {
		code = fmt.Sprintf(`streamv3.WriteCSV("", %s)`, inputVar)
	} else {
		code = fmt.Sprintf(`streamv3.WriteCSV(%q, %s)`, filename, inputVar)
	}

	// Create final fragment (no output variable)
	frag := lib.NewFinalFragment(inputVar, code, nil)

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}
