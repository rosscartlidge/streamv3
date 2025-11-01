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
			args = append(args, arg)
		}
	}
	return strings.Join(args, " ")
}
