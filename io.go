package streamv3

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ============================================================================
// I/O OPERATIONS WITH DUAL ERROR HANDLING
// ============================================================================

// ============================================================================
// CSV OPERATIONS
// ============================================================================

// CSVConfig configures CSV reading and writing
type CSVConfig struct {
	HasHeaders bool
	Delimiter  rune
	Comment    rune
}

// DefaultCSVConfig provides sensible defaults for CSV processing
func DefaultCSVConfig() CSVConfig {
	return CSVConfig{
		HasHeaders: true,
		Delimiter:  ',',
		Comment:    '#',
	}
}

// ReadCSV reads CSV from a file and returns a StreamBuilder
func ReadCSV(filename string, config ...CSVConfig) *StreamBuilder[Record] {
	cfg := DefaultCSVConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &StreamBuilder[Record]{
		seq: func(yield func(Record) bool) {
			file, err := os.Open(filename)
			if err != nil {
				// For simple API, we skip errors - use ReadCSVSafe for error handling
				return
			}
			defer file.Close()

			reader := csv.NewReader(file)
			reader.Comma = cfg.Delimiter
			reader.Comment = cfg.Comment

			var headers []string
			if cfg.HasHeaders {
				headerRow, err := reader.Read()
				if err != nil {
					return
				}
				headers = headerRow
			}

			rowNum := 0
			for {
				row, err := reader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					continue // Skip invalid rows in simple API
				}

				record := make(Record)
				if cfg.HasHeaders && len(headers) > 0 {
					for i, value := range row {
						if i < len(headers) {
							record[headers[i]] = parseValue(value)
						}
					}
				} else {
					for i, value := range row {
						record[fmt.Sprintf("col_%d", i)] = parseValue(value)
					}
				}

				// Add row metadata
				record["_row_number"] = int64(rowNum)
				rowNum++

				if !yield(record) {
					return
				}
			}
		},
	}
}

// ReadCSVSafe reads CSV with error handling
func ReadCSVSafe(filename string, config ...CSVConfig) *StreamBuilderWithErrors[Record] {
	cfg := DefaultCSVConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &StreamBuilderWithErrors[Record]{
		seq: func(yield func(Record, error) bool) {
			file, err := os.Open(filename)
			if err != nil {
				if !yield(Record{}, fmt.Errorf("failed to open file %s: %w", filename, err)) {
					return
				}
				return
			}
			defer file.Close()

			reader := csv.NewReader(file)
			reader.Comma = cfg.Delimiter
			reader.Comment = cfg.Comment

			var headers []string
			if cfg.HasHeaders {
				headerRow, err := reader.Read()
				if err != nil {
					if !yield(Record{}, fmt.Errorf("failed to read CSV headers: %w", err)) {
						return
					}
					return
				}
				headers = headerRow
			}

			rowNum := 0
			for {
				row, err := reader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					if !yield(Record{}, fmt.Errorf("error reading CSV row %d: %w", rowNum, err)) {
						return
					}
					continue
				}

				record := make(Record)
				if cfg.HasHeaders && len(headers) > 0 {
					for i, value := range row {
						if i < len(headers) {
							record[headers[i]] = parseValue(value)
						}
					}
				} else {
					for i, value := range row {
						record[fmt.Sprintf("col_%d", i)] = parseValue(value)
					}
				}

				// Add row metadata
				record["_row_number"] = int64(rowNum)
				rowNum++

				if !yield(record, nil) {
					return
				}
			}
		},
	}
}

// WriteCSV writes records to a CSV file
func WriteCSV(sb *StreamBuilder[Record], filename string, fields []string, config ...CSVConfig) error {
	cfg := DefaultCSVConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Comma = cfg.Delimiter

	// Write headers if configured
	if cfg.HasHeaders {
		if err := writer.Write(fields); err != nil {
			return fmt.Errorf("failed to write CSV headers: %w", err)
		}
	}

	// Write records
	for record := range sb.Iter() {
		row := make([]string, len(fields))
		for i, field := range fields {
			if value, exists := record[field]; exists {
				row[i] = formatValue(value)
			}
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ============================================================================
// JSON OPERATIONS
// ============================================================================

// ReadJSON reads JSON records from a file (one JSON object per line)
func ReadJSON(filename string) *StreamBuilder[Record] {
	return &StreamBuilder[Record]{
		seq: func(yield func(Record) bool) {
			file, err := os.Open(filename)
			if err != nil {
				return // Skip errors in simple API
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			lineNum := 0
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}

				var record Record
				if err := json.Unmarshal([]byte(line), &record); err != nil {
					continue // Skip invalid JSON in simple API
				}

				// Add line metadata
				record["_line_number"] = int64(lineNum)
				lineNum++

				if !yield(record) {
					return
				}
			}
		},
	}
}

// ReadJSONSafe reads JSON with error handling
func ReadJSONSafe(filename string) *StreamBuilderWithErrors[Record] {
	return &StreamBuilderWithErrors[Record]{
		seq: func(yield func(Record, error) bool) {
			file, err := os.Open(filename)
			if err != nil {
				if !yield(Record{}, fmt.Errorf("failed to open file %s: %w", filename, err)) {
					return
				}
				return
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			lineNum := 0
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}

				var record Record
				if err := json.Unmarshal([]byte(line), &record); err != nil {
					if !yield(Record{}, fmt.Errorf("invalid JSON on line %d: %w", lineNum, err)) {
						return
					}
					continue
				}

				// Add line metadata
				record["_line_number"] = int64(lineNum)
				lineNum++

				if !yield(record, nil) {
					return
				}
			}

			if err := scanner.Err(); err != nil {
				yield(Record{}, fmt.Errorf("error reading file: %w", err))
			}
		},
	}
}

// WriteJSON writes records as JSON (one object per line)
func WriteJSON(sb *StreamBuilder[Record], filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for record := range sb.Iter() {
		if err := encoder.Encode(record); err != nil {
			return fmt.Errorf("failed to write JSON record: %w", err)
		}
	}

	return nil
}

// ============================================================================
// TEXT LINE OPERATIONS
// ============================================================================

// ReadLines reads text lines from a file
func ReadLines(filename string) *StreamBuilder[Record] {
	return &StreamBuilder[Record]{
		seq: func(yield func(Record) bool) {
			file, err := os.Open(filename)
			if err != nil {
				return // Skip errors in simple API
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			lineNum := 0
			for scanner.Scan() {
				record := Record{
					"line":        scanner.Text(),
					"line_number": int64(lineNum),
				}
				lineNum++

				if !yield(record) {
					return
				}
			}
		},
	}
}

// ReadLinesSafe reads text lines with error handling
func ReadLinesSafe(filename string) *StreamBuilderWithErrors[Record] {
	return &StreamBuilderWithErrors[Record]{
		seq: func(yield func(Record, error) bool) {
			file, err := os.Open(filename)
			if err != nil {
				if !yield(Record{}, fmt.Errorf("failed to open file %s: %w", filename, err)) {
					return
				}
				return
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			lineNum := 0
			for scanner.Scan() {
				record := Record{
					"line":        scanner.Text(),
					"line_number": int64(lineNum),
				}
				lineNum++

				if !yield(record, nil) {
					return
				}
			}

			if err := scanner.Err(); err != nil {
				yield(Record{}, fmt.Errorf("error reading file: %w", err))
			}
		},
	}
}

// WriteLines writes records as text lines (using "line" field)
func WriteLines(sb *StreamBuilder[Record], filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for record := range sb.Iter() {
		line := ""
		if value, exists := record["line"]; exists {
			line = formatValue(value)
		}
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write line: %w", err)
		}
	}

	return nil
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// parseValue attempts to parse a string value into appropriate Go types
func parseValue(s string) any {
	s = strings.TrimSpace(s)

	// Try boolean
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}

	// Try integer
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}

	// Try float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	// Default to string
	return s
}

// formatValue converts a value to string for output
func formatValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ============================================================================
// COMMAND OUTPUT OPERATIONS
// ============================================================================

// CommandConfig configures command output parsing
type CommandConfig struct {
	HasHeaders    bool   // Whether first line contains column headers
	TrimSpaces    bool   // Whether to trim leading/trailing spaces from fields
	SkipEmpty     bool   // Whether to skip empty lines
	HeaderPattern string // Optional regex pattern to identify header line
}

// DefaultCommandConfig provides sensible defaults for command output parsing
func DefaultCommandConfig() CommandConfig {
	return CommandConfig{
		HasHeaders: true,
		TrimSpaces: true,
		SkipEmpty:  true,
	}
}

// ReadCommandOutput reads command output with column-aligned data
func ReadCommandOutput(filename string, config ...CommandConfig) *StreamBuilder[Record] {
	cfg := DefaultCommandConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &StreamBuilder[Record]{
		seq: func(yield func(Record) bool) {
			file, err := os.Open(filename)
			if err != nil {
				return // Skip errors in simple API
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)

			var columnPositions []ColumnInfo
			lineNum := 0
			headerProcessed := false

			for scanner.Scan() {
				line := scanner.Text()

				// Skip empty lines if configured
				if cfg.SkipEmpty && strings.TrimSpace(line) == "" {
					continue
				}

				// Process header line
				if cfg.HasHeaders && !headerProcessed {
					columnPositions = parseHeaderLine(line)
					headerProcessed = true
					lineNum++
					continue
				}

				// Parse data line
				record := parseDataLine(line, columnPositions, cfg.TrimSpaces)
				record["_line_number"] = int64(lineNum)
				record["_raw_line"] = line

				lineNum++

				if !yield(record) {
					return
				}
			}
		},
	}
}

// ReadCommandOutputSafe reads command output with error handling
func ReadCommandOutputSafe(filename string, config ...CommandConfig) *StreamBuilderWithErrors[Record] {
	cfg := DefaultCommandConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &StreamBuilderWithErrors[Record]{
		seq: func(yield func(Record, error) bool) {
			file, err := os.Open(filename)
			if err != nil {
				if !yield(Record{}, fmt.Errorf("failed to open file %s: %w", filename, err)) {
					return
				}
				return
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)

			var columnPositions []ColumnInfo
			lineNum := 0
			headerProcessed := false

			for scanner.Scan() {
				line := scanner.Text()

				// Skip empty lines if configured
				if cfg.SkipEmpty && strings.TrimSpace(line) == "" {
					continue
				}

				// Process header line
				if cfg.HasHeaders && !headerProcessed {
					columnPositions = parseHeaderLine(line)
					if len(columnPositions) == 0 {
						if !yield(Record{}, fmt.Errorf("failed to parse header line: %s", line)) {
							return
						}
						continue
					}
					headerProcessed = true
					lineNum++
					continue
				}

				// Parse data line
				record, parseErr := parseDataLineSafe(line, columnPositions, cfg.TrimSpaces)
				if parseErr != nil {
					if !yield(Record{}, fmt.Errorf("error parsing line %d: %w", lineNum, parseErr)) {
						return
					}
					continue
				}

				record["_line_number"] = int64(lineNum)
				record["_raw_line"] = line

				lineNum++

				if !yield(record, nil) {
					return
				}
			}

			if err := scanner.Err(); err != nil {
				yield(Record{}, fmt.Errorf("error reading file: %w", err))
			}
		},
	}
}

// ExecCommand executes a command and returns its output as a stream
func ExecCommand(command string, args []string, config ...CommandConfig) *StreamBuilder[Record] {
	cfg := DefaultCommandConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &StreamBuilder[Record]{
		seq: func(yield func(Record) bool) {
			cmd := exec.Command(command, args...)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				return // Skip errors in simple API
			}

			if err := cmd.Start(); err != nil {
				return
			}
			defer cmd.Wait()

			scanner := bufio.NewScanner(stdout)

			var columnPositions []ColumnInfo
			lineNum := 0
			headerProcessed := false

			for scanner.Scan() {
				line := scanner.Text()

				// Skip empty lines if configured
				if cfg.SkipEmpty && strings.TrimSpace(line) == "" {
					continue
				}

				// Process header line
				if cfg.HasHeaders && !headerProcessed {
					columnPositions = parseHeaderLine(line)
					headerProcessed = true
					lineNum++
					continue
				}

				// Parse data line
				record := parseDataLine(line, columnPositions, cfg.TrimSpaces)
				record["_line_number"] = int64(lineNum)
				record["_raw_line"] = line
				record["_command"] = command

				lineNum++

				if !yield(record) {
					return
				}
			}
		},
	}
}

// ExecCommandSafe executes a command with error handling
func ExecCommandSafe(command string, args []string, config ...CommandConfig) *StreamBuilderWithErrors[Record] {
	cfg := DefaultCommandConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &StreamBuilderWithErrors[Record]{
		seq: func(yield func(Record, error) bool) {
			cmd := exec.Command(command, args...)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				if !yield(Record{}, fmt.Errorf("failed to create stdout pipe: %w", err)) {
					return
				}
				return
			}

			if err := cmd.Start(); err != nil {
				if !yield(Record{}, fmt.Errorf("failed to start command: %w", err)) {
					return
				}
				return
			}
			defer func() {
				if waitErr := cmd.Wait(); waitErr != nil {
					yield(Record{}, fmt.Errorf("command failed: %w", waitErr))
				}
			}()

			scanner := bufio.NewScanner(stdout)

			var columnPositions []ColumnInfo
			lineNum := 0
			headerProcessed := false

			for scanner.Scan() {
				line := scanner.Text()

				// Skip empty lines if configured
				if cfg.SkipEmpty && strings.TrimSpace(line) == "" {
					continue
				}

				// Process header line
				if cfg.HasHeaders && !headerProcessed {
					columnPositions = parseHeaderLine(line)
					if len(columnPositions) == 0 {
						if !yield(Record{}, fmt.Errorf("failed to parse header line: %s", line)) {
							return
						}
						continue
					}
					headerProcessed = true
					lineNum++
					continue
				}

				// Parse data line
				record, parseErr := parseDataLineSafe(line, columnPositions, cfg.TrimSpaces)
				if parseErr != nil {
					if !yield(Record{}, fmt.Errorf("error parsing line %d: %w", lineNum, parseErr)) {
						return
					}
					continue
				}

				record["_line_number"] = int64(lineNum)
				record["_raw_line"] = line
				record["_command"] = command

				lineNum++

				if !yield(record, nil) {
					return
				}
			}

			if err := scanner.Err(); err != nil {
				yield(Record{}, fmt.Errorf("error reading command output: %w", err))
			}
		},
	}
}

// ============================================================================
// COMMAND OUTPUT PARSING HELPERS
// ============================================================================

// ColumnInfo holds information about a column's position and name
type ColumnInfo struct {
	Name  string
	Start int
	End   int // -1 for last column (extends to end of line)
}

// parseHeaderLine analyzes the header line to determine column positions
func parseHeaderLine(line string) []ColumnInfo {
	if strings.TrimSpace(line) == "" {
		return nil
	}

	var columns []ColumnInfo
	var currentCol ColumnInfo
	inField := false

	for i, char := range line {
		if char != ' ' && char != '\t' {
			if !inField {
				// Start of new field
				currentCol = ColumnInfo{Start: i}
				inField = true
			}
		} else {
			if inField {
				// End of current field
				currentCol.Name = strings.TrimSpace(line[currentCol.Start:i])
				currentCol.End = i
				columns = append(columns, currentCol)
				inField = false
			}
		}
	}

	// Handle last column
	if inField {
		currentCol.Name = strings.TrimSpace(line[currentCol.Start:])
		currentCol.End = -1 // Last column extends to end of line
		columns = append(columns, currentCol)
	}

	return columns
}

// parseDataLine extracts field values from a data line using column positions
func parseDataLine(line string, columns []ColumnInfo, trimSpaces bool) Record {
	record := make(Record)

	for i, col := range columns {
		var value string

		if col.Start >= len(line) {
			value = ""
		} else if col.End == -1 || i == len(columns)-1 {
			// Last column - take everything from start to end of line
			value = line[col.Start:]
		} else if col.End <= len(line) {
			value = line[col.Start:col.End]
		} else {
			value = line[col.Start:]
		}

		if trimSpaces {
			value = strings.TrimSpace(value)
		}

		record[col.Name] = parseCommandValue(value)
	}

	return record
}

// parseDataLineSafe is like parseDataLine but returns errors
func parseDataLineSafe(line string, columns []ColumnInfo, trimSpaces bool) (Record, error) {
	if len(columns) == 0 {
		return nil, fmt.Errorf("no column information available")
	}

	record := make(Record)

	for i, col := range columns {
		var value string

		if col.Start >= len(line) {
			value = ""
		} else if col.End == -1 || i == len(columns)-1 {
			// Last column - take everything from start to end of line
			value = line[col.Start:]
		} else if col.End <= len(line) {
			value = line[col.Start:col.End]
		} else {
			value = line[col.Start:]
		}

		if trimSpaces {
			value = strings.TrimSpace(value)
		}

		record[col.Name] = parseCommandValue(value)
	}

	return record, nil
}

// parseCommandValue parses command output values with integer priority over boolean
func parseCommandValue(s string) any {
	s = strings.TrimSpace(s)

	// Empty string
	if s == "" {
		return ""
	}

	// Try integer first (before boolean to avoid "1"/"0" being parsed as bool)
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}

	// Try float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	// Try boolean (only for explicit true/false, not 1/0)
	if s == "true" || s == "false" {
		if b, err := strconv.ParseBool(s); err == nil {
			return b
		}
	}

	// Default to string
	return s
}

// ============================================================================
// CHANNEL OPERATIONS
// ============================================================================

// ToChannel converts a StreamBuilder to a channel
func ToChannel[T any](sb *StreamBuilder[T]) <-chan T {
	ch := make(chan T)
	go func() {
		defer close(ch)
		for item := range sb.Iter() {
			ch <- item
		}
	}()
	return ch
}

// ToChannelWithErrors converts a StreamBuilderWithErrors to channels
func ToChannelWithErrors[T any](sb *StreamBuilderWithErrors[T]) (<-chan T, <-chan error) {
	itemCh := make(chan T)
	errCh := make(chan error, 1)

	go func() {
		defer close(itemCh)
		defer close(errCh)

		for item, err := range sb.Iter() {
			if err != nil {
				errCh <- err
				return
			}
			itemCh <- item
		}
	}()

	return itemCh, errCh
}

// FromChannelSafe creates an error-aware StreamBuilder from channels
func FromChannelSafe[T any](itemCh <-chan T, errCh <-chan error) *StreamBuilderWithErrors[T] {
	return &StreamBuilderWithErrors[T]{
		seq: func(yield func(T, error) bool) {
			for {
				select {
				case item, ok := <-itemCh:
					if !ok {
						return // Channel closed
					}
					if !yield(item, nil) {
						return
					}
				case err, ok := <-errCh:
					if !ok {
						return // Error channel closed
					}
					var zero T
					if !yield(zero, err) {
						return
					}
				}
			}
		},
	}
}