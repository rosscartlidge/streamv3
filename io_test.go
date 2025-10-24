package streamv3

import (
	"bytes"
	"iter"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// ============================================================================
// CSV TESTS
// ============================================================================

func TestDefaultCSVConfig(t *testing.T) {
	config := DefaultCSVConfig()

	if config.Delimiter != ',' {
		t.Errorf("Default delimiter should be ',', got %c", config.Delimiter)
	}
	if config.Comment != '#' {
		t.Errorf("Default comment should be '#', got %c", config.Comment)
	}
	if !config.HasHeaders {
		t.Error("Default HasHeaders should be true")
	}
}

func TestReadCSVFromReader(t *testing.T) {
	csvData := `name,age,city
Alice,30,NYC
Bob,25,LA
Charlie,35,SF`

	reader := strings.NewReader(csvData)
	seq := ReadCSVFromReader(reader)
	result := slices.Collect(seq)

	if len(result) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(result))
	}

	if result[0]["name"] != "Alice" {
		t.Errorf("First record name should be Alice, got %v", result[0]["name"])
	}
	// CSV parsing converts numbers automatically
	if result[1]["age"] != int64(25) {
		t.Errorf("Second record age should be 25 (int64), got %v (type %T)", result[1]["age"], result[1]["age"])
	}
}

func TestReadCSVFromReaderWithCustomDelimiter(t *testing.T) {
	csvData := `name|age|city
Alice|30|NYC
Bob|25|LA`

	config := DefaultCSVConfig()
	config.Delimiter = '|'

	reader := strings.NewReader(csvData)
	seq := ReadCSVFromReader(reader, config)
	result := slices.Collect(seq)

	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}

	if result[0]["name"] != "Alice" {
		t.Errorf("Name should be Alice, got %v", result[0]["name"])
	}
}

func TestReadCSVSafeFromReader(t *testing.T) {
	csvData := `name,age
Alice,30
Bob,25`

	reader := strings.NewReader(csvData)
	seq := ReadCSVSafeFromReader(reader)

	var result []Record
	for record, err := range seq {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		result = append(result, record)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}
}

func TestWriteCSVToWriter(t *testing.T) {
	records := slices.Values([]Record{
		{"name": "Alice", "age": "30"},
		{"name": "Bob", "age": "25"},
	})

	var buf bytes.Buffer
	err := WriteCSVToWriter(records, &buf)
	if err != nil {
		t.Fatalf("WriteCSVToWriter failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Alice") {
		t.Error("Output should contain Alice")
	}
	if !strings.Contains(output, "Bob") {
		t.Error("Output should contain Bob")
	}
}

func TestReadWriteCSV(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.csv")

	// Write CSV
	records := slices.Values([]Record{
		{"name": "Alice", "age": int64(30), "city": "NYC"},
		{"name": "Bob", "age": int64(25), "city": "LA"},
	})

	err := WriteCSV(records, filename)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	// Read it back
	seq, err := ReadCSV(filename)
	if err != nil {
		t.Fatalf("ReadCSV failed: %v", err)
	}
	result := slices.Collect(seq)

	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}

	if result[0]["name"] != "Alice" {
		t.Errorf("Name should be Alice, got %v", result[0]["name"])
	}
}

func TestReadCSVSafe(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.csv")

	// Write test data
	data := []byte("name,age\nAlice,30\nBob,25\n")
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Read with safe version
	seq := ReadCSVSafe(filename)
	var result []Record
	for record, err := range seq {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		result = append(result, record)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}
}

// ============================================================================
// JSON TESTS
// ============================================================================

func TestReadJSONFromReader(t *testing.T) {
	jsonData := `{"name":"Alice","age":30}
{"name":"Bob","age":25}
{"name":"Charlie","age":35}`

	reader := strings.NewReader(jsonData)
	seq := ReadJSONFromReader(reader)
	result := slices.Collect(seq)

	if len(result) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(result))
	}

	if result[0]["name"] != "Alice" {
		t.Errorf("First record name should be Alice, got %v", result[0]["name"])
	}
}

func TestReadJSONSafeFromReader(t *testing.T) {
	jsonData := `{"name":"Alice","age":30}
{"name":"Bob","age":25}`

	reader := strings.NewReader(jsonData)
	seq := ReadJSONSafeFromReader(reader)

	var result []Record
	for record, err := range seq {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		result = append(result, record)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}
}

func TestWriteJSONToWriter(t *testing.T) {
	records := slices.Values([]Record{
		{"name": "Alice", "age": float64(30)},
		{"name": "Bob", "age": float64(25)},
	})

	var buf bytes.Buffer
	err := WriteJSONToWriter(records, &buf)
	if err != nil {
		t.Fatalf("WriteJSONToWriter failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Alice") {
		t.Error("Output should contain Alice")
	}
	if !strings.Contains(output, "Bob") {
		t.Error("Output should contain Bob")
	}
}

func TestReadWriteJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.json")

	// Write JSON
	records := slices.Values([]Record{
		{"name": "Alice", "age": float64(30)},
		{"name": "Bob", "age": float64(25)},
	})

	err := WriteJSON(records, filename)
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	// Read it back
	seq, err := ReadJSON(filename)
	if err != nil {
		t.Fatalf("ReadJSON failed: %v", err)
	}
	result := slices.Collect(seq)

	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}

	if result[0]["name"] != "Alice" {
		t.Errorf("Name should be Alice, got %v", result[0]["name"])
	}
}

func TestReadJSONSafe(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.json")

	// Write test data
	data := []byte(`{"name":"Alice","age":30}
{"name":"Bob","age":25}
`)
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Read with safe version
	seq := ReadJSONSafe(filename)
	var result []Record
	for record, err := range seq {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		result = append(result, record)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}
}

// ============================================================================
// LINES TESTS
// ============================================================================

func TestReadLines(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.txt")

	// Write test data
	data := []byte("line1\nline2\nline3\n")
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Read lines
	seq, err := ReadLines(filename)
	if err != nil {
		t.Fatalf("ReadLines failed: %v", err)
	}
	result := slices.Collect(seq)

	if len(result) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(result))
	}

	if result[0]["line"] != "line1" {
		t.Errorf("First line should be 'line1', got %v", result[0]["line"])
	}
}

func TestReadLinesSafe(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.txt")

	// Write test data
	data := []byte("line1\nline2\n")
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Read with safe version
	seq := ReadLinesSafe(filename)
	var result []Record
	for record, err := range seq {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		result = append(result, record)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}
}

func TestWriteLines(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.txt")

	// Write lines
	records := slices.Values([]Record{
		{"line": "first line"},
		{"line": "second line"},
	})

	err := WriteLines(records, filename)
	if err != nil {
		t.Fatalf("WriteLines failed: %v", err)
	}

	// Read it back
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "first line") {
		t.Error("Content should contain 'first line'")
	}
	if !strings.Contains(content, "second line") {
		t.Error("Content should contain 'second line'")
	}
}

// ============================================================================
// COMMAND TESTS
// ============================================================================

func TestDefaultCommandConfig(t *testing.T) {
	config := DefaultCommandConfig()

	if !config.HasHeaders {
		t.Error("Default HasHeaders should be true")
	}
	if !config.TrimSpaces {
		t.Error("Default TrimSpaces should be true")
	}
}

func TestExecCommand(t *testing.T) {
	// Test with echo command - need to disable headers since echo output isn't column-aligned
	config := DefaultCommandConfig()
	config.HasHeaders = false

	seq, err := ExecCommand("echo", []string{"hello world"}, config)
	if err != nil {
		t.Fatalf("ExecCommand failed: %v", err)
	}
	result := slices.Collect(seq)

	// Should have at least one record
	if len(result) == 0 {
		t.Error("ExecCommand should return at least one record")
	}

	// Verify the output contains our text
	if len(result) > 0 {
		rawLine := GetOr(result[0], "_raw_line", "")
		if !strings.Contains(rawLine, "hello world") {
			t.Errorf("Expected output to contain 'hello world', got: %s", rawLine)
		}
	}
}

func TestExecCommandSafe(t *testing.T) {
	// Test with printf to create column-aligned output
	// printf "NAME   AGE\nAlice  30\nBob    25"
	config := DefaultCommandConfig()
	config.HasHeaders = true

	seq := ExecCommandSafe("printf", []string{"NAME   AGE\\nAlice  30\\nBob    25"}, config)

	var result []Record
	var hasError bool
	for record, err := range seq {
		if err != nil {
			hasError = true
			t.Logf("Got error: %v", err)
			continue
		}
		result = append(result, record)
	}

	// We should get 2 data records (Alice and Bob)
	if !hasError && len(result) < 2 {
		t.Errorf("Expected at least 2 records, got %d", len(result))
	}
}

func TestExecCommandWithConfig(t *testing.T) {
	config := DefaultCommandConfig()
	config.TrimSpaces = true
	config.HasHeaders = false  // Disable headers for echo output

	seq, err := ExecCommand("echo", []string{"hello"}, config)
	if err != nil {
		t.Fatalf("ExecCommand failed: %v", err)
	}
	result := slices.Collect(seq)

	if len(result) == 0 {
		t.Fatal("Expected at least one record")
	}
}

// ============================================================================
// CHANNEL CONVERSION TESTS
// ============================================================================

func TestToChannel(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5})
	ch := ToChannel(input)

	var result []int
	for v := range ch {
		result = append(result, v)
	}

	expected := []int{1, 2, 3, 4, 5}
	if !slices.Equal(result, expected) {
		t.Errorf("ToChannel failed: expected %v, got %v", expected, result)
	}
}

func TestToChannelWithErrors(t *testing.T) {
	input := Safe(slices.Values([]int{1, 2, 3}))
	itemCh, errCh := ToChannelWithErrors(input)

	var result []int
	var errors []error

	done := make(chan bool)
	go func() {
		for err := range errCh {
			errors = append(errors, err)
		}
		done <- true
	}()

	for v := range itemCh {
		result = append(result, v)
	}

	<-done

	expected := []int{1, 2, 3}
	if !slices.Equal(result, expected) {
		t.Errorf("ToChannelWithErrors failed: expected %v, got %v", expected, result)
	}

	if len(errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(errors))
	}
}

func TestFromChannelSafe(t *testing.T) {
	itemCh := make(chan int, 3)
	errCh := make(chan error, 1) // Buffer the error channel

	// Send some values
	go func() {
		itemCh <- 1
		itemCh <- 2
		itemCh <- 3
		close(itemCh)
		close(errCh)
	}()

	seq := FromChannelSafe(itemCh, errCh)
	var result []int
	for v, err := range seq {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			continue
		}
		result = append(result, v)
	}

	expected := []int{1, 2, 3}
	if !slices.Equal(result, expected) {
		t.Errorf("FromChannelSafe failed: expected %v, got %v", expected, result)
	}
}

func TestChannelRoundTrip(t *testing.T) {
	// Test converting to channel and back
	input := slices.Values([]int{1, 2, 3, 4, 5})

	// Convert to channels
	itemCh, errCh := ToChannelWithErrors(Safe(input))

	// Convert back to iter.Seq2
	seq := FromChannelSafe(itemCh, errCh)

	// Collect results
	var result []int
	for v, err := range seq {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		result = append(result, v)
	}

	expected := []int{1, 2, 3, 4, 5}
	if !slices.Equal(result, expected) {
		t.Errorf("Channel round-trip failed: expected %v, got %v", expected, result)
	}
}

// ============================================================================
// INTEGRATION TESTS
// ============================================================================

func TestCSVPipeline(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.csv")
	outputFile := filepath.Join(tmpDir, "output.csv")

	// Write input CSV
	inputData := []byte("name,age,score\nAlice,30,85\nBob,25,90\nCharlie,35,75\n")
	err := os.WriteFile(inputFile, inputData, 0644)
	if err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	// Read, filter, and write
	input, err := ReadCSV(inputFile)
	if err != nil {
		t.Fatalf("ReadCSV failed: %v", err)
	}

	filtered := Where(func(r Record) bool {
		// CSV parses numbers automatically, so age is int64, not string
		ageInt := GetOr(r, "age", int64(0))
		return ageInt != int64(25)
	})(input)

	err = WriteCSV(filtered, outputFile)
	if err != nil {
		t.Fatalf("Failed to write output: %v", err)
	}

	// Read output and verify
	output, err := ReadCSV(outputFile)
	if err != nil {
		t.Fatalf("ReadCSV failed: %v", err)
	}
	result := slices.Collect(output)

	// Should have Alice and Charlie (filtered out Bob who has age 25)
	// CSV parsing converts numbers, so age is int64
	if len(result) != 2 {
		t.Fatalf("Expected 2 records after filtering, got %d", len(result))
	}

	// Verify Bob (age 25) was filtered out
	for _, r := range result {
		age := GetOr(r, "age", int64(0))
		if age == int64(25) {
			name := GetOr(r, "name", "")
			t.Errorf("Record with age 25 (%s) should have been filtered out", name)
		}
	}

	// Verify Alice and Charlie are present
	names := make(map[string]bool)
	for _, r := range result {
		name := GetOr(r, "name", "")
		names[name] = true
	}

	if !names["Alice"] {
		t.Error("Alice should be in the results")
	}
	if !names["Charlie"] {
		t.Error("Charlie should be in the results")
	}
}

func TestJSONPipeline(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.json")

	// Create and write records
	records := slices.Values([]Record{
		{"name": "Alice", "value": float64(100)},
		{"name": "Bob", "value": float64(200)},
		{"name": "Charlie", "value": float64(150)},
	})

	err := WriteJSON(records, filename)
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	// Read and process
	input, err := ReadJSON(filename)
	if err != nil {
		t.Fatalf("ReadJSON failed: %v", err)
	}
	filtered := Where(func(r Record) bool {
		value, ok := r["value"].(float64)
		return ok && value >= 150
	})(input)

	result := slices.Collect(filtered)

	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}
}

// ============================================================================
// ERROR HANDLING TESTS
// ============================================================================

func TestReadCSVNonExistentFile(t *testing.T) {
	seq := ReadCSVSafe("/nonexistent/file.csv")

	var hasError bool
	for _, err := range seq {
		if err != nil {
			hasError = true
			break
		}
	}

	if !hasError {
		t.Error("ReadCSVSafe should produce error for nonexistent file")
	}
}

func TestReadJSONNonExistentFile(t *testing.T) {
	seq := ReadJSONSafe("/nonexistent/file.json")

	var hasError bool
	for _, err := range seq {
		if err != nil {
			hasError = true
			break
		}
	}

	if !hasError {
		t.Error("ReadJSONSafe should produce error for nonexistent file")
	}
}

func TestReadLinesNonExistentFile(t *testing.T) {
	seq := ReadLinesSafe("/nonexistent/file.txt")

	var hasError bool
	for _, err := range seq {
		if err != nil {
			hasError = true
			break
		}
	}

	if !hasError {
		t.Error("ReadLinesSafe should produce error for nonexistent file")
	}
}

func TestWriteCSVInvalidPath(t *testing.T) {
	records := slices.Values([]Record{
		{"name": "Alice"},
	})

	err := WriteCSV(records, "/invalid/path/file.csv")
	if err == nil {
		t.Error("WriteCSV should return error for invalid path")
	}
}

func TestWriteJSONInvalidPath(t *testing.T) {
	records := slices.Values([]Record{
		{"name": "Alice"},
	})

	err := WriteJSON(records, "/invalid/path/file.json")
	if err == nil {
		t.Error("WriteJSON should return error for invalid path")
	}
}

// ============================================================================
// EDGE CASE TESTS
// ============================================================================

func TestReadEmptyCSV(t *testing.T) {
	csvData := ``
	reader := strings.NewReader(csvData)
	seq := ReadCSVFromReader(reader)
	result := slices.Collect(seq)

	if len(result) != 0 {
		t.Errorf("Empty CSV should return 0 records, got %d", len(result))
	}
}

func TestReadCSVHeaderOnly(t *testing.T) {
	csvData := `name,age,city`
	reader := strings.NewReader(csvData)
	seq := ReadCSVFromReader(reader)
	result := slices.Collect(seq)

	if len(result) != 0 {
		t.Errorf("CSV with only header should return 0 records, got %d", len(result))
	}
}

// TestCSVTypeParsing tests that CSV parsing correctly identifies types
// Regression test for bug where "1" was parsed as bool(true) instead of int64(1)
func TestCSVTypeParsing(t *testing.T) {
	csvData := `value,type
1,integer
0,integer
true,boolean
false,boolean
1.5,float
hello,string`

	reader := strings.NewReader(csvData)
	seq := ReadCSVFromReader(reader)
	result := slices.Collect(seq)

	if len(result) != 6 {
		t.Fatalf("Expected 6 records, got %d", len(result))
	}

	// Test case 1: "1" should be int64, NOT bool
	val1 := result[0]["value"]
	if _, ok := val1.(int64); !ok {
		t.Errorf("Value '1' should parse as int64, got %T(%v)", val1, val1)
	}
	if val1 != int64(1) {
		t.Errorf("Value '1' should equal int64(1), got %v", val1)
	}

	// Test case 2: "0" should be int64, NOT bool
	val0 := result[1]["value"]
	if _, ok := val0.(int64); !ok {
		t.Errorf("Value '0' should parse as int64, got %T(%v)", val0, val0)
	}
	if val0 != int64(0) {
		t.Errorf("Value '0' should equal int64(0), got %v", val0)
	}

	// Test case 3: "true" should be bool
	valTrue := result[2]["value"]
	if _, ok := valTrue.(bool); !ok {
		t.Errorf("Value 'true' should parse as bool, got %T(%v)", valTrue, valTrue)
	}
	if valTrue != true {
		t.Errorf("Value 'true' should equal true, got %v", valTrue)
	}

	// Test case 4: "false" should be bool
	valFalse := result[3]["value"]
	if _, ok := valFalse.(bool); !ok {
		t.Errorf("Value 'false' should parse as bool, got %T(%v)", valFalse, valFalse)
	}
	if valFalse != false {
		t.Errorf("Value 'false' should equal false, got %v", valFalse)
	}

	// Test case 5: "1.5" should be float64
	val15 := result[4]["value"]
	if _, ok := val15.(float64); !ok {
		t.Errorf("Value '1.5' should parse as float64, got %T(%v)", val15, val15)
	}
	if val15 != float64(1.5) {
		t.Errorf("Value '1.5' should equal float64(1.5), got %v", val15)
	}

	// Test case 6: "hello" should be string
	valHello := result[5]["value"]
	if _, ok := valHello.(string); !ok {
		t.Errorf("Value 'hello' should parse as string, got %T(%v)", valHello, valHello)
	}
	if valHello != "hello" {
		t.Errorf("Value 'hello' should equal 'hello', got %v", valHello)
	}
}

func TestReadEmptyJSON(t *testing.T) {
	jsonData := ``
	reader := strings.NewReader(jsonData)
	seq := ReadJSONFromReader(reader)
	result := slices.Collect(seq)

	if len(result) != 0 {
		t.Errorf("Empty JSON should return 0 records, got %d", len(result))
	}
}

func TestWriteEmptySequence(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "empty.csv")

	empty := func(yield func(Record) bool) {
		// Yield nothing
	}

	err := WriteCSV(iter.Seq[Record](empty), filename)
	if err != nil {
		t.Errorf("WriteCSV should handle empty sequence: %v", err)
	}
}

// ============================================================================
// ADVANCED INTEGRATION TESTS
// ============================================================================

// TestJSONComplexTypesRoundTrip tests JSON round-trip with complex types
// including iter.Seq fields, nested Records, and JSONString fields
func TestJSONComplexTypesRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "complex_roundtrip.json")

	// Create complex records with various types
	tags := slices.Values([]string{"urgent", "security"})
	scores := slices.Values([]int{95, 88, 92})
	weights := slices.Values([]float64{1.5, 2.3, 0.8})

	metadata := MakeMutableRecord().
		String("priority", "high").
		Int("version", 2)

	// Create JSONString field
	configJSON, err := NewJSONString(map[string]any{
		"timeout": 30,
		"retries": 3,
	})
	if err != nil {
		t.Fatalf("Failed to create JSONString: %v", err)
	}

	originalRecords := []Record{
		MakeMutableRecord().
			String("id", "TASK-001").
			String("title", "Security Update").
			Int("priority_num", 1).
			Float("score", 95.5).
			Bool("completed", false).
			StringSeq("tags", tags).
			IntSeq("scores", scores).
			Float64Seq("weights", weights).
			Nested("metadata", metadata.Freeze()).
			JSONString("config", configJSON).
			Freeze(),
		MakeMutableRecord().
			String("id", "TASK-002").
			String("title", "Feature Request").
			Int("priority_num", 2).
			Float("score", 87.2).
			Bool("completed", true).
			Freeze(),
	}

	// Write to JSON
	originalStream := slices.Values(originalRecords)
	err = WriteJSON(originalStream, filename)
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	// Read back from JSON
	reconstructedStream, err := ReadJSON(filename)
	if err != nil {
		t.Fatalf("ReadJSON failed: %v", err)
	}
	reconstructedRecords := slices.Collect(reconstructedStream)

	if len(reconstructedRecords) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(reconstructedRecords))
	}

	// Verify basic fields are preserved
	if reconstructedRecords[0]["id"] != "TASK-001" {
		t.Errorf("ID should be TASK-001, got %v", reconstructedRecords[0]["id"])
	}

	if reconstructedRecords[0]["title"] != "Security Update" {
		t.Errorf("Title should be 'Security Update', got %v", reconstructedRecords[0]["title"])
	}

	// Verify numeric fields (JSON converts to float64)
	scoreValue, ok := reconstructedRecords[0]["score"].(float64)
	if !ok || scoreValue != 95.5 {
		t.Errorf("Score should be 95.5 (float64), got %v (%T)", reconstructedRecords[0]["score"], reconstructedRecords[0]["score"])
	}

	// Verify boolean field
	completed, ok := reconstructedRecords[0]["completed"].(bool)
	if !ok || completed != false {
		t.Errorf("Completed should be false (bool), got %v (%T)", reconstructedRecords[0]["completed"], reconstructedRecords[0]["completed"])
	}

	// Verify iter.Seq fields become arrays
	tagsValue, ok := reconstructedRecords[0]["tags"].([]any)
	if !ok {
		t.Errorf("Tags should be array after round-trip, got %T", reconstructedRecords[0]["tags"])
	} else {
		if len(tagsValue) != 2 {
			t.Errorf("Tags should have 2 elements, got %d", len(tagsValue))
		}
		if tagsValue[0] != "urgent" || tagsValue[1] != "security" {
			t.Errorf("Tags data not preserved correctly: %v", tagsValue)
		}
	}

	// Verify Record fields become map[string]any
	metadataValue, ok := reconstructedRecords[0]["metadata"].(map[string]any)
	if !ok {
		t.Errorf("Metadata should be map after round-trip, got %T", reconstructedRecords[0]["metadata"])
	} else {
		if metadataValue["priority"] != "high" {
			t.Errorf("Metadata priority should be 'high', got %v", metadataValue["priority"])
		}
	}

	// Verify JSONString is parsed (not double-encoded)
	configValue, ok := reconstructedRecords[0]["config"].(map[string]any)
	if !ok {
		t.Errorf("Config should be parsed map, got %T", reconstructedRecords[0]["config"])
	} else {
		// JSON converts all numbers to float64
		if timeout, ok := configValue["timeout"].(float64); !ok || timeout != 30 {
			t.Errorf("Config timeout should be 30, got %v (%T)", configValue["timeout"], configValue["timeout"])
		}
	}
}

// TestJSONStreamProcessing tests process chaining via readers/writers
// simulating stdin/stdout processing
func TestJSONStreamProcessing(t *testing.T) {
	// Step 1: Create sample data
	salesData := []Record{
		MakeMutableRecord().
			String("product", "Laptop").
			Float("price", 1999.99).
			Int("quantity", 1).
			String("region", "North").
			Freeze(),
		MakeMutableRecord().
			String("product", "Phone").
			Float("price", 899.99).
			Int("quantity", 2).
			String("region", "South").
			Freeze(),
		MakeMutableRecord().
			String("product", "Tablet").
			Float("price", 399.99).
			Int("quantity", 1).
			String("region", "North").
			Freeze(),
	}

	// Step 2: Write to buffer (simulating first process output)
	var step1Output bytes.Buffer
	err := WriteJSONToWriter(slices.Values(salesData), &step1Output)
	if err != nil {
		t.Fatalf("Step 1 WriteJSONToWriter failed: %v", err)
	}

	// Step 3: Read from buffer and filter (simulating second process)
	step2Input := bytes.NewReader(step1Output.Bytes())
	inputStream := ReadJSONFromReader(step2Input)

	var filteredRecords []Record
	for record := range inputStream {
		price := GetOr(record, "price", float64(0))
		if price >= 500.0 {
			// Add calculated field
			quantity, _ := Get[float64](record, "quantity")
			record["total_value"] = price * quantity
			filteredRecords = append(filteredRecords, record)
		}
	}

	// Step 4: Write filtered output (simulating third process input)
	var step2Output bytes.Buffer
	err = WriteJSONToWriter(slices.Values(filteredRecords), &step2Output)
	if err != nil {
		t.Fatalf("Step 2 WriteJSONToWriter failed: %v", err)
	}

	// Step 5: Read and verify final output
	step3Input := bytes.NewReader(step2Output.Bytes())
	finalStream := ReadJSONFromReader(step3Input)
	finalRecords := slices.Collect(finalStream)

	// Should have Laptop and Phone (filtered out Tablet with price < 500)
	if len(finalRecords) != 2 {
		t.Fatalf("Expected 2 filtered records, got %d", len(finalRecords))
	}

	// Verify calculated field exists
	if _, ok := finalRecords[0]["total_value"]; !ok {
		t.Error("total_value field should be added during filtering")
	}

	// Verify data integrity through pipeline
	foundLaptop := false
	foundPhone := false
	for _, record := range finalRecords {
		product := GetOr(record, "product", "")
		if product == "Laptop" {
			foundLaptop = true
			totalValue := GetOr(record, "total_value", float64(0))
			if totalValue < 1999 { // Should be 1999.99 * 1
				t.Errorf("Laptop total_value incorrect: %v", totalValue)
			}
		}
		if product == "Phone" {
			foundPhone = true
		}
		if product == "Tablet" {
			t.Error("Tablet should have been filtered out")
		}
	}

	if !foundLaptop || !foundPhone {
		t.Error("Expected both Laptop and Phone in results")
	}
}

// TestFunctionalPipelineComposition tests complex functional composition
// with Chain, GroupBy, and Aggregate
func TestFunctionalPipelineComposition(t *testing.T) {
	// Create test data
	sales := []Record{
		MakeMutableRecord().String("region", "North").String("product", "Laptop").Float("amount", 1200).Freeze(),
		MakeMutableRecord().String("region", "South").String("product", "Phone").Float("amount", 800).Freeze(),
		MakeMutableRecord().String("region", "North").String("product", "Phone").Float("amount", 900).Freeze(),
		MakeMutableRecord().String("region", "East").String("product", "Laptop").Float("amount", 1100).Freeze(),
		MakeMutableRecord().String("region", "South").String("product", "Laptop").Float("amount", 1300).Freeze(),
		MakeMutableRecord().String("region", "North").String("product", "Tablet").Float("amount", 400).Freeze(),
	}

	// Test 1: Chain multiple Where filters
	chained := Chain(
		Where(func(r Record) bool {
			amount := GetOr(r, "amount", 0.0)
			return amount >= 800 // Filter high-value sales
		}),
		Where(func(r Record) bool {
			product := GetOr(r, "product", "")
			return product != "Tablet" // Exclude tablets
		}),
	)(slices.Values(sales))

	filteredCount := 0
	var filtered []Record
	for record := range chained {
		filteredCount++
		filtered = append(filtered, record)
	}

	// Should have 5 records (all except Tablet with 400, which is both < 800 and is Tablet)
	if filteredCount != 5 {
		t.Errorf("Expected 5 filtered records, got %d", filteredCount)
	}

	// Test 2: GroupBy and Aggregate composition
	grouped := GroupByFields("sales_data", "region")(slices.Values(filtered))

	aggregated := Aggregate("sales_data", map[string]AggregateFunc{
		"total_revenue": Sum("amount"),
		"avg_amount":    Avg("amount"),
		"count":         Count(),
	})(grouped)

	results := slices.Collect(aggregated)

	// Should have 3 regions (North, South, East)
	if len(results) != 3 {
		t.Errorf("Expected 3 regional summaries, got %d", len(results))
	}

	// Verify aggregation worked correctly
	regionTotals := make(map[string]float64)
	for _, result := range results {
		region := GetOr(result, "region", "")
		total := GetOr(result, "total_revenue", 0.0)
		count := GetOr(result, "count", int64(0))

		regionTotals[region] = total

		// Verify count is reasonable
		if count < 1 {
			t.Errorf("Region %s should have at least 1 sale, got %d", region, count)
		}

		// Verify average was calculated
		if _, ok := result["avg_amount"]; !ok {
			t.Errorf("avg_amount should be present for region %s", region)
		}
	}

	// Test 3: Step-by-step functional composition (same result)
	filtered2 := Where(func(r Record) bool {
		amount := GetOr(r, "amount", 0.0)
		product := GetOr(r, "product", "")
		return amount >= 800 && product != "Tablet"
	})(slices.Values(sales))

	grouped2 := GroupByFields("sales_data", "region")(filtered2)

	aggregated2 := Aggregate("sales_data", map[string]AggregateFunc{
		"total_revenue": Sum("amount"),
	})(grouped2)

	results2 := slices.Collect(aggregated2)

	// Should produce same results as chained version
	if len(results2) != len(results) {
		t.Errorf("Step-by-step composition should produce same number of results: expected %d, got %d", len(results), len(results2))
	}

	// Verify totals match between both approaches
	regionTotals2 := make(map[string]float64)
	for _, result := range results2 {
		region := GetOr(result, "region", "")
		total := GetOr(result, "total_revenue", 0.0)
		regionTotals2[region] = total
	}

	for region, total := range regionTotals {
		if regionTotals2[region] != total {
			t.Errorf("Region %s totals don't match: chained=%v, step-by-step=%v", region, total, regionTotals2[region])
		}
	}
}
