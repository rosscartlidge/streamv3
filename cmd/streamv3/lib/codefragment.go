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

	// Collect all imports and deduplicate
	importSet := make(map[string]bool)
	importSet["github.com/rosscartlidge/streamv3"] = true // Always needed

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

	// Add all fragment code statements
	for _, frag := range fragments {
		code += "\t" + frag.Code + "\n"
	}

	code += "}\n"

	return code, nil
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
