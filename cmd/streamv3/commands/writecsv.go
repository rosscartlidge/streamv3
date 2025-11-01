package commands

import (
	"bufio"
	"context"
	"fmt"
	"sort"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// writeCSVCommand implements the write-csv command
type writeCSVCommand struct {
	cmd *cf.Command
}

func init() {
	RegisterCommand(newWriteCSVCommand())
}

func newWriteCSVCommand() *writeCSVCommand {
	var outputFile string
	var generate bool

	cmd := cf.NewCommand("write-csv").
		Description("Read JSONL stream and write as CSV file").
		Flag("-generate", "-g").
			Bool().
			Bind(&generate).
			Global().
			Help("Generate Go code instead of executing").
			Done().
		Flag("FILE").
			String().
			Completer(&cf.FileCompleter{Pattern: "*.csv"}).
			Bind(&outputFile).
			Global().
			Default("").
			Help("Output CSV file (or stdout if not specified)").
			Done().
		Handler(func(ctx *cf.Context) error {
			// If -generate flag is set, generate Go code instead of executing
			if shouldGenerate(generate) {
				return generateWriteCSVCode(outputFile)
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

			// Wrap in buffered writer for performance
			writer := bufio.NewWriter(output)
			defer writer.Flush()

			// Write header
			fmt.Fprintf(writer, "%s\n", formatCSVRow(fieldOrder))

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
				fmt.Fprintf(writer, "%s\n", formatCSVRow(values))
			}

			return nil
		}).
		Build()

	return &writeCSVCommand{cmd: cmd}
}

func (c *writeCSVCommand) Name() string {
	return "write-csv"
}

func (c *writeCSVCommand) Description() string {
	return "Read JSONL stream and write as CSV file"
}

func (c *writeCSVCommand) GetCFCommand() *cf.Command {
	return c.cmd
}


func (c *writeCSVCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before completionflags framework takes over
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

	return c.cmd.Execute(args)
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

// generateWriteCSVCode generates Go code for the write-csv command
func generateWriteCSVCode(filename string) error {
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

	// Generate WriteCSV call (signature: WriteCSV(stream, filename))
	// Use WriteCSVToWriter for stdout, WriteCSV for files
	var code string
	var imports []string
	if filename == "" {
		code = fmt.Sprintf(`streamv3.WriteCSVToWriter(%s, os.Stdout)`, inputVar)
		imports = append(imports, "os")
	} else {
		code = fmt.Sprintf(`streamv3.WriteCSV(%s, %q)`, inputVar, filename)
	}

	// Create final fragment (no output variable)
	frag := lib.NewFinalFragment(inputVar, code, imports, getCommandString())

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}
