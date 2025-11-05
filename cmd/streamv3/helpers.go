package main

import (
	"fmt"
	"iter"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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

// generateGroupByCode generates Go code for the group-by command
func generateGroupByCode(ctx *cf.Context, groupByFields []string) error {
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

	// Validate group-by fields
	if len(groupByFields) == 0 {
		return fmt.Errorf("no group-by field specified (use -by)")
	}

	// Parse aggregation specifications from clauses
	type aggSpec struct {
		function string
		field    string
		result   string
	}

	var aggSpecs []aggSpec
	for _, clause := range ctx.Clauses {
		function, _ := clause.Flags["-function"].(string)
		field, _ := clause.Flags["-field"].(string)
		result, _ := clause.Flags["-result"].(string)

		// Skip empty clauses
		if function == "" && result == "" {
			continue
		}

		if function == "" {
			return fmt.Errorf("aggregation missing -function")
		}

		if result == "" {
			return fmt.Errorf("aggregation missing -result")
		}

		// Validate function
		switch function {
		case "count", "sum", "avg", "min", "max":
			// Valid
		default:
			return fmt.Errorf("unknown aggregation function: %s", function)
		}

		// For non-count functions, field is required
		if function != "count" && field == "" {
			return fmt.Errorf("aggregation function %s requires -field", function)
		}

		aggSpecs = append(aggSpecs, aggSpec{
			function: function,
			field:    field,
			result:   result,
		})
	}

	if len(aggSpecs) == 0 {
		return fmt.Errorf("no aggregations specified (use -function and -result)")
	}

	// Generate TWO fragments: one for GroupByFields, one for Aggregate
	// This allows each to be extracted cleanly for Chain()

	// Fragment 1: GroupByFields
	groupCode := "grouped := streamv3.GroupByFields(\"_group\""
	for _, field := range groupByFields {
		groupCode += fmt.Sprintf(", %q", field)
	}
	groupCode += fmt.Sprintf(")(%s)", inputVar)

	frag1 := lib.NewStmtFragment("grouped", inputVar, groupCode, nil, getCommandString())
	if err := lib.WriteCodeFragment(frag1); err != nil {
		return fmt.Errorf("writing GroupByFields fragment: %w", err)
	}

	// Fragment 2: Aggregate
	// Note: Empty command string since this is part of the same CLI command as Fragment 1
	aggCode := "aggregated := streamv3.Aggregate(\"_group\", map[string]streamv3.AggregateFunc{\n"
	for i, spec := range aggSpecs {
		if i > 0 {
			aggCode += ",\n"
		}
		aggCode += fmt.Sprintf("\t\t%q: %s", spec.result, generateAggregatorCode(spec))
	}
	aggCode += ",\n\t})(grouped)"

	frag2 := lib.NewStmtFragment("aggregated", "grouped", aggCode, nil, "")
	return lib.WriteCodeFragment(frag2)
}

// generateAggregatorCode generates code for a single aggregator
func generateAggregatorCode(spec struct {
	function string
	field    string
	result   string
}) string {
	switch spec.function {
	case "count":
		return "streamv3.Count()"
	case "sum":
		return fmt.Sprintf("streamv3.Sum(%q)", spec.field)
	case "avg":
		return fmt.Sprintf("streamv3.Avg(%q)", spec.field)
	case "min":
		return fmt.Sprintf("streamv3.Min[float64](%q)", spec.field)
	case "max":
		return fmt.Sprintf("streamv3.Max[float64](%q)", spec.field)
	default:
		return ""
	}
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

// generateChartCode generates Go code for the chart command
func generateChartCode(xField, yField, outputFile string) error {
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
	code := fmt.Sprintf(`if err := streamv3.QuickChart(%s, %q, %q, %q); err != nil {
		return fmt.Errorf("creating chart: %%w", err)
	}
	fmt.Printf("Chart created: %%s\n", %q)`, inputVar, xField, yField, outputFile, outputFile)

	// Create final fragment (chart is a terminal operation with side effects)
	frag := lib.NewFinalFragment(inputVar, code, []string{"fmt"}, getCommandString())
	return lib.WriteCodeFragment(frag)
}

// parseValue attempts to parse a string value into the appropriate Go type
// Handles int64, float64, bool, time.Time, and defaults to string
func parseValue(s string) any {
	// Try bool
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	// Try int64
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}

	// Try float64
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	// Try time.Time (common formats)
	timeFormats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
	}
	for _, format := range timeFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}

	// Default to string
	return s
}

// generateUpdateCode generates Go code for the update command with conditional clauses
func generateUpdateCode(ctx *cf.Context, inputFile string) error {
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

	// Get input variable from last fragment (or default to "records")
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}

	// Parse clauses - each clause has optional -match conditions and required -set operations
	type updateClause struct {
		conditions []struct {
			field string
			op    string
			value string
		}
		updates []struct {
			field string
			value string
		}
	}

	var clauses []updateClause

	for _, clause := range ctx.Clauses {
		uc := updateClause{}

		// Parse -match conditions (optional)
		if matchesRaw, ok := clause.Flags["-match"]; ok && matchesRaw != nil {
			matches, ok := matchesRaw.([]any)
			if ok {
				for _, matchRaw := range matches {
					matchMap, ok := matchRaw.(map[string]any)
					if !ok {
						continue
					}

					field, _ := matchMap["field"].(string)
					op, _ := matchMap["operator"].(string)
					value, _ := matchMap["value"].(string)

					if field != "" && op != "" {
						uc.conditions = append(uc.conditions, struct {
							field string
							op    string
							value string
						}{field, op, value})
					}
				}
			}
		}

		// Parse -set operations (required)
		if setOpsRaw, ok := clause.Flags["-set"]; ok && setOpsRaw != nil {
			setList, ok := setOpsRaw.([]any)
			if ok {
				for _, setRaw := range setList {
					setMap, ok := setRaw.(map[string]any)
					if !ok {
						continue
					}

					field, _ := setMap["field"].(string)
					value, _ := setMap["value"].(string)

					if field != "" {
						uc.updates = append(uc.updates, struct {
							field string
							value string
						}{field, value})
					}
				}
			}
		}

		if len(uc.updates) > 0 {
			clauses = append(clauses, uc)
		}
	}

	if len(clauses) == 0 {
		return fmt.Errorf("no -set operations specified")
	}

	// Generate Update code with conditional clauses
	var codeBody strings.Builder
	needsTime := false
	needsStrings := false
	needsRegexp := false

	codeBody.WriteString("\t\tfrozen := mut.Freeze()\n\n")

	// Generate clause evaluation (first-match-wins)
	for i, clause := range clauses {
		indent := "\t\t"

		// Generate condition check for this clause
		if len(clause.conditions) > 0 {
			if i == 0 {
				codeBody.WriteString(indent + "if ")
			} else {
				codeBody.WriteString(indent + "} else if ")
			}

			// Generate all conditions with AND logic
			for j, cond := range clause.conditions {
				if j > 0 {
					codeBody.WriteString(" && ")
				}
				codeBody.WriteString(generateConditionCode(cond.field, cond.op, cond.value))

				// Track which imports are needed
				switch cond.op {
				case "contains", "startswith", "endswith":
					needsStrings = true
				case "pattern", "regexp", "regex":
					needsRegexp = true
				}
			}
			codeBody.WriteString(" {\n")
			indent = "\t\t\t"
		} else if i > 0 {
			// Default case (no conditions) - use else
			codeBody.WriteString("\t\t} else {\n")
			indent = "\t\t\t"
		}

		// Generate update statements
		for _, upd := range clause.updates {
			parsedValue := parseValue(upd.value)

			var stmt string
			switch v := parsedValue.(type) {
			case int64:
				stmt = fmt.Sprintf("%smut = mut.Int(%q, int64(%d))", indent, upd.field, v)
			case float64:
				stmt = fmt.Sprintf("%smut = mut.Float(%q, %f)", indent, upd.field, v)
			case bool:
				stmt = fmt.Sprintf("%smut = mut.Bool(%q, %t)", indent, upd.field, v)
			case time.Time:
				timeExpr := fmt.Sprintf("time.Date(%d, %d, %d, %d, %d, %d, %d, time.UTC)",
					v.Year(), v.Month(), v.Day(), v.Hour(), v.Minute(), v.Second(), v.Nanosecond())
				stmt = fmt.Sprintf("%smut = streamv3.Set(mut, %q, %s)", indent, upd.field, timeExpr)
				needsTime = true
			case string:
				stmt = fmt.Sprintf("%smut = mut.String(%q, %q)", indent, upd.field, v)
			default:
				stmt = fmt.Sprintf("%smut = mut.String(%q, %q)", indent, upd.field, upd.value)
			}

			codeBody.WriteString(stmt + "\n")
		}
	}

	// Close the if-else chain if we had conditions
	if len(clauses) > 0 && (len(clauses[0].conditions) > 0 || len(clauses) > 1) {
		codeBody.WriteString("\t\t}")
	}

	outputVar := "updated"
	code := fmt.Sprintf(`%s := streamv3.Update(func(mut streamv3.MutableRecord) streamv3.MutableRecord {
%s
		return mut
	})(%s)`, outputVar, codeBody.String(), inputVar)

	// Determine imports needed
	imports := []string{}
	if needsTime {
		imports = append(imports, "time")
	}
	if needsStrings {
		imports = append(imports, "strings")
	}
	if needsRegexp {
		imports = append(imports, "regexp")
	}

	// Create and write fragment
	frag := lib.NewStmtFragment(outputVar, inputVar, code, imports, getCommandString())
	return lib.WriteCodeFragment(frag)
}

// generateConditionCode generates the Go code for a single condition check
func generateConditionCode(field, op, value string) string {
	switch op {
	case "eq":
		return fmt.Sprintf("streamv3.GetOr(frozen, %q, %s) == %s",
			field, getDefaultValueForComparison(value), getComparisonValue(value))
	case "ne":
		return fmt.Sprintf("streamv3.GetOr(frozen, %q, %s) != %s",
			field, getDefaultValueForComparison(value), getComparisonValue(value))
	case "gt", "ge", "lt", "le":
		// Numeric comparisons
		return fmt.Sprintf("streamv3.GetOr(frozen, %q, float64(0)) %s %s",
			field, getOperatorCode(op), getComparisonValue(value))
	case "contains":
		return fmt.Sprintf("strings.Contains(streamv3.GetOr(frozen, %q, \"\"), %s)",
			field, getComparisonValue(value))
	case "startswith":
		return fmt.Sprintf("strings.HasPrefix(streamv3.GetOr(frozen, %q, \"\"), %s)",
			field, getComparisonValue(value))
	case "endswith":
		return fmt.Sprintf("strings.HasSuffix(streamv3.GetOr(frozen, %q, \"\"), %s)",
			field, getComparisonValue(value))
	case "pattern", "regexp", "regex":
		// For regexp, we need to compile the pattern
		return fmt.Sprintf("regexp.MustCompile(%s).MatchString(streamv3.GetOr(frozen, %q, \"\"))",
			getComparisonValue(value), field)
	default:
		// Fallback to equality
		return fmt.Sprintf("streamv3.GetOr(frozen, %q, %s) == %s",
			field, getDefaultValueForComparison(value), getComparisonValue(value))
	}
}

// getDefaultValueForComparison returns the default value for GetOr based on the comparison value's type
func getDefaultValueForComparison(value string) string {
	parsedValue := parseValue(value)
	switch parsedValue.(type) {
	case int64, float64:
		return "float64(0)"
	case bool:
		return "false"
	case time.Time:
		return "time.Time{}"
	default:
		return `""`
	}
}

// getOperatorCode converts operator string to Go comparison operator
func getOperatorCode(op string) string {
	switch op {
	case "eq":
		return "=="
	case "ne":
		return "!="
	case "gt":
		return ">"
	case "ge":
		return ">="
	case "lt":
		return "<"
	case "le":
		return "<="
	default:
		return "=="
	}
}

// getComparisonValue formats a value for comparison in generated code
func getComparisonValue(value string) string {
	parsedValue := parseValue(value)
	switch v := parsedValue.(type) {
	case int64:
		return fmt.Sprintf("float64(%d)", v)
	case float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case string:
		return fmt.Sprintf("%q", v)
	default:
		return fmt.Sprintf("%q", value)
	}
}

// generateTableCode generates Go code for the table command
func generateTableCode(maxWidth int) error {
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

	// Get input variable from last fragment
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}

	// Generate simple call to DisplayTable
	code := fmt.Sprintf("\tstreamv3.DisplayTable(%s, %d)\n", inputVar, maxWidth)

	// Create final fragment (table is a sink - no output variable)
	frag := lib.NewFinalFragment(inputVar, code, nil, getCommandString())
	return lib.WriteCodeFragment(frag)
}

// generateIncludeCode generates Go code for the include command
func generateIncludeCode(fields []string) error {
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

	// Get input variable from last fragment
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}

	// Generate field list
	var fieldsList strings.Builder
	fieldsList.WriteString("[]string{")
	for i, field := range fields {
		if i > 0 {
			fieldsList.WriteString(", ")
		}
		fieldsList.WriteString(fmt.Sprintf("%q", field))
	}
	fieldsList.WriteString("}")

	// Generate code
	outputVar := "included"
	code := fmt.Sprintf(`%s := streamv3.Select(func(r streamv3.Record) streamv3.Record {
		fields := %s
		mut := streamv3.MakeMutableRecord()
		for _, field := range fields {
			if val, ok := streamv3.Get[any](r, field); ok {
				mut = mut.SetAny(field, val)
			}
		}
		return mut.Freeze()
	})(%s)`, outputVar, fieldsList.String(), inputVar)

	// Create stmt fragment
	frag := lib.NewStmtFragment(outputVar, inputVar, code, nil, getCommandString())
	return lib.WriteCodeFragment(frag)
}

// generateExcludeCode generates Go code for the exclude command
func generateExcludeCode(fields []string) error {
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

	// Get input variable from last fragment
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}

	// Generate excluded map
	var excludedMap strings.Builder
	excludedMap.WriteString("map[string]bool{")
	for i, field := range fields {
		if i > 0 {
			excludedMap.WriteString(", ")
		}
		excludedMap.WriteString(fmt.Sprintf("%q: true", field))
	}
	excludedMap.WriteString("}")

	// Generate code
	outputVar := "excluded"
	code := fmt.Sprintf(`%s := streamv3.Select(func(r streamv3.Record) streamv3.Record {
		excluded := %s
		mut := streamv3.MakeMutableRecord()
		for k, v := range r.All() {
			if !excluded[k] {
				mut = mut.SetAny(k, v)
			}
		}
		return mut.Freeze()
	})(%s)`, outputVar, excludedMap.String(), inputVar)

	// Create stmt fragment
	frag := lib.NewStmtFragment(outputVar, inputVar, code, nil, getCommandString())
	return lib.WriteCodeFragment(frag)
}

// generateRenameCode generates Go code for the rename command
func generateRenameCode(renames []struct{ oldField, newField string }) error {
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

	// Get input variable from last fragment
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}

	// Generate rename map
	var renameMap strings.Builder
	renameMap.WriteString("map[string]string{")
	for i, r := range renames {
		if i > 0 {
			renameMap.WriteString(", ")
		}
		renameMap.WriteString(fmt.Sprintf("%q: %q", r.oldField, r.newField))
	}
	renameMap.WriteString("}")

	// Generate code
	outputVar := "renamed"
	code := fmt.Sprintf(`%s := streamv3.Select(func(r streamv3.Record) streamv3.Record {
		renames := %s
		mut := streamv3.MakeMutableRecord()
		for k, v := range r.All() {
			if newName, ok := renames[k]; ok {
				mut = mut.SetAny(newName, v)
			} else {
				mut = mut.SetAny(k, v)
			}
		}
		return mut.Freeze()
	})(%s)`, outputVar, renameMap.String(), inputVar)

	// Create stmt fragment
	frag := lib.NewStmtFragment(outputVar, inputVar, code, nil, getCommandString())
	return lib.WriteCodeFragment(frag)
}

// generateReadJSONCode generates Go code for the read-json command
func generateReadJSONCode(filename string) error {
	// No previous fragments for init command
	outputVar := "records"
	imports := []string{"fmt", "os"}

	code := fmt.Sprintf(`records, err := streamv3.ReadJSONAuto(%q)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %%v\n", fmt.Errorf("reading JSON: %%w", err))
		os.Exit(1)
	}`, filename)

	frag := lib.NewInitFragment(outputVar, code, imports, getCommandString())
	return lib.WriteCodeFragment(frag)
}

// generateWriteJSONCode generates Go code for the write-json command
func generateWriteJSONCode(filename string, pretty bool) error {
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

	// Get input variable from last fragment
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}

	var code string
	if filename == "" {
		// Write to stdout
		if pretty {
			code = fmt.Sprintf(`	// Collect and pretty-print records to stdout
	var recordMaps []map[string]interface{}
	for record := range %s {
		data := make(map[string]interface{})
		for k, v := range record.All() {
			data[k] = v
		}
		recordMaps = append(recordMaps, data)
	}
	jsonBytes, err := json.MarshalIndent(recordMaps, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %%v\n", err)
		os.Exit(1)
	}
	if _, err := os.Stdout.Write(jsonBytes); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing JSON: %%v\n", err)
		os.Exit(1)
	}
	os.Stdout.Write([]byte("\n"))`, inputVar)
		} else {
			code = fmt.Sprintf(`	if err := streamv3.WriteJSONToWriter(%s, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing JSON: %%v\n", err)
		os.Exit(1)
	}`, inputVar)
		}
	} else {
		// Write to file
		if pretty {
			code = fmt.Sprintf(`	if err := streamv3.WriteJSONPretty(%s, %q); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing JSON: %%v\n", err)
		os.Exit(1)
	}`, inputVar, filename)
		} else {
			code = fmt.Sprintf(`	if err := streamv3.WriteJSON(%s, %q); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing JSON: %%v\n", err)
		os.Exit(1)
	}`, inputVar, filename)
		}
	}

	imports := []string{"fmt", "os"}
	// Add encoding/json import if pretty printing to stdout
	if filename == "" && pretty {
		imports = append(imports, "encoding/json")
	}
	frag := lib.NewFinalFragment(inputVar, code, imports, getCommandString())
	return lib.WriteCodeFragment(frag)
}
