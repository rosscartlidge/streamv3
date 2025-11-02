package main

import (
	"fmt"
	"iter"
	"os"
	"regexp"
	"strconv"
	"strings"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// Helper functions for command handlers

func extractNumeric(val any) float64 {
	switch v := val.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	case string:
		// For strings, use 0 (they'll maintain relative order)
		return 0
	default:
		return 0
	}
}

// applyOperator applies a comparison operator for where command
func applyOperator(fieldValue any, op string, compareValue string) bool {
	switch op {
	case "eq":
		return compareEqual(fieldValue, compareValue)
	case "ne":
		return !compareEqual(fieldValue, compareValue)
	case "gt":
		return compareGreater(fieldValue, compareValue)
	case "ge":
		return compareGreater(fieldValue, compareValue) || compareEqual(fieldValue, compareValue)
	case "lt":
		return compareLess(fieldValue, compareValue)
	case "le":
		return compareLess(fieldValue, compareValue) || compareEqual(fieldValue, compareValue)
	case "contains":
		return compareContains(fieldValue, compareValue)
	case "startswith":
		return compareStartsWith(fieldValue, compareValue)
	case "endswith":
		return compareEndsWith(fieldValue, compareValue)
	case "pattern", "regexp", "regex":
		return comparePattern(fieldValue, compareValue)
	default:
		return false
	}
}

func compareEqual(fieldValue any, compareValue string) bool {
	switch v := fieldValue.(type) {
	case string:
		return v == compareValue
	case int64:
		if num, err := strconv.ParseInt(compareValue, 10, 64); err == nil {
			return v == num
		}
	case float64:
		if num, err := strconv.ParseFloat(compareValue, 64); err == nil {
			return v == num
		}
	case bool:
		if b, err := strconv.ParseBool(compareValue); err == nil {
			return v == b
		}
	}
	return fmt.Sprintf("%v", fieldValue) == compareValue
}

func compareGreater(fieldValue any, compareValue string) bool {
	switch v := fieldValue.(type) {
	case int64:
		if num, err := strconv.ParseInt(compareValue, 10, 64); err == nil {
			return v > num
		}
	case float64:
		if num, err := strconv.ParseFloat(compareValue, 64); err == nil {
			return v > num
		}
	case string:
		return v > compareValue
	}
	return false
}

func compareLess(fieldValue any, compareValue string) bool {
	switch v := fieldValue.(type) {
	case int64:
		if num, err := strconv.ParseInt(compareValue, 10, 64); err == nil {
			return v < num
		}
	case float64:
		if num, err := strconv.ParseFloat(compareValue, 64); err == nil {
			return v < num
		}
	case string:
		return v < compareValue
	}
	return false
}

func compareContains(fieldValue any, compareValue string) bool {
	if str, ok := fieldValue.(string); ok {
		return contains(str, compareValue)
	}
	return false
}

func compareStartsWith(fieldValue any, compareValue string) bool {
	if str, ok := fieldValue.(string); ok {
		return len(str) >= len(compareValue) && str[:len(compareValue)] == compareValue
	}
	return false
}

func compareEndsWith(fieldValue any, compareValue string) bool {
	if str, ok := fieldValue.(string); ok {
		return len(str) >= len(compareValue) && str[len(str)-len(compareValue):] == compareValue
	}
	return false
}

func comparePattern(fieldValue any, pattern string) bool {
	if str, ok := fieldValue.(string); ok {
		matched, err := regexp.MatchString(pattern, str)
		if err != nil {
			return false
		}
		return matched
	}
	return false
}

func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// buildAggregator builds a StreamV3 AggregateFunc from a spec (for group-by command)
func buildAggregator(function, field string) (streamv3.AggregateFunc, error) {
	switch function {
	case "count":
		return streamv3.Count(), nil
	case "sum":
		return streamv3.Sum(field), nil
	case "avg":
		return streamv3.Avg(field), nil
	case "min":
		return streamv3.Min[float64](field), nil
	case "max":
		return streamv3.Max[float64](field), nil
	default:
		return nil, fmt.Errorf("unknown aggregation function: %s", function)
	}
}

// unionRecordToKey converts a record to a string key for deduplication (for union command)
func unionRecordToKey(r streamv3.Record) string {
	// Use JSON representation as unique key
	return fmt.Sprintf("%v", r)
}

// chainRecords chains multiple data sources into a single stream (for union command)
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
					// Skip file on error
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

// shouldGenerate checks if code generation is enabled via flag or environment variable
// Returns true if:
//   - The generate flag is explicitly set to true, OR
//   - The STREAMV3_GENERATE_GO environment variable is set to "1" or "true"
func shouldGenerate(flagValue bool) bool {
	if flagValue {
		return true
	}
	envValue := os.Getenv("STREAMV3_GENERATE_GO")
	return envValue == "1" || envValue == "true"
}

// getCommandString returns the command line that invoked this command
// Filters out the -generate flag since it's implied by the code generation context
// Returns something like "streamv3 read-csv data.csv" or "streamv3 where -match age gt 18"
// Properly quotes arguments that contain shell special characters
func getCommandString() string {
	// Filter out -generate and -g flags
	var args []string
	skipNext := false
	for i, arg := range os.Args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "-generate" || arg == "-g" {
			continue
		}
		// For the binary name, use just "streamv3" instead of full path
		if i == 0 {
			args = append(args, "streamv3")
		} else {
			// Quote the argument if it needs quoting for shell safety
			args = append(args, shellQuote(arg))
		}
	}
	return strings.Join(args, " ")
}

// shellQuote quotes a string for safe use in shell commands
// Returns the string with appropriate quoting if needed
func shellQuote(s string) string {
	// If the string is simple (alphanumeric, dash, underscore, dot, slash, colon), no quoting needed
	needsQuoting := false
	for _, c := range s {
		if !isSimpleShellChar(c) {
			needsQuoting = true
			break
		}
	}

	if !needsQuoting {
		return s
	}

	// If string contains single quotes, use double quotes and escape special chars
	if strings.Contains(s, "'") {
		// Use double quotes, escape $, `, \, ", and !
		escaped := strings.ReplaceAll(s, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		escaped = strings.ReplaceAll(escaped, `$`, `\$`)
		escaped = strings.ReplaceAll(escaped, "`", "\\`")
		return `"` + escaped + `"`
	}

	// Otherwise use single quotes (most literal, safest)
	return "'" + s + "'"
}

// isSimpleShellChar returns true if the character is safe in shell without quoting
func isSimpleShellChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_' || c == '.' || c == '/' || c == ':'
}

// generateReadCSVCode generates Go code for the read-csv command
func generateReadCSVCode(filename string) error {
	// Generate ReadCSV call with error handling
	var code string
	var imports []string

	if filename == "" {
		// Reading from stdin - use ReadCSVFromReader
		code = `records := streamv3.ReadCSVFromReader(os.Stdin)`
		imports = []string{"os"}
	} else {
		// Reading from file - use ReadCSV with error handling
		code = fmt.Sprintf(`records, err := streamv3.ReadCSV(%q)
	if err != nil {
		return fmt.Errorf("reading CSV: %%w", err)
	}`, filename)
		imports = []string{"fmt"}
	}

	// Create init fragment (first in pipeline)
	frag := lib.NewInitFragment("records", code, imports, getCommandString())

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}

// generateWhereCode generates Go code for the where command
func generateWhereCode(ctx *cf.Context, inputFile string) error {
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

	// Get input variable name from last fragment
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}

	// Generate filter code from clauses
	filterCode, imports := generateWhereCodeFromClauses(ctx.Clauses)

	// Build complete statement
	outputVar := "filtered"
	code := fmt.Sprintf("%s := streamv3.Where(%s)(%s)", outputVar, filterCode, inputVar)

	// Create code fragment
	frag := lib.NewStmtFragment(outputVar, inputVar, code, imports, getCommandString())

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}

// generateWhereCodeFromClauses generates the filter function code
func generateWhereCodeFromClauses(clauses []cf.Clause) (string, []string) {
	var imports []string
	var clauseConditions []string

	// Build conditions for each clause (OR logic between clauses)
	for _, clause := range clauses {
		matchesRaw, ok := clause.Flags["-match"]
		if !ok || matchesRaw == nil {
			continue
		}

		matches, ok := matchesRaw.([]any)
		if !ok || len(matches) == 0 {
			continue
		}

		// AND logic within clause: all matches must pass
		var andConditions []string
		for _, matchRaw := range matches {
			matchMap, ok := matchRaw.(map[string]any)
			if !ok {
				continue
			}

			field, _ := matchMap["field"].(string)
			op, _ := matchMap["operator"].(string)
			value, _ := matchMap["value"].(string)

			if field == "" || op == "" {
				continue
			}

			// Generate condition code
			cond, imp := generateCondition(field, op, value)
			andConditions = append(andConditions, cond)
			imports = append(imports, imp...)
		}

		// Combine AND conditions for this clause
		if len(andConditions) > 0 {
			if len(andConditions) == 1 {
				clauseConditions = append(clauseConditions, andConditions[0])
			} else {
				clauseConditions = append(clauseConditions, "("+strings.Join(andConditions, " && ")+")")
			}
		}
	}

	// Combine clauses with OR
	var finalCondition string
	if len(clauseConditions) == 0 {
		finalCondition = "return true"
	} else if len(clauseConditions) == 1 {
		finalCondition = "return " + clauseConditions[0]
	} else {
		finalCondition = "return " + strings.Join(clauseConditions, " || ")
	}

	// Build function
	code := fmt.Sprintf("func(r streamv3.Record) bool {\n\t\t%s\n\t}", finalCondition)

	return code, dedupeImports(imports)
}

// generateCondition generates code for a single where condition
func generateCondition(field, op, value string) (string, []string) {
	var imports []string

	// Detect if value is numeric
	_, err := strconv.ParseFloat(value, 64)
	isNum := err == nil

	switch op {
	case "eq":
		if isNum {
			return fmt.Sprintf("streamv3.GetOr(r, %q, float64(0)) == %s", field, value), nil
		}
		return fmt.Sprintf("streamv3.GetOr(r, %q, \"\") == %q", field, value), nil

	case "ne":
		if isNum {
			return fmt.Sprintf("streamv3.GetOr(r, %q, float64(0)) != %s", field, value), nil
		}
		return fmt.Sprintf("streamv3.GetOr(r, %q, \"\") != %q", field, value), nil

	case "gt":
		return fmt.Sprintf("streamv3.GetOr(r, %q, float64(0)) > %s", field, value), nil

	case "ge":
		return fmt.Sprintf("streamv3.GetOr(r, %q, float64(0)) >= %s", field, value), nil

	case "lt":
		return fmt.Sprintf("streamv3.GetOr(r, %q, float64(0)) < %s", field, value), nil

	case "le":
		return fmt.Sprintf("streamv3.GetOr(r, %q, float64(0)) <= %s", field, value), nil

	case "contains":
		imports = append(imports, "strings")
		return fmt.Sprintf("strings.Contains(streamv3.GetOr(r, %q, \"\"), %q)", field, value), imports

	case "startswith":
		imports = append(imports, "strings")
		return fmt.Sprintf("strings.HasPrefix(streamv3.GetOr(r, %q, \"\"), %q)", field, value), imports

	case "endswith":
		imports = append(imports, "strings")
		return fmt.Sprintf("strings.HasSuffix(streamv3.GetOr(r, %q, \"\"), %q)", field, value), imports

	case "pattern", "regexp", "regex":
		imports = append(imports, "regexp")
		return fmt.Sprintf("regexp.MustCompile(%q).MatchString(streamv3.GetOr(r, %q, \"\"))", value, field), imports

	default:
		return "false", nil
	}
}

// dedupeImports removes duplicate imports
func dedupeImports(imports []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, imp := range imports {
		if imp != "" && !seen[imp] {
			seen[imp] = true
			result = append(result, imp)
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

	// Generate WriteCSV call
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

// generateLimitCode generates Go code for the limit command
func generateLimitCode(n int) error {
	fragments, err := lib.ReadAllCodeFragments()
	if err != nil {
		return fmt.Errorf("reading code fragments: %w", err)
	}
	for _, frag := range fragments {
		if err := lib.WriteCodeFragment(frag); err != nil {
			return fmt.Errorf("writing previous fragment: %w", err)
		}
	}
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}
	outputVar := "limited"
	code := fmt.Sprintf("%s := streamv3.Limit[streamv3.Record](%d)(%s)", outputVar, n, inputVar)
	frag := lib.NewStmtFragment(outputVar, inputVar, code, nil, getCommandString())
	return lib.WriteCodeFragment(frag)
}

// generateOffsetCode generates Go code for the offset command  
func generateOffsetCode(n int) error {
	fragments, err := lib.ReadAllCodeFragments()
	if err != nil {
		return fmt.Errorf("reading code fragments: %w", err)
	}
	for _, frag := range fragments {
		if err := lib.WriteCodeFragment(frag); err != nil {
			return fmt.Errorf("writing previous fragment: %w", err)
		}
	}
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}
	outputVar := "skipped"
	code := fmt.Sprintf("%s := streamv3.Offset[streamv3.Record](%d)(%s)", outputVar, n, inputVar)
	frag := lib.NewStmtFragment(outputVar, inputVar, code, nil, getCommandString())
	return lib.WriteCodeFragment(frag)
}

// generateDistinctCode generates Go code for the distinct command
func generateDistinctCode() error {
	fragments, err := lib.ReadAllCodeFragments()
	if err != nil {
		return fmt.Errorf("reading code fragments: %w", err)
	}
	for _, frag := range fragments {
		if err := lib.WriteCodeFragment(frag); err != nil {
			return fmt.Errorf("writing previous fragment: %w", err)
		}
	}
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}
	outputVar := "distinct"
	code := fmt.Sprintf(`%s := streamv3.DistinctBy(func(r streamv3.Record) string {
		return fmt.Sprintf("%%v", r)
	})(%s)`, outputVar, inputVar)
	frag := lib.NewStmtFragment(outputVar, inputVar, code, []string{"fmt"}, getCommandString())
	return lib.WriteCodeFragment(frag)
}

// generateSortCode generates Go code for the sort command
func generateSortCode(field string, desc bool) error {
	fragments, err := lib.ReadAllCodeFragments()
	if err != nil {
		return fmt.Errorf("reading code fragments: %w", err)
	}
	for _, frag := range fragments {
		if err := lib.WriteCodeFragment(frag); err != nil {
			return fmt.Errorf("writing previous fragment: %w", err)
		}
	}
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}
	outputVar := "sorted"
	var sortFunc string
	if desc {
		sortFunc = fmt.Sprintf(`streamv3.SortBy(func(r streamv3.Record) float64 {
		val, _ := streamv3.Get[any](r, %q)
		return -extractNumeric(val)
	})`, field)
	} else {
		sortFunc = fmt.Sprintf(`streamv3.SortBy(func(r streamv3.Record) float64 {
		val, _ := streamv3.Get[any](r, %q)
		return extractNumeric(val)
	})`, field)
	}
	code := fmt.Sprintf("%s := %s(%s)", outputVar, sortFunc, inputVar)
	frag := lib.NewStmtFragment(outputVar, inputVar, code, nil, getCommandString())
	return lib.WriteCodeFragment(frag)
}
