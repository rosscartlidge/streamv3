package commands

import "os"

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
