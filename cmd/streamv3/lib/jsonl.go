package lib

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"os"

	"github.com/rosscartlidge/streamv3"
)

// Stdout is a convenience variable for writing to stdout
var Stdout io.WriteCloser = os.Stdout

// ReadJSONL reads JSONL (JSON Lines) from a reader and returns an iterator of Records
func ReadJSONL(r io.Reader) iter.Seq[streamv3.Record] {
	return func(yield func(streamv3.Record) bool) {
		scanner := bufio.NewScanner(r)

		// Increase buffer size for large lines
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024) // 1MB max token size

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue // Skip empty lines
			}

			// Parse JSON object
			var data map[string]interface{}
			if err := json.Unmarshal(line, &data); err != nil {
				// Skip malformed lines silently in streaming context
				continue
			}

			// Convert to Record directly (not using TypedRecord builder)
			record := streamv3.MakeMutableRecord()
			for k, v := range data {
				record = record.SetAny(k, convertJSONValue(v))
			}

			if !yield(record.Freeze()) {
				return
			}
		}
	}
}

// WriteJSONL writes Records to a writer as JSONL (JSON Lines)
func WriteJSONL(w io.Writer, records iter.Seq[streamv3.Record]) error {
	writer := bufio.NewWriter(w)
	defer writer.Flush()

	for record := range records {
		// Convert Record to map for JSON encoding
		data := make(map[string]interface{})

		// Extract all fields from record
		for k, v := range record.All() {
			data[k] = convertRecordValue(v)
		}

		// Encode as JSON
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("encoding record as JSON: %w", err)
		}

		// Write line
		if _, err := writer.Write(jsonBytes); err != nil {
			return fmt.Errorf("writing JSON line: %w", err)
		}
		if _, err := writer.Write([]byte("\n")); err != nil {
			return fmt.Errorf("writing newline: %w", err)
		}
	}

	return writer.Flush()
}

// OpenInput opens an input source (file or stdin)
func OpenInput(filename string) (io.ReadCloser, error) {
	if filename == "" || filename == "-" {
		// Check if stdin has data
		stat, err := os.Stdin.Stat()
		if err != nil {
			return nil, fmt.Errorf("checking stdin: %w", err)
		}
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return nil, fmt.Errorf("no input provided (use file or pipe data to stdin)")
		}
		return os.Stdin, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", filename, err)
	}
	return file, nil
}

// OpenOutput opens an output destination (file or stdout)
func OpenOutput(filename string) (io.WriteCloser, error) {
	if filename == "" || filename == "-" {
		return os.Stdout, nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("creating file %s: %w", filename, err)
	}
	return file, nil
}

// convertJSONValue converts JSON values to StreamV3 canonical types
func convertJSONValue(v interface{}) interface{} {
	switch val := v.(type) {
	case float64:
		// JSON numbers are always float64
		// Check if it's actually an integer
		if val == float64(int64(val)) {
			return int64(val)
		}
		return val
	case bool, string:
		return val
	case nil:
		return nil
	case []interface{}:
		// Convert array to slice for Record
		return val
	case map[string]interface{}:
		// Nested object - convert to Record
		record := streamv3.MakeMutableRecord()
		for k, subv := range val {
			record = record.SetAny(k, convertJSONValue(subv))
		}
		return record.Freeze()
	default:
		// For sequences and other types, try to convert to simple representation
		return fmt.Sprintf("%v", v)
	}
}

// convertRecordValue converts StreamV3 Record values to JSON-friendly types
func convertRecordValue(v interface{}) interface{} {
	switch val := v.(type) {
	case streamv3.Record:
		// Convert nested Record to map
		result := make(map[string]interface{})
		for k, subv := range val.All() {
			result[k] = convertRecordValue(subv)
		}
		return result
	case int64, float64, bool, string, nil:
		// Canonical types pass through
		return val
	default:
		// For sequences and other types, try to convert to simple representation
		return fmt.Sprintf("%v", v)
	}
}
