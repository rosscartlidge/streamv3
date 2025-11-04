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

// TestAllCommandsSupportGeneration ensures every command supports code generation
// This is a critical test to prevent losing the generation feature
func TestAllCommandsSupportGeneration(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "/tmp/streamv3_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build streamv3: %v", err)
	}
	defer os.Remove("/tmp/streamv3_test")

	// Create test CSV file for commands that need input files
	csvContent := "name,age,dept,salary\nAlice,30,Engineering,95000\nBob,25,Sales,75000\nCharlie,35,Engineering,105000\n"
	tmpFile := "/tmp/test_all_commands.csv"
	if err := os.WriteFile(tmpFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}
	defer os.Remove(tmpFile)

	tests := []struct {
		name          string
		cmdLine       string
		expectFragment bool // false for commands that shouldn't generate (like generate-go)
		wantSubstring string // substring to verify in generated code
	}{
		{
			name:           "read-csv",
			cmdLine:        "STREAMV3_GENERATE_GO=1 /tmp/streamv3_test read-csv " + tmpFile,
			expectFragment: true,
			wantSubstring:  `"type":"init"`,
		},
		{
			name:           "where",
			cmdLine:        `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test where -match age gt 25`,
			expectFragment: true,
			wantSubstring:  `streamv3.Where`,
		},
		{
			name:           "limit",
			cmdLine:        `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test limit -n 10`,
			expectFragment: true,
			wantSubstring:  `streamv3.Limit`,
		},
		{
			name:           "offset",
			cmdLine:        `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test offset -n 5`,
			expectFragment: true,
			wantSubstring:  `streamv3.Offset`,
		},
		{
			name:           "sort",
			cmdLine:        `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test sort -field age`,
			expectFragment: true,
			wantSubstring:  `streamv3.SortBy`,
		},
		{
			name:           "distinct",
			cmdLine:        `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test distinct`,
			expectFragment: true,
			wantSubstring:  `streamv3.DistinctBy`,
		},
		{
			name:           "group-by",
			cmdLine:        `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test group-by -by dept -func count -result count`,
			expectFragment: true,
			wantSubstring:  `streamv3.GroupByFields`,
		},
		{
			name:           "write-csv",
			cmdLine:        `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test write-csv /tmp/out.csv`,
			expectFragment: true,
			wantSubstring:  `streamv3.WriteCSV`,
		},
		{
			name:           "chart",
			cmdLine:        `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test chart -x age -y salary`,
			expectFragment: true,
			wantSubstring:  `streamv3.QuickChart`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("bash", "-c", tt.cmdLine)
			output, err := cmd.CombinedOutput()
			if err != nil && tt.expectFragment {
				// Some commands may error for other reasons, but should still generate
				t.Logf("Command had error (may be ok): %v\nOutput: %s", err, output)
			}

			outputStr := string(output)

			if tt.expectFragment {
				// Verify it generates a code fragment (JSONL output)
				if !strings.Contains(outputStr, `"type":`) {
					t.Errorf("Command %q did not generate a code fragment.\nOutput: %s", tt.name, outputStr)
				}

				// Verify expected code appears in fragment
				if !strings.Contains(outputStr, tt.wantSubstring) {
					t.Errorf("Command %q fragment missing expected substring %q.\nOutput: %s",
						tt.name, tt.wantSubstring, outputStr)
				}
			}
		})
	}
}

// TestChartGeneration specifically tests that chart generates code instead of creating HTML
func TestChartGeneration(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "/tmp/streamv3_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build streamv3: %v", err)
	}
	defer os.Remove("/tmp/streamv3_test")

	// Test that chart generates code when STREAMV3_GENERATE_GO=1
	cmdLine := `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test chart -x z_kind -y count`
	cmd := exec.Command("bash", "-c", cmdLine)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Command output: %s", output)
	}

	outputStr := string(output)

	// Should generate a code fragment
	if !strings.Contains(outputStr, `"type":"final"`) {
		t.Errorf("Chart command did not generate a final fragment.\nOutput: %s", outputStr)
	}

	// Should contain QuickChart call
	if !strings.Contains(outputStr, `streamv3.QuickChart`) {
		t.Errorf("Chart fragment missing QuickChart call.\nOutput: %s", outputStr)
	}

	// Verify chart.html was NOT created (generation shouldn't execute)
	if _, err := os.Stat("chart.html"); err == nil {
		t.Error("chart.html file was created when it should only generate code")
		os.Remove("chart.html") // Clean up
	}
}

// TestUpdateGeneration tests that the update command generates code correctly
func TestUpdateGeneration(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "/tmp/streamv3_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build streamv3: %v", err)
	}
	defer os.Remove("/tmp/streamv3_test")

	tests := []struct {
		name     string
		cmdLine  string
		wantStrs []string // substrings that should appear in output
	}{
		{
			name:    "single field update",
			cmdLine: `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test update -set status processed`,
			wantStrs: []string{
				`"type":"stmt"`,
				`"var":"updated"`,
				`streamv3.Update`,
				`mut = mut.String(\"status\", \"processed\")`,
			},
		},
		{
			name:    "multiple field update",
			cmdLine: `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test update -set status done -set count 42`,
			wantStrs: []string{
				`"type":"stmt"`,
				`streamv3.Update`,
				`mut = mut.String(\"status\", \"done\")`,
				`mut = mut.Int(\"count\", int64(42))`,
			},
		},
		{
			name:    "type inference - bool",
			cmdLine: `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test update -set active true`,
			wantStrs: []string{
				`mut = mut.Bool(\"active\", true)`,
			},
		},
		{
			name:    "type inference - float",
			cmdLine: `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test update -set price 99.99`,
			wantStrs: []string{
				`mut = mut.Float(\"price\", 99.9`,
			},
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

			for _, want := range tt.wantStrs {
				if !strings.Contains(outputStr, want) {
					t.Errorf("Expected output to contain %q, got: %s", want, outputStr)
				}
			}
		})
	}
}
