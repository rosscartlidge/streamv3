package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// CodeFragment represents a piece of generated Go code in a pipeline
type CodeFragment struct {
	Type    string   `json:"type"`    // "stmt" (statement), "final" (no output var), "init" (first in chain)
	Var     string   `json:"var"`     // Output variable name (e.g., "filtered0")
	Input   string   `json:"input"`   // Input variable name from previous command
	Code    string   `json:"code"`    // Go code for this operation
	Imports []string `json:"imports"` // Required imports (e.g., ["strings", "log"])
}

// ReadCodeFragment reads a code fragment from stdin
// Returns nil if stdin is empty (first command in pipeline)
func ReadCodeFragment() (*CodeFragment, error) {
	// Check if stdin has data
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("checking stdin: %w", err)
	}

	// If stdin is empty (no pipe), return nil (we're the first command)
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, nil
	}

	// Try to read a code fragment
	var frag CodeFragment
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&frag); err != nil {
		if err == io.EOF {
			return nil, nil // Empty stdin
		}
		return nil, fmt.Errorf("decoding code fragment: %w", err)
	}

	return &frag, nil
}

// ReadAllCodeFragments reads all code fragments from stdin
// Returns empty slice if stdin is empty (first command in pipeline)
func ReadAllCodeFragments() ([]*CodeFragment, error) {
	// Check if stdin has data
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("checking stdin: %w", err)
	}

	// If stdin is empty (no pipe), return empty slice (we're the first command)
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, nil
	}

	// Read all fragments
	var fragments []*CodeFragment
	decoder := json.NewDecoder(os.Stdin)

	for {
		var frag CodeFragment
		if err := decoder.Decode(&frag); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decoding fragment: %w", err)
		}
		fragments = append(fragments, &frag)
	}

	return fragments, nil
}

// WriteCodeFragment writes a code fragment to stdout as JSONL
func WriteCodeFragment(frag *CodeFragment) error {
	encoder := json.NewEncoder(os.Stdout)
	if err := encoder.Encode(frag); err != nil {
		return fmt.Errorf("encoding code fragment: %w", err)
	}
	return nil
}

// NewInitFragment creates the first fragment in a pipeline (e.g., read-csv)
func NewInitFragment(varName, code string, imports []string) *CodeFragment {
	return &CodeFragment{
		Type:    "init",
		Var:     varName,
		Input:   "",
		Code:    code,
		Imports: imports,
	}
}

// NewStmtFragment creates a statement fragment that transforms data
func NewStmtFragment(varName, inputVar, code string, imports []string) *CodeFragment {
	return &CodeFragment{
		Type:    "stmt",
		Var:     varName,
		Input:   inputVar,
		Code:    code,
		Imports: imports,
	}
}

// NewFinalFragment creates a final fragment with no output variable (e.g., write-csv)
func NewFinalFragment(inputVar, code string, imports []string) *CodeFragment {
	return &CodeFragment{
		Type:    "final",
		Var:     "",
		Input:   inputVar,
		Code:    code,
		Imports: imports,
	}
}

// GetInputVar returns the input variable name, or "records" if this is the first command
func (f *CodeFragment) GetInputVar() string {
	if f == nil || f.Input == "" {
		return "records"
	}
	return f.Input
}

// NextVarName generates the next variable name in sequence
// Pattern: records -> filtered0 -> filtered1 -> selected0 -> limited0 -> sorted0
func NextVarName(prefix string, input *CodeFragment) string {
	if input == nil {
		return "records"
	}

	// Count how many operations we've done
	// This is a simple approach - just use a counter suffix
	// For now, use the prefix with a 0 suffix
	return prefix + "0"
}

// AssembleCodeFragments reads all code fragments from stdin and assembles them into a complete Go program
// using streamv3.Chain() for better readability
func AssembleCodeFragments(input io.Reader) (string, error) {
	// Read all fragments from stdin
	var fragments []*CodeFragment
	decoder := json.NewDecoder(input)

	for {
		var frag CodeFragment
		if err := decoder.Decode(&frag); err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("decoding fragment: %w", err)
		}
		fragments = append(fragments, &frag)
	}

	if len(fragments) == 0 {
		return "", fmt.Errorf("no code fragments received")
	}

	// Separate init fragments (setup) from stmt fragments (transformations)
	var initFragments []*CodeFragment
	var stmtFragments []*CodeFragment
	var finalFragments []*CodeFragment

	for _, frag := range fragments {
		switch frag.Type {
		case "init":
			initFragments = append(initFragments, frag)
		case "stmt":
			stmtFragments = append(stmtFragments, frag)
		case "final":
			finalFragments = append(finalFragments, frag)
		}
	}

	// Collect all imports and deduplicate
	importSet := make(map[string]bool)
	importSet["github.com/rosscartlidge/streamv3"] = true // Always needed

	// If there are no final fragments, we'll auto-add JSONL output
	if len(finalFragments) == 0 {
		importSet["os"] = true
		importSet["fmt"] = true
	}

	for _, frag := range fragments {
		for _, imp := range frag.Imports {
			if imp != "" {
				importSet[imp] = true
			}
		}
	}

	// Build imports section
	var imports []string
	for imp := range importSet {
		imports = append(imports, imp)
	}

	// Sort imports for consistent output
	sortImports(imports)

	// Build the complete program
	var code string
	code += "package main\n\n"

	// Add imports
	if len(imports) > 0 {
		code += "import (\n"
		for _, imp := range imports {
			code += fmt.Sprintf("\t%q\n", imp)
		}
		code += ")\n\n"
	}

	// Add main function
	code += "func main() {\n"

	// Add init fragments (with proper error handling)
	for _, frag := range initFragments {
		code += "\t" + fixErrorHandling(frag.Code) + "\n"
	}

	// Build Chain() call if we have multiple stmt fragments
	if len(stmtFragments) > 1 {
		// Extract the input variable (from first stmt fragment or last init fragment)
		var inputVar string
		if len(initFragments) > 0 {
			inputVar = initFragments[len(initFragments)-1].Var
		} else {
			inputVar = "records"
		}

		// Extract the output variable (from last stmt fragment)
		outputVar := stmtFragments[len(stmtFragments)-1].Var

		// Build filters array
		code += "\t" + outputVar + " := streamv3.Chain(\n"
		for _, frag := range stmtFragments {
			// Extract just the filter function from the code
			// Pattern: "var := filter(input)" -> "filter"
			filterCode := extractFilter(frag.Code)
			code += "\t\t" + filterCode + ",\n"
		}
		code += "\t)(" + inputVar + ")\n"
	} else if len(stmtFragments) == 1 {
		// Single transformation - use directly
		code += "\t" + fixErrorHandling(stmtFragments[0].Code) + "\n"
	}

	// Add final fragments (e.g., write-csv)
	if len(finalFragments) > 0 {
		for _, frag := range finalFragments {
			code += "\t" + fixErrorHandling(frag.Code) + "\n"
		}
	} else {
		// No final fragment - auto-add JSONL output
		// Find the last output variable
		var outputVar string
		if len(stmtFragments) > 0 {
			outputVar = stmtFragments[len(stmtFragments)-1].Var
		} else if len(initFragments) > 0 {
			outputVar = initFragments[len(initFragments)-1].Var
		} else {
			outputVar = "records"
		}

		// Add JSONL output code using streamv3.WriteJSONToWriter
		code += "\t// Output records as JSONL\n"
		code += fmt.Sprintf("\tif err := streamv3.WriteJSONToWriter(%s, os.Stdout); err != nil {\n", outputVar)
		code += "\t\tfmt.Fprintf(os.Stderr, \"Error writing output: %v\\n\", err)\n"
		code += "\t\tos.Exit(1)\n"
		code += "\t}\n"
	}

	code += "}\n"

	return code, nil
}

// extractFilter extracts the filter function from a statement like "var := filter(input)"
// Returns just "filter" for use in Chain()
func extractFilter(code string) string {
	// Pattern: "outputVar := filterCall(inputVar)" or "outputVar := filterCall(...)(inputVar)"
	// We need to extract everything between ":=" and the final "(inputVar)"

	colonEqIdx := findString(code, ":=")
	if colonEqIdx == -1 {
		return code // Fallback: return as-is
	}

	// Start after ":= "
	start := colonEqIdx + 2
	for start < len(code) && (code[start] == ' ' || code[start] == '\t' || code[start] == '\n') {
		start++
	}

	// Find the last ")(" pattern which separates the filter from its application
	// E.g., "streamv3.Where(func...)(records)" - we want everything up to the last "("
	lastApplyIdx := findLastApplyParen(code)
	if lastApplyIdx == -1 {
		// No application found, might be already a filter
		return code[start:]
	}

	return code[start:lastApplyIdx]
}

// findLastApplyParen finds the last "(" that applies the filter to input
// Looks for ")(" pattern and returns the index of the second "("
func findLastApplyParen(code string) int {
	// Search backwards for ")(" pattern
	for i := len(code) - 1; i > 0; i-- {
		if code[i] == '(' && i > 0 && code[i-1] == ')' {
			return i
		}
	}
	return -1
}

// findString finds substring in string (simple helper)
func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// findLastParen finds the last opening parenthesis (for extracting filter from "filter(input)")
func findLastParen(s string, start int) int {
	depth := 0
	lastOpen := -1

	for i := start; i < len(s); i++ {
		if s[i] == '(' {
			if depth == 0 {
				lastOpen = i
			}
			depth++
		} else if s[i] == ')' {
			depth--
			if depth == 0 {
				return lastOpen
			}
		}
	}

	return lastOpen
}

// fixErrorHandling fixes "return fmt.Errorf(...)" to proper main() error handling
func fixErrorHandling(code string) string {
	// Replace "return fmt.Errorf(...)" with proper error handling for main()
	// Pattern: if err != nil { return fmt.Errorf("...", err) }

	// For now, use simple string replacement
	// TODO: Use more sophisticated parsing if needed

	replaced := code

	// Pattern 1: return fmt.Errorf("message: %w", err)
	if findString(replaced, "return fmt.Errorf(") != -1 {
		replaced = replaceReturnError(replaced)
	}

	return replaced
}

// replaceReturnError replaces "return fmt.Errorf(...)" with proper error handling
func replaceReturnError(code string) string {
	// Find "return fmt.Errorf(" and replace with proper error handling
	returnIdx := findString(code, "return fmt.Errorf(")
	if returnIdx == -1 {
		return code
	}

	// Find the full error message (up to the closing paren)
	msgStart := returnIdx + len("return fmt.Errorf(")
	depth := 1
	msgEnd := msgStart

	for msgEnd < len(code) && depth > 0 {
		if code[msgEnd] == '(' {
			depth++
		} else if code[msgEnd] == ')' {
			depth--
		}
		if depth > 0 {
			msgEnd++
		}
	}

	errorMsg := code[msgStart:msgEnd]

	// Build replacement: fmt.Fprintf(os.Stderr, "Error: %v\n", fmt.Errorf(...)) + os.Exit(1)
	replacement := fmt.Sprintf("fmt.Fprintf(os.Stderr, \"Error: %%v\\n\", fmt.Errorf(%s))\n\t\tos.Exit(1)", errorMsg)

	return code[:returnIdx] + replacement + code[msgEnd+1:]
}

// sortImports sorts imports with standard library first, then third-party
func sortImports(imports []string) {
	// Simple bubble sort - good enough for small import lists
	for i := 0; i < len(imports); i++ {
		for j := i + 1; j < len(imports); j++ {
			// Standard library imports (no dots) come first
			iStd := !containsChar(imports[i], '.')
			jStd := !containsChar(imports[j], '.')

			if (!iStd && jStd) || (iStd == jStd && imports[i] > imports[j]) {
				imports[i], imports[j] = imports[j], imports[i]
			}
		}
	}
}

// containsChar checks if string contains a character
func containsChar(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}
