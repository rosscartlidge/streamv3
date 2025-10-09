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
	seq := ReadCSV(filename)
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
	seq := ReadJSON(filename)
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
	seq := ReadLines(filename)
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

	seq := ExecCommand("echo", []string{"hello world"}, config)
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

	seq := ExecCommand("echo", []string{"hello"}, config)
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
	input := ReadCSV(inputFile)

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
	output := ReadCSV(outputFile)
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
	input := ReadJSON(filename)
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
