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
	Argv string `gs:"file,global,last,help=Output CSV file (or stdout if not specified),suffix=.csv"`
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

func (c *writeCSVCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag
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
		fmt.Println("  streamv3 read-csv data.csv | streamv3 where - field age - op gt - value 18 | streamv3 write-csv")
		return nil
	}

	// Parse arguments using gs framework
	clauses, err := c.cmd.Parse(args)
	if err != nil {
		return fmt.Errorf("parsing arguments: %w", err)
	}

	// Get output file from config (set by gs framework)
	outputFile := c.config.Argv

	// Read JSONL from stdin
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

	// Suppress unused warning for clauses
	_ = clauses

	return nil
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

// Execute implements gs.Commander interface (not used, we use command.Execute)
func (c *WriteCSVConfig) Execute(ctx context.Context, clauses []gs.ClauseSet) error {
	return nil
}
