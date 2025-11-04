package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestGenerationWithEnvVar tests that STREAMV3_GENERATE_GO env var works
func TestGenerationWithEnvVar(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "/tmp/streamv3_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build streamv3: %v", err)
	}
	defer os.Remove("/tmp/streamv3_test")

	tests := []struct {
		name    string
		cmdLine string // Full shell command line
		want    string // substring that should appear in output
	}{
		{
			name:    "read-csv generation",
			cmdLine: "export STREAMV3_GENERATE_GO=1 && /tmp/streamv3_test read-csv test.csv",
			want:    `"type":"init"`,
		},
		{
			name:    "where generation",
			cmdLine: `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test where -match age gt 18`,
			want:    `"type":"stmt"`,
		},
		{
			name:    "write-csv generation",
			cmdLine: `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test write-csv out.csv`,
			want:    `"type":"final"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run command using bash -c for proper pipeline execution
			cmd := exec.Command("bash", "-c", tt.cmdLine)

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("Command output: %s", output)
				// Some commands may error if files don't exist, but should still generate
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, tt.want) {
				t.Errorf("Expected output to contain %q, got: %s", tt.want, outputStr)
			}
		})
	}
}

// TestGenerationFlag tests that -generate flag works
func TestGenerationFlag(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "/tmp/streamv3_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build streamv3: %v", err)
	}
	defer os.Remove("/tmp/streamv3_test")

	cmd := exec.Command("/tmp/streamv3_test", "read-csv", "-generate", "test.csv")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Command output: %s", output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, `"type":"init"`) {
		t.Errorf("Expected output to contain init fragment, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, `streamv3.ReadCSV`) {
		t.Errorf("Expected output to contain ReadCSV call, got: %s", outputStr)
	}
}

// TestFullPipeline tests a complete generation pipeline
func TestFullPipeline(t *testing.T) {
	// This test ensures the full pipeline works end-to-end
	// Create a temporary CSV file
	csvContent := "name,age\nAlice,30\nBob,25\n"
	tmpFile := "/tmp/test_pipeline.csv"
	if err := os.WriteFile(tmpFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}
	defer os.Remove(tmpFile)

	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "/tmp/streamv3_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build streamv3: %v", err)
	}
	defer os.Remove("/tmp/streamv3_test")

	// Run pipeline: read-csv | where | generate-go
	pipeline := `export STREAMV3_GENERATE_GO=1 && /tmp/streamv3_test read-csv ` + tmpFile + ` | /tmp/streamv3_test where -match age gt 25 | /tmp/streamv3_test generate-go`
	cmd := exec.Command("bash", "-c", pipeline)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Pipeline failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	
	// Check for expected elements in generated code
	expectations := []string{
		"package main",
		"streamv3.ReadCSV",
		"streamv3.Where",
		"func(r streamv3.Record) bool",
		`streamv3.GetOr(r, "age"`,
	}

	for _, expected := range expectations {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Generated code missing expected element: %q\nGot: %s", expected, outputStr)
		}
	}
}

// TestGeneratedCodeCompiles tests that generated code actually compiles and runs
func TestGeneratedCodeCompiles(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "/tmp/streamv3_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build streamv3: %v", err)
	}
	defer os.Remove("/tmp/streamv3_test")

	// Create test CSV file
	csvContent := "name,age\nAlice,30\nBob,25\n"
	tmpFile := "/tmp/test_compile.csv"
	if err := os.WriteFile(tmpFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}
	defer os.Remove(tmpFile)

	// Generate code
	pipeline := `export STREAMV3_GENERATE_GO=1 && /tmp/streamv3_test read-csv ` + tmpFile + ` | /tmp/streamv3_test where -match age gt 25 | /tmp/streamv3_test generate-go`
	cmd := exec.Command("bash", "-c", pipeline)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Pipeline failed: %v\nOutput: %s", err, output)
	}

	// Write generated code to temp file
	generatedFile := "/tmp/test_generated.go"
	if err := os.WriteFile(generatedFile, output, 0644); err != nil {
		t.Fatalf("Failed to write generated code: %v", err)
	}
	defer os.Remove(generatedFile)

	// Check that generated code includes "os" import
	generatedCode := string(output)
	if !strings.Contains(generatedCode, `"os"`) {
		t.Errorf("Generated code missing 'os' import. This is needed for error handling.\nGenerated code:\n%s", generatedCode)
	}

	// Try to compile the generated code
	compileCmd := exec.Command("go", "build", "-o", "/tmp/test_generated_binary", generatedFile)
	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Generated code failed to compile: %v\nCompiler output:\n%s\nGenerated code:\n%s",
			err, compileOutput, generatedCode)
	}
	defer os.Remove("/tmp/test_generated_binary")

	t.Log("Generated code compiled successfully")
}

// TestLimitOffsetSortDistinct tests generation for limit, offset, sort, distinct commands
func TestLimitOffsetSortDistinct(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "/tmp/streamv3_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build streamv3: %v", err)
	}
	defer os.Remove("/tmp/streamv3_test")

	// Create test CSV file
	csvContent := "name,age\nAlice,30\nBob,25\nCharlie,35\n"
	tmpFile := "/tmp/test_limit_offset_sort.csv"
	if err := os.WriteFile(tmpFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}
	defer os.Remove(tmpFile)

	tests := []struct {
		name    string
		cmdLine string
		want    []string // substrings that should appear in output
	}{
		{
			name:    "limit command",
			cmdLine: `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test limit -n 5`,
			want:    []string{`"type":"stmt"`, `"var":"limited"`, `Limit[streamv3.Record](5)`},
		},
		{
			name:    "offset command",
			cmdLine: `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test offset -n 10`,
			want:    []string{`"type":"stmt"`, `"var":"skipped"`, `Offset[streamv3.Record](10)`},
		},
		{
			name:    "sort command",
			cmdLine: `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test sort -field age`,
			want:    []string{`"type":"stmt"`, `"var":"sorted"`, `SortBy`, `age`},
		},
		{
			name:    "distinct command",
			cmdLine: `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test distinct`,
			want:    []string{`"type":"stmt"`, `"var":"distinct"`, `DistinctBy`},
		},
		{
			name:    "pipeline with all commands",
			cmdLine: `export STREAMV3_GENERATE_GO=1 && /tmp/streamv3_test read-csv ` + tmpFile + ` | /tmp/streamv3_test where -match age gt 25 | /tmp/streamv3_test limit -n 5 | /tmp/streamv3_test offset -n 1 | /tmp/streamv3_test sort -field age -desc | /tmp/streamv3_test distinct | /tmp/streamv3_test generate-go`,
			want:    []string{"package main", "streamv3.ReadCSV", "streamv3.Where", "streamv3.Limit", "streamv3.Offset", "streamv3.SortBy", "streamv3.DistinctBy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("bash", "-c", tt.cmdLine)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("Command output: %s", output)
			}

			outputStr := string(output)
			for _, expected := range tt.want {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain %q, got: %s", expected, outputStr)
				}
			}
		})
	}
}
