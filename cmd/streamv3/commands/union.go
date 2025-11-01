package commands

import (
	"context"
	"fmt"
	"iter"
	"os"
	"strings"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// unionCommand implements the union command
type unionCommand struct {
	cmd *cf.Command
}

func init() {
	RegisterCommand(newUnionCommand())
}

func newUnionCommand() *unionCommand {
	var inputFile string
	var unionAll bool
	var generate bool

	cmd := cf.NewCommand("union").
		Description("Combine records from multiple sources (SQL UNION)").
		Flag("-file", "-f").
			String().
			Completer(&cf.FileCompleter{Pattern: "*.{csv,jsonl}"}).
			Accumulate().
			Local().
			Help("Additional file to union (CSV or JSONL)").
			Done().
		Flag("-all", "-a").
			Bool().
			Bind(&unionAll).
			Global().
			Help("Keep duplicates (UNION ALL instead of UNION)").
			Done().
		Flag("-generate", "-g").
			Bool().
			Bind(&generate).
			Global().
			Help("Generate Go code instead of executing").
			Done().
		Flag("-input", "-i").
			String().
			Completer(&cf.FileCompleter{Pattern: "*.jsonl"}).
			Bind(&inputFile).
			Global().
			Default("").
			Help("First input JSONL file (or stdin if not specified)").
			Done().
		Handler(func(ctx *cf.Context) error {
			// If -generate flag is set, generate Go code instead of executing
			if shouldGenerate(generate) {
				return generateUnionCode(ctx, unionAll, inputFile)
			}

			// Get additional files from -file flags
			var additionalFiles []string
			if len(ctx.Clauses) > 0 {
				clause := ctx.Clauses[0]
				if filesRaw, ok := clause.Flags["-file"]; ok {
					if filesSlice, ok := filesRaw.([]any); ok {
						for _, v := range filesSlice {
							if file, ok := v.(string); ok && file != "" {
								additionalFiles = append(additionalFiles, file)
							}
						}
					}
				}
			}

			if len(additionalFiles) == 0 {
				return fmt.Errorf("at least one file required for union (use -file)")
			}

			// Read first input (stdin or file)
			firstInput, err := lib.OpenInput(inputFile)
			if err != nil {
				return fmt.Errorf("opening first input: %w", err)
			}
			defer firstInput.Close()

			firstRecords := lib.ReadJSONL(firstInput)

			// Chain all iterators together
			combined := chainRecords(firstRecords, additionalFiles)

			// Apply distinct if not UNION ALL
			var result iter.Seq[streamv3.Record]
			if unionAll {
				result = combined
			} else {
				// Apply distinct using DistinctBy with full record key
				distinct := streamv3.DistinctBy(func(r streamv3.Record) string {
					return unionRecordToKey(r)
				})
				result = distinct(combined)
			}

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, result); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Build()

	return &unionCommand{cmd: cmd}
}

func (c *unionCommand) Name() string {
	return "union"
}

func (c *unionCommand) Description() string {
	return "Combine records from multiple sources (SQL UNION)"
}

func (c *unionCommand) GetCFCommand() *cf.Command {
	return c.cmd
}

func (c *unionCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before completionflags framework takes over
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("union - Combine records from multiple sources (SQL UNION)")
		fmt.Println()
		fmt.Println("Usage: streamv3 union -file <file1> [-file <file2>]... [-all]")
		fmt.Println()
		fmt.Println("Combines records from multiple data sources. By default removes")
		fmt.Println("duplicates (UNION). Use -all to keep duplicates (UNION ALL).")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  -file <file>  Additional file to union (can repeat)")
		fmt.Println("  -all          Keep duplicates (UNION ALL)")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # UNION (remove duplicates)")
		fmt.Println("  streamv3 read-csv file1.csv | \\")
		fmt.Println("    streamv3 union -file file2.csv -file file3.csv")
		fmt.Println()
		fmt.Println("  # UNION ALL (keep duplicates)")
		fmt.Println("  streamv3 read-csv customers.csv | \\")
		fmt.Println("    streamv3 union -all -file suppliers.csv")
		fmt.Println()
		fmt.Println("  # SQL equivalent:")
		fmt.Println("  # SELECT * FROM customers")
		fmt.Println("  # UNION")
		fmt.Println("  # SELECT * FROM suppliers")
		fmt.Println("  streamv3 read-csv customers.csv | \\")
		fmt.Println("    streamv3 union -file suppliers.csv")
		fmt.Println()
		fmt.Println("Note: All inputs should have compatible schemas (same fields)")
		return nil
	}

	return c.cmd.Execute(args)
}

// chainRecords chains multiple data sources into a single stream
func chainRecords(firstRecords iter.Seq[streamv3.Record], additionalFiles []string) iter.Seq[streamv3.Record] {
	return func(yield func(streamv3.Record) bool) {
		// Yield from first stream
		for record := range firstRecords {
			if !yield(record) {
				return
			}
		}

		// Yield from each additional file
		for _, file := range additionalFiles {
			var records iter.Seq[streamv3.Record]

			if strings.HasSuffix(file, ".csv") {
				// Read CSV
				csvRecords, err := streamv3.ReadCSV(file)
				if err != nil {
					// Skip file on error (or could yield error)
					continue
				}
				records = csvRecords
			} else {
				// Read JSONL
				f, err := os.Open(file)
				if err != nil {
					continue
				}
				records = lib.ReadJSONL(f)
				defer f.Close()
			}

			// Yield from this file
			for record := range records {
				if !yield(record) {
					return
				}
			}
		}
	}
}

// unionRecordToKey converts a record to a string key for deduplication
func unionRecordToKey(r streamv3.Record) string {
	// Create a stable string representation of the record
	// Sort keys to ensure consistency and exclude _row_number
	var keys []string
	for k := range r.KeysIter() {
		if k != "_row_number" {
			keys = append(keys, k)
		}
	}

	// Sort keys for deterministic output
	sortUnionStrings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, streamv3.GetOr(r, k, "")))
	}
	return strings.Join(parts, "|")
}

// sortUnionStrings sorts a slice of strings in place (simple bubble sort for small slices)
func sortUnionStrings(s []string) {
	n := len(s)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if s[j] > s[j+1] {
				s[j], s[j+1] = s[j+1], s[j]
			}
		}
	}
}

// generateUnionCode generates Go code for the union command
func generateUnionCode(ctx *cf.Context, unionAll bool, inputFile string) error {
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

	// Get additional files
	var additionalFiles []string
	if len(ctx.Clauses) > 0 {
		clause := ctx.Clauses[0]
		if filesRaw, ok := clause.Flags["-file"]; ok {
			if filesSlice, ok := filesRaw.([]any); ok {
				for _, v := range filesSlice {
					if file, ok := v.(string); ok && file != "" {
						additionalFiles = append(additionalFiles, file)
					}
				}
			}
		}
	}

	// Generate code to read additional files
	var readFilesCode []string
	for i, file := range additionalFiles {
		varName := fmt.Sprintf("file%d", i+1)
		if strings.HasSuffix(file, ".csv") {
			readFilesCode = append(readFilesCode, fmt.Sprintf(`%s, err := streamv3.ReadCSV(%q)
	if err != nil {
		return fmt.Errorf("reading CSV: %%w", err)
	}`, varName, file))
		} else {
			readFilesCode = append(readFilesCode, fmt.Sprintf(`%sFile, err := os.Open(%q)
	if err != nil {
		return fmt.Errorf("opening file: %%w", err)
	}
	defer %sFile.Close()
	%s := /* read JSONL */`, varName, file, varName, varName))
		}
	}

	// Generate chain code
	var chainCode strings.Builder
	chainCode.WriteString("combined := func(yield func(streamv3.Record) bool) {\n")
	chainCode.WriteString(fmt.Sprintf("\t\tfor record := range %s {\n", inputVar))
	chainCode.WriteString("\t\t\tif !yield(record) { return }\n")
	chainCode.WriteString("\t\t}\n")

	for i := range additionalFiles {
		varName := fmt.Sprintf("file%d", i+1)
		chainCode.WriteString(fmt.Sprintf("\t\tfor record := range %s {\n", varName))
		chainCode.WriteString("\t\t\tif !yield(record) { return }\n")
		chainCode.WriteString("\t\t}\n")
	}
	chainCode.WriteString("\t}")

	// Generate distinct code if needed
	var finalCode string
	if unionAll {
		finalCode = fmt.Sprintf(`%s
	%s
	result := combined`, strings.Join(readFilesCode, "\n\t"), chainCode.String())
	} else {
		finalCode = fmt.Sprintf(`%s
	%s
	result := streamv3.DistinctBy(func(r streamv3.Record) string {
		var parts []string
		for k, v := range r.All() {
			parts = append(parts, fmt.Sprintf("%%s=%%v", k, v))
		}
		return strings.Join(parts, "|")
	})(combined)`, strings.Join(readFilesCode, "\n\t"), chainCode.String())
	}

	// Create code fragment
	imports := []string{"fmt", "os", "strings"}
	frag := lib.NewStmtFragment("result", inputVar, finalCode, imports, getCommandString())

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}
