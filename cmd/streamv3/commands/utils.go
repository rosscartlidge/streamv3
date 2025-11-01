package commands

import (
	"os"
	"strings"
)

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
