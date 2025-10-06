package streamv3

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"iter"
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

// ============================================================================
// CSV OPERATIONS WITH IO.READER/IO.WRITER
// ============================================================================

// ReadCSVFromReader reads CSV data from an io.Reader and returns an iterator
func ReadCSVFromReader(reader io.Reader, config ...CSVConfig) iter.Seq[Record] {
	cfg := DefaultCSVConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(yield func(Record) bool) {
		csvReader := csv.NewReader(reader)
		csvReader.Comma = cfg.Delimiter
		csvReader.Comment = cfg.Comment

		var headers []string
		if cfg.HasHeaders {
			headerRow, err := csvReader.Read()
			if err != nil {
				return
			}
			headers = headerRow
		}

		rowIndex := int64(0)
		for {
			row, err := csvReader.Read()
			if err != nil {
				return // EOF or error
			}

			record := make(Record)
			if cfg.HasHeaders && len(headers) > 0 {
				// Use headers as field names
				for i, value := range row {
					if i < len(headers) {
						record[headers[i]] = parseValue(value)
					}
				}
			} else {
				// Generate column names: col_0, col_1, etc.
				for i, value := range row {
					record[fmt.Sprintf("col_%d", i)] = parseValue(value)
				}
			}

			// Add row number
			record["_row_number"] = rowIndex
			rowIndex++

			if !yield(record) {
				return
			}
		}
	}
}

// ReadCSVSafeFromReader reads CSV data from an io.Reader with error handling
func ReadCSVSafeFromReader(reader io.Reader, config ...CSVConfig) *StreamWithErrors[Record] {
	cfg := DefaultCSVConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &StreamWithErrors[Record]{
		seq: func(yield func(Record, error) bool) {
			csvReader := csv.NewReader(reader)
			csvReader.Comma = cfg.Delimiter
			csvReader.Comment = cfg.Comment

			var headers []string
			if cfg.HasHeaders {
				headerRow, err := csvReader.Read()
				if err != nil {
					yield(nil, fmt.Errorf("failed to read CSV headers: %w", err))
					return
				}
				headers = headerRow
			}

			rowIndex := int64(0)
			for {
				row, err := csvReader.Read()
				if err == io.EOF {
					return
				}
				if err != nil {
					if !yield(nil, fmt.Errorf("failed to read CSV row %d: %w", rowIndex, err)) {
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

				record["_row_number"] = rowIndex
				rowIndex++

				if !yield(record, nil) {
					return
				}
			}
		},
	}
}

// WriteCSVToWriter writes records as CSV to an io.Writer
func WriteCSVToWriter(sb *Stream[Record], writer io.Writer, fields []string, config ...CSVConfig) error {
	cfg := DefaultCSVConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	csvWriter := csv.NewWriter(writer)
	csvWriter.Comma = cfg.Delimiter

	// Write headers if enabled
	if cfg.HasHeaders {
		if err := csvWriter.Write(fields); err != nil {
			return fmt.Errorf("failed to write CSV headers: %w", err)
		}
	}

	// Write data rows
	for record := range sb.Iter() {
		row := make([]string, len(fields))
		for i, field := range fields {
			if value, exists := record[field]; exists {
				row[i] = formatValue(value)
			}
		}

		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	csvWriter.Flush()
	return csvWriter.Error()
}

// ============================================================================
// CSV FILE CONVENIENCE FUNCTIONS
// ============================================================================

// ReadCSV reads CSV from a file and returns an iterator
func ReadCSV(filename string, config ...CSVConfig) iter.Seq[Record] {
	return func(yield func(Record) bool) {
		file, err := os.Open(filename)
		if err != nil {
			// For simple API, we skip errors - use ReadCSVSafe for error handling
			return
		}
		defer file.Close()

		// Use the io.Reader version
		for record := range ReadCSVFromReader(file, config...) {
			if !yield(record) {
				return
			}
		}
	}
}

// ReadCSVSafe reads CSV with error handling
func ReadCSVSafe(filename string, config ...CSVConfig) *StreamWithErrors[Record] {
	return &StreamWithErrors[Record]{
		seq: func(yield func(Record, error) bool) {
			file, err := os.Open(filename)
			if err != nil {
				if !yield(Record{}, fmt.Errorf("failed to open file %s: %w", filename, err)) {
					return
				}
				return
			}
			defer file.Close()

			// Use the io.Reader version
			for record, err := range ReadCSVSafeFromReader(file, config...).Iter() {
				if !yield(record, err) {
					return
				}
			}
		},
	}
}

// WriteCSV writes records to a CSV file
func WriteCSV(sb *Stream[Record], filename string, fields []string, config ...CSVConfig) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	// Use the io.Writer version
	return WriteCSVToWriter(sb, file, fields, config...)
}

// ============================================================================
// JSON OPERATIONS WITH IO.READER/IO.WRITER
// ============================================================================

// ReadJSONFromReader reads JSON records from an io.Reader (one JSON object per line)
func ReadJSONFromReader(reader io.Reader) *Stream[Record] {
	return &Stream[Record]{
		seq: func(yield func(Record) bool) {
			scanner := bufio.NewScanner(reader)
			lineNumber := int64(0)

			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					lineNumber++
					continue
				}

				var record Record
				if err := json.Unmarshal([]byte(line), &record); err != nil {
					// For simple API, skip invalid JSON lines
					lineNumber++
					continue
				}

				// Add line number metadata
				record["_line_number"] = lineNumber
				lineNumber++

				if !yield(record) {
					return
				}
			}
		},
	}
}

// ReadJSONSafeFromReader reads JSON records from an io.Reader with error handling
func ReadJSONSafeFromReader(reader io.Reader) *StreamWithErrors[Record] {
	return &StreamWithErrors[Record]{
		seq: func(yield func(Record, error) bool) {
			scanner := bufio.NewScanner(reader)
			lineNumber := int64(0)

			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					lineNumber++
					continue
				}

				var record Record
				if err := json.Unmarshal([]byte(line), &record); err != nil {
					if !yield(nil, fmt.Errorf("failed to parse JSON on line %d: %w", lineNumber, err)) {
						return
					}
					lineNumber++
					continue
				}

				record["_line_number"] = lineNumber
				lineNumber++

				if !yield(record, nil) {
					return
				}
			}

			// Check for scanner errors
			if err := scanner.Err(); err != nil {
				yield(nil, fmt.Errorf("error reading input: %w", err))
			}
		},
	}
}

// WriteJSONToWriter writes records as JSON to an io.Writer (one object per line)
func WriteJSONToWriter(sb *Stream[Record], writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	for record := range sb.Iter() {
		// Convert complex fields for JSON compatibility
		jsonRecord := make(Record)
		for key, value := range record {
			switch v := value.(type) {
			case JSONString:
				// Parse JSONString back to structured data to avoid double-encoding
				if parsed, err := v.Parse(); err == nil {
					jsonRecord[key] = parsed
				} else {
					// Fallback to string if parsing fails
					jsonRecord[key] = string(v)
				}
			default:
				if isIterSeq(value) {
					// Convert iter.Seq to array for JSON
					jsonRecord[key] = materializeSequence(value)
				} else {
					jsonRecord[key] = value
				}
			}
		}

		if err := encoder.Encode(jsonRecord); err != nil {
			return fmt.Errorf("failed to write JSON record: %w", err)
		}
	}

	return nil
}

// ============================================================================
// JSON FILE CONVENIENCE FUNCTIONS
// ============================================================================

// ReadJSON reads JSON records from a file (one JSON object per line)
func ReadJSON(filename string) *Stream[Record] {
	return &Stream[Record]{
		seq: func(yield func(Record) bool) {
			file, err := os.Open(filename)
			if err != nil {
				return // Skip errors in simple API
			}
			defer file.Close()

			// Use the io.Reader version
			for record := range ReadJSONFromReader(file).Iter() {
				if !yield(record) {
					return
				}
			}
		},
	}
}

// ReadJSONSafe reads JSON with error handling
func ReadJSONSafe(filename string) *StreamWithErrors[Record] {
	return &StreamWithErrors[Record]{
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
func WriteJSON(sb *Stream[Record], filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	// Use the io.Writer version
	return WriteJSONToWriter(sb, file)
}

// ============================================================================
// TEXT LINE OPERATIONS
// ============================================================================

// ReadLines reads text lines from a file
func ReadLines(filename string) *Stream[Record] {
	return &Stream[Record]{
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
func ReadLinesSafe(filename string) *StreamWithErrors[Record] {
	return &StreamWithErrors[Record]{
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
func WriteLines(sb *Stream[Record], filename string) error {
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
	case JSONString:
		return string(v) // For CSV, output the raw JSON string
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
		// Check if it's an iter.Seq type and materialize it
		if isIterSeq(value) {
			return formatIterSeq(value)
		}
		return fmt.Sprintf("%v", v)
	}
}

// formatIterSeq materializes an iter.Seq and formats it as a comma-separated string
func formatIterSeq(value any) string {
	// Materialize the sequence using the existing function
	values := materializeSequence(value)

	// Convert to strings and join
	var stringValues []string
	for _, val := range values {
		stringValues = append(stringValues, fmt.Sprintf("%v", val))
	}

	return strings.Join(stringValues, ",")
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
func ReadCommandOutput(filename string, config ...CommandConfig) *Stream[Record] {
	cfg := DefaultCommandConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &Stream[Record]{
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
func ReadCommandOutputSafe(filename string, config ...CommandConfig) *StreamWithErrors[Record] {
	cfg := DefaultCommandConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &StreamWithErrors[Record]{
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
func ExecCommand(command string, args []string, config ...CommandConfig) *Stream[Record] {
	cfg := DefaultCommandConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &Stream[Record]{
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
func ExecCommandSafe(command string, args []string, config ...CommandConfig) *StreamWithErrors[Record] {
	cfg := DefaultCommandConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return &StreamWithErrors[Record]{
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

// ToChannel converts a Stream to a channel
func ToChannel[T any](sb *Stream[T]) <-chan T {
	ch := make(chan T)
	go func() {
		defer close(ch)
		for item := range sb.Iter() {
			ch <- item
		}
	}()
	return ch
}

// ToChannelWithErrors converts a StreamWithErrors to channels
func ToChannelWithErrors[T any](sb *StreamWithErrors[T]) (<-chan T, <-chan error) {
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

// FromChannelSafe creates an error-aware Stream from channels
func FromChannelSafe[T any](itemCh <-chan T, errCh <-chan error) *StreamWithErrors[T] {
	return &StreamWithErrors[T]{
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