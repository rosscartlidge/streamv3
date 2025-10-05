package streamv3

import (
	"encoding/json"
	"fmt"
	"iter"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ============================================================================
// STREAMV3 - PURE ITER.SEQ DESIGN WITH DUAL ERROR HANDLING
// ============================================================================

// ============================================================================
// CORE FUNCTIONAL TYPES
// ============================================================================

// Filter transforms one iterator to another with full type flexibility
// This is the heart of StreamV3 - same concept as StreamV2 but iter.Seq based
type Filter[T, U any] func(iter.Seq[T]) iter.Seq[U]

// FilterSameType is a convenience alias for same-type transformations
type FilterSameType[T any] func(iter.Seq[T]) iter.Seq[T]

// Error-aware filter variants for robust error handling
type FilterWithErrors[T, U any] func(iter.Seq2[T, error]) iter.Seq2[U, error]
type FilterWithErrorsSameType[T any] func(iter.Seq2[T, error]) iter.Seq2[T, error]

// ============================================================================
// COMPOSITION FUNCTIONS - IDENTICAL TO STREAMV2 PATTERNS
// ============================================================================

// Pipe composes two filters sequentially (T -> U -> V)
func Pipe[T, U, V any](f1 Filter[T, U], f2 Filter[U, V]) Filter[T, V] {
	return func(input iter.Seq[T]) iter.Seq[V] {
		return f2(f1(input))
	}
}

// Pipe3 composes three filters sequentially (T -> U -> V -> W)
func Pipe3[T, U, V, W any](f1 Filter[T, U], f2 Filter[U, V], f3 Filter[V, W]) Filter[T, W] {
	return func(input iter.Seq[T]) iter.Seq[W] {
		return f3(f2(f1(input)))
	}
}

// Chain applies multiple same-type filters in sequence
func Chain[T any](filters ...FilterSameType[T]) FilterSameType[T] {
	return func(input iter.Seq[T]) iter.Seq[T] {
		if len(filters) == 0 {
			return input // Identity function for no filters
		}
		result := input
		for _, filter := range filters {
			result = filter(result)
		}
		return result
	}
}

// Error-aware composition functions
func PipeWithErrors[T, U, V any](f1 FilterWithErrors[T, U], f2 FilterWithErrors[U, V]) FilterWithErrors[T, V] {
	return func(input iter.Seq2[T, error]) iter.Seq2[V, error] {
		return f2(f1(input))
	}
}

func ChainWithErrors[T any](filters ...FilterWithErrorsSameType[T]) FilterWithErrorsSameType[T] {
	return func(input iter.Seq2[T, error]) iter.Seq2[T, error] {
		if len(filters) == 0 {
			return input
		}
		result := input
		for _, filter := range filters {
			result = filter(result)
		}
		return result
	}
}

// ============================================================================
// RECORD SYSTEM - COMPATIBLE WITH STREAMV2
// ============================================================================

// Record represents structured data with native Go types
type Record map[string]any

// JSONString represents a string containing valid JSON data.
// This provides type safety and rich methods for working with JSON-structured data.
type JSONString string

// Parse parses the JSON string back to its original Go value
func (js JSONString) Parse() (any, error) {
	var result any
	err := json.Unmarshal([]byte(js), &result)
	return result, err
}

// MustParse parses the JSON string, panicking on error
func (js JSONString) MustParse() any {
	result, err := js.Parse()
	if err != nil {
		panic(fmt.Sprintf("failed to parse JSONString: %v", err))
	}
	return result
}

// IsValid returns true if the string contains valid JSON
func (js JSONString) IsValid() bool {
	var temp any
	return json.Unmarshal([]byte(js), &temp) == nil
}

// Pretty returns formatted JSON with indentation
func (js JSONString) Pretty() string {
	var temp any
	if err := json.Unmarshal([]byte(js), &temp); err != nil {
		return string(js) // Return original if invalid
	}
	pretty, err := json.MarshalIndent(temp, "", "  ")
	if err != nil {
		return string(js) // Return original if formatting fails
	}
	return string(pretty)
}

// String returns the underlying JSON string
func (js JSONString) String() string {
	return string(js)
}

// NewJSONString creates a JSONString by marshaling the given value
func NewJSONString(value any) (JSONString, error) {
	// Convert complex types to JSON-serializable form first
	jsonValue := convertToJSONValue(value)
	bytes, err := json.Marshal(jsonValue)
	if err != nil {
		return "", err
	}
	return JSONString(bytes), nil
}

// Value constraint for type-safe record values - matches StreamV2
type Value interface {
	// Integer types
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |

		// Float types
		~float32 | ~float64 |

		// Other basic types
		~bool | string | time.Time |

		// JSON and Record types for structured data
		JSONString | Record |

		// Iterator types for streams
		iter.Seq[int] | iter.Seq[int8] | iter.Seq[int16] | iter.Seq[int32] | iter.Seq[int64] |
		iter.Seq[uint] | iter.Seq[uint8] | iter.Seq[uint16] | iter.Seq[uint32] | iter.Seq[uint64] |
		iter.Seq[float32] | iter.Seq[float64] |
		iter.Seq[bool] | iter.Seq[string] | iter.Seq[time.Time] |
		iter.Seq[Record]
}

// ============================================================================
// TYPE-SAFE RECORD BUILDER - COMPATIBLE WITH STREAMV2
// ============================================================================

// TypedRecord provides type-safe field setting with method chaining
type TypedRecord struct {
	data map[string]any
}

// NewRecord creates a new type-safe Record builder
func NewRecord() *TypedRecord {
	return &TypedRecord{data: make(map[string]any)}
}

// Set adds a field with compile-time type safety
func Set[V Value](tr *TypedRecord, key string, value V) *TypedRecord {
	tr.data[key] = value
	return tr
}

// String adds a string field
func (tr *TypedRecord) String(key, value string) *TypedRecord {
	return Set(tr, key, value)
}

// JSONString adds a JSON string field with type safety
func (tr *TypedRecord) JSONString(key string, value JSONString) *TypedRecord {
	return Set(tr, key, value)
}

// Int adds an integer field
func (tr *TypedRecord) Int(key string, value int64) *TypedRecord {
	return Set(tr, key, value)
}

// Float adds a float field
func (tr *TypedRecord) Float(key string, value float64) *TypedRecord {
	return Set(tr, key, value)
}

// Bool adds a boolean field
func (tr *TypedRecord) Bool(key string, value bool) *TypedRecord {
	return Set(tr, key, value)
}

// Time adds a time field
func (tr *TypedRecord) Time(key string, value time.Time) *TypedRecord {
	return Set(tr, key, value)
}

// Record adds a nested record field
func (tr *TypedRecord) Record(key string, value Record) *TypedRecord {
	return Set(tr, key, value)
}

// ============================================================================
// ITER.SEQ FIELD METHODS
// ============================================================================

// IntSeq adds an iter.Seq[int] field
func (tr *TypedRecord) IntSeq(key string, value iter.Seq[int]) *TypedRecord {
	return Set(tr, key, value)
}

// Int8Seq adds an iter.Seq[int8] field
func (tr *TypedRecord) Int8Seq(key string, value iter.Seq[int8]) *TypedRecord {
	return Set(tr, key, value)
}

// Int16Seq adds an iter.Seq[int16] field
func (tr *TypedRecord) Int16Seq(key string, value iter.Seq[int16]) *TypedRecord {
	return Set(tr, key, value)
}

// Int32Seq adds an iter.Seq[int32] field
func (tr *TypedRecord) Int32Seq(key string, value iter.Seq[int32]) *TypedRecord {
	return Set(tr, key, value)
}

// Int64Seq adds an iter.Seq[int64] field
func (tr *TypedRecord) Int64Seq(key string, value iter.Seq[int64]) *TypedRecord {
	return Set(tr, key, value)
}

// UintSeq adds an iter.Seq[uint] field
func (tr *TypedRecord) UintSeq(key string, value iter.Seq[uint]) *TypedRecord {
	return Set(tr, key, value)
}

// Uint8Seq adds an iter.Seq[uint8] field
func (tr *TypedRecord) Uint8Seq(key string, value iter.Seq[uint8]) *TypedRecord {
	return Set(tr, key, value)
}

// Uint16Seq adds an iter.Seq[uint16] field
func (tr *TypedRecord) Uint16Seq(key string, value iter.Seq[uint16]) *TypedRecord {
	return Set(tr, key, value)
}

// Uint32Seq adds an iter.Seq[uint32] field
func (tr *TypedRecord) Uint32Seq(key string, value iter.Seq[uint32]) *TypedRecord {
	return Set(tr, key, value)
}

// Uint64Seq adds an iter.Seq[uint64] field
func (tr *TypedRecord) Uint64Seq(key string, value iter.Seq[uint64]) *TypedRecord {
	return Set(tr, key, value)
}

// Float32Seq adds an iter.Seq[float32] field
func (tr *TypedRecord) Float32Seq(key string, value iter.Seq[float32]) *TypedRecord {
	return Set(tr, key, value)
}

// Float64Seq adds an iter.Seq[float64] field
func (tr *TypedRecord) Float64Seq(key string, value iter.Seq[float64]) *TypedRecord {
	return Set(tr, key, value)
}

// BoolSeq adds an iter.Seq[bool] field
func (tr *TypedRecord) BoolSeq(key string, value iter.Seq[bool]) *TypedRecord {
	return Set(tr, key, value)
}

// StringSeq adds an iter.Seq[string] field
func (tr *TypedRecord) StringSeq(key string, value iter.Seq[string]) *TypedRecord {
	return Set(tr, key, value)
}

// TimeSeq adds an iter.Seq[time.Time] field
func (tr *TypedRecord) TimeSeq(key string, value iter.Seq[time.Time]) *TypedRecord {
	return Set(tr, key, value)
}

// RecordSeq adds an iter.Seq[Record] field
func (tr *TypedRecord) RecordSeq(key string, value iter.Seq[Record]) *TypedRecord {
	return Set(tr, key, value)
}

// Build finalizes the record construction
func (tr *TypedRecord) Build() Record {
	return Record(tr.data)
}

// ============================================================================
// TYPE-SAFE RECORD ACCESS WITH AUTOMATIC CONVERSION - FROM STREAMV2
// ============================================================================

// Get retrieves a typed value from a record with automatic conversion
func Get[T any](r Record, field string) (T, bool) {
	val, exists := r[field]
	if !exists {
		var zero T
		return zero, false
	}

	// Direct type assertion first (fast path)
	if typed, ok := val.(T); ok {
		return typed, true
	}

	// Smart type conversion (slower path)
	if converted, ok := convertTo[T](val); ok {
		return converted, true
	}

	var zero T
	return zero, false
}

// GetOr retrieves a typed value with a default fallback
func GetOr[T any](r Record, field string, defaultVal T) T {
	if val, ok := Get[T](r, field); ok {
		return val
	}
	return defaultVal
}

// SetField assigns a value to a record field with compile-time type safety
func SetField[V Value](r Record, field string, value V) Record {
	result := make(Record, len(r)+1)
	for k, v := range r {
		result[k] = v
	}
	result[field] = value
	return result
}

// Has checks if a field exists
func (r Record) Has(field string) bool {
	_, exists := r[field]
	return exists
}

// Keys returns all field names
func (r Record) Keys() []string {
	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}
	return keys
}

// Set creates a new Record with an additional field - immutable update
func (r Record) Set(field string, value any) Record {
	result := make(Record, len(r)+1)
	for k, v := range r {
		result[k] = v
	}
	result[field] = value
	return result
}

// ============================================================================
// SMART TYPE CONVERSION SYSTEM - FROM STREAMV2
// ============================================================================

func convertTo[T any](val any) (T, bool) {
	var zero T
	targetType := reflect.TypeOf(zero)

	// Handle nil
	if val == nil {
		return zero, false
	}

	sourceVal := reflect.ValueOf(val)

	// Try direct conversion for basic types
	if sourceVal.Type().ConvertibleTo(targetType) {
		converted := sourceVal.Convert(targetType)
		return converted.Interface().(T), true
	}

	// Custom conversions for common cases
	switch target := any(zero).(type) {
	case int64:
		if converted, ok := convertToInt64(val); ok {
			return any(converted).(T), true
		}
		return zero, false
	case float64:
		if converted, ok := convertToFloat64(val); ok {
			return any(converted).(T), true
		}
		return zero, false
	case string:
		if converted, ok := convertToString(val); ok {
			return any(converted).(T), true
		}
		return zero, false
	case bool:
		if converted, ok := convertToBool(val); ok {
			return any(converted).(T), true
		}
		return zero, false
	case time.Time:
		if converted, ok := convertToTime(val); ok {
			return any(converted).(T), true
		}
		return zero, false
	default:
		_ = target
		return zero, false
	}
}

func convertToInt64(val any) (int64, bool) {
	switch v := val.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case int16:
		return int64(v), true
	case int8:
		return int64(v), true
	case uint64:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint8:
		return int64(v), true
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return parsed, true
		}
		return 0, false
	default:
		return 0, false
	}
}

func convertToFloat64(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int16:
		return float64(v), true
	case int8:
		return float64(v), true
	case uint64:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint8:
		return float64(v), true
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed, true
		}
		return 0, false
	default:
		return 0, false
	}
}

func convertToString(val any) (string, bool) {
	switch v := val.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	default:
		return fmt.Sprintf("%v", val), true
	}
}

func convertToBool(val any) (bool, bool) {
	switch v := val.(type) {
	case bool:
		return v, true
	case int64:
		return v != 0, true
	case int:
		return v != 0, true
	case float64:
		return v != 0, true
	case string:
		return v != "", true
	default:
		return false, false
	}
}

func convertToTime(val any) (time.Time, bool) {
	switch v := val.(type) {
	case time.Time:
		return v, true
	case string:
		// Try RFC3339 first (most common for APIs)
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t, true
		}
		// Try standard SQL datetime format
		if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			return t, true
		}
		// Try RFC3339 without timezone (assume UTC)
		if t, err := time.Parse("2006-01-02T15:04:05", v); err == nil {
			return t.UTC(), true
		}
		return time.Time{}, false
	case int64:
		// Unix timestamp - always UTC
		return time.Unix(v, 0).UTC(), true
	default:
		return time.Time{}, false
	}
}

// ============================================================================
// RUNTIME VALIDATION HELPERS - FROM STREAMV2
// ============================================================================

// validateRecord checks if a Record has only Value-compatible field types
func ValidateRecord(r Record) error {
	for field, value := range r {
		if !isValueType(value) {
			return fmt.Errorf("field '%s' has invalid type %T", field, value)
		}
	}
	return nil
}

// IsValueType checks if a value conforms to the Value interface using type assertions
func isValueType(value any) bool {
	switch value.(type) {
	// Integer types
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	// Float types
	case float32, float64:
		return true
	// Other basic types
	case bool, string:
		return true
	case time.Time:
		return true
	// Record type
	case Record:
		return true
	// Iterator types for streams
	case iter.Seq[int], iter.Seq[int8], iter.Seq[int16], iter.Seq[int32], iter.Seq[int64]:
		return true
	case iter.Seq[uint], iter.Seq[uint8], iter.Seq[uint16], iter.Seq[uint32], iter.Seq[uint64]:
		return true
	case iter.Seq[float32], iter.Seq[float64]:
		return true
	case iter.Seq[bool], iter.Seq[string]:
		return true
	case iter.Seq[time.Time], iter.Seq[Record]:
		return true
	default:
		return false
	}
}

// Field creates a single-field Record with compile-time type safety
func Field[V Value](key string, value V) Record {
	return Record{key: value}
}

// ============================================================================
// NUMERIC AND COMPARABLE CONSTRAINTS - MATCHES STREAMV2
// ============================================================================

// Numeric represents types that support arithmetic operations
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Comparable represents types that can be compared and sorted
type Comparable interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string
}

// ============================================================================
// CONVERSION UTILITIES
// ============================================================================

// Safe converts a simple iterator to an error-aware iterator (never errors)
func Safe[T any](seq iter.Seq[T]) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for v := range seq {
			if !yield(v, nil) {
				return
			}
		}
	}
}

// Unsafe converts an error-aware iterator to simple iterator (panics on errors)
func Unsafe[T any](seq iter.Seq2[T, error]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v, err := range seq {
			if err != nil {
				panic(err)
			}
			if !yield(v) {
				return
			}
		}
	}
}

// IgnoreErrors converts an error-aware iterator to simple iterator (ignores errors)
func IgnoreErrors[T any](seq iter.Seq2[T, error]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v, err := range seq {
			if err != nil {
				continue // Skip items with errors
			}
			if !yield(v) {
				return
			}
		}
	}
}

// ============================================================================
// SEQUENCE MATERIALIZATION - EFFICIENT GROUPING SUPPORT
// ============================================================================

// Materialize converts an iter.Seq field to a string representation for efficient grouping.
// This is much more efficient than using CrossFlatten/DotFlatten when you only need
// to group by sequence content, not expand records.
//
// Example: {"tags": iter.Seq["urgent", "work"]} → {"tags": iter.Seq[...], "tags_key": "urgent,work"}
func Materialize(sourceField, targetField, separator string) FilterSameType[Record] {
	if separator == "" {
		separator = ","
	}

	return func(input iter.Seq[Record]) iter.Seq[Record] {
		return func(yield func(Record) bool) {
			for record := range input {
				result := make(Record)

				// Copy all existing fields
				for key, value := range record {
					result[key] = value
				}

				// Materialize the source field if it's an iter.Seq
				if sourceValue, exists := record[sourceField]; exists && isIterSeq(sourceValue) {
					values := materializeSequence(sourceValue)
					var stringValues []string
					for _, val := range values {
						stringValues = append(stringValues, fmt.Sprintf("%v", val))
					}
					result[targetField] = strings.Join(stringValues, separator)
				} else if exists {
					// Source field exists but isn't a sequence - convert to string
					result[targetField] = fmt.Sprintf("%v", sourceValue)
				}
				// If source field doesn't exist, don't add target field

				if !yield(result) {
					return
				}
			}
		}
	}
}

// MaterializeJSON converts complex fields (iter.Seq, Record, etc.) to JSON strings for grouping.
// This provides better type preservation and handles any JSON-serializable data structure.
// Unlike Materialize(), this works with Records, nested structures, and preserves type information.
func MaterializeJSON(sourceField, targetField string) FilterSameType[Record] {
	return func(input iter.Seq[Record]) iter.Seq[Record] {
		return func(yield func(Record) bool) {
			for record := range input {
				result := make(Record)

				// Copy all existing fields
				for key, value := range record {
					result[key] = value
				}

				// Materialize the source field if it exists
				if sourceValue, exists := record[sourceField]; exists {
					// Convert any complex field to JSON representation
					jsonValue := convertToJSONValue(sourceValue)
					if jsonBytes, err := json.Marshal(jsonValue); err == nil {
						result[targetField] = JSONString(jsonBytes)
					} else {
						// Fallback to string representation if JSON fails
						result[targetField] = fmt.Sprintf("%v", sourceValue)
					}
				}
				// If source field doesn't exist, don't add target field

				if !yield(result) {
					return
				}
			}
		}
	}
}

// ============================================================================
// RECORD FLATTENING OPERATIONS - FROM STREAMV2
// ============================================================================

// DotFlatten flattens nested records using dot product flattening (single output per input).
// Nested records become prefixed fields: {"user": {"name": "Alice"}} → {"user.name": "Alice"}
// iter.Seq fields are expanded using dot product (linear, one-to-one mapping).
// When sequences have different lengths, uses minimum length and discards excess elements.
// Example with sequences: {"id": 1, "tags": iter.Seq["a", "b"], "scores": iter.Seq[10, 20]} →
//   [{"id": 1, "tags": "a", "scores": 10}, {"id": 1, "tags": "b", "scores": 20}]
// Example with different lengths: {"short": iter.Seq["a", "b"], "long": iter.Seq[1, 2, 3, 4]} →
//   [{"short": "a", "long": 1}, {"short": "b", "long": 2}] (elements 3, 4 discarded)
func DotFlatten(separator string, fields ...string) FilterSameType[Record] {
	if separator == "" {
		separator = "."
	}

	return func(input iter.Seq[Record]) iter.Seq[Record] {
		return func(yield func(Record) bool) {
			for record := range input {
				// Expand the record (handling both nested records and sequences)
				expandedRecords := dotFlattenRecordWithSeqs(record, "", separator, fields...)

				// Yield all expanded records
				for _, expanded := range expandedRecords {
					if !yield(expanded) {
						return
					}
				}
			}
		}
	}
}

// CrossFlatten expands iter.Seq fields using cross product (cartesian product) expansion.
// Creates multiple output records from each input record containing sequence fields.
// Example: {"id": 1, "tags": iter.Seq["a", "b"]} → [{"id": 1, "tags": "a"}, {"id": 1, "tags": "b"}]
func CrossFlatten(separator string, fields ...string) FilterSameType[Record] {
	if separator == "" {
		separator = "."
	}

	return func(input iter.Seq[Record]) iter.Seq[Record] {
		return func(yield func(Record) bool) {
			for record := range input {
				// Use field-specific flatten algorithm
				flattened := crossFlattenRecord(record, separator, fields...)

				// Yield all flattened records
				for _, expanded := range flattened {
					if !yield(expanded) {
						return
					}
				}
			}
		}
	}
}

// dotFlattenRecord recursively flattens a record using dot notation
// If fields are specified, only flattens those top-level fields
func dotFlattenRecord(record Record, prefix, separator string, fields ...string) Record {
	result := make(Record)

	// Create a set of fields to flatten for quick lookup
	fieldsToFlatten := make(map[string]bool)
	if len(fields) > 0 {
		for _, field := range fields {
			fieldsToFlatten[field] = true
		}
	}

	for key, value := range record {
		newKey := key
		if prefix != "" {
			newKey = prefix + separator + key
		}

		// Check if this field should be flattened (only applies to top-level fields)
		shouldFlatten := len(fields) == 0 || prefix != "" || fieldsToFlatten[key]

		// If the value is a nested record, flatten it recursively
		if nestedRecord, ok := value.(Record); ok && shouldFlatten {
			flattened := dotFlattenRecord(nestedRecord, newKey, separator)
			for flatKey, flatValue := range flattened {
				result[flatKey] = flatValue
			}
		} else {
			// For non-record values (including sequences), or fields not to be flattened, keep as-is
			result[newKey] = value
		}
	}

	return result
}

// dotFlattenRecordWithSeqs flattens a record using dot product expansion for sequences
// Returns multiple records when sequences are present (dot product expansion)
// Uses minimum length when sequences have different lengths, discarding excess elements
func dotFlattenRecordWithSeqs(record Record, prefix, separator string, fields ...string) []Record {
	// Create a set of fields to flatten for quick lookup
	fieldsToFlatten := make(map[string]bool)
	if len(fields) > 0 {
		for _, field := range fields {
			fieldsToFlatten[field] = true
		}
	}

	// Collect all sequence fields that should be expanded
	var seqFields []string
	var seqValues [][]any
	var nonSeqRecord Record = make(Record)

	for key, value := range record {
		newKey := key
		if prefix != "" {
			newKey = prefix + separator + key
		}

		// Check if this field should be flattened (only applies to top-level fields)
		shouldFlatten := len(fields) == 0 || prefix != "" || fieldsToFlatten[key]

		// If the value is a nested record, flatten it recursively
		if nestedRecord, ok := value.(Record); ok && shouldFlatten {
			flattened := dotFlattenRecord(nestedRecord, newKey, separator)
			for flatKey, flatValue := range flattened {
				nonSeqRecord[flatKey] = flatValue
			}
		} else if shouldFlatten && isIterSeq(value) {
			// This is an iter.Seq field - collect its values for dot product expansion
			values := materializeSequence(value)
			if len(values) > 0 {
				seqFields = append(seqFields, newKey)
				seqValues = append(seqValues, values)
			}
		} else {
			// For non-record, non-sequence values, or fields not to be flattened, keep as-is
			nonSeqRecord[newKey] = value
		}
	}

	// If no sequence fields, return single record
	if len(seqFields) == 0 {
		return []Record{nonSeqRecord}
	}

	// Determine the length for dot product (use minimum length of all sequences)
	minLen := len(seqValues[0])
	for _, values := range seqValues[1:] {
		if len(values) < minLen {
			minLen = len(values)
		}
	}

	// Create dot product expansion - pair corresponding elements from each sequence
	var results []Record
	for i := 0; i < minLen; i++ {
		result := make(Record)

		// Copy non-sequence fields
		for key, value := range nonSeqRecord {
			result[key] = value
		}

		// Add corresponding element from each sequence
		for j, fieldName := range seqFields {
			result[fieldName] = seqValues[j][i]
		}

		results = append(results, result)
	}

	return results
}

// crossFlattenRecord expands specified sequence fields using cartesian product
// If no fields specified, expands all sequence fields
func crossFlattenRecord(r Record, sep string, fields ...string) []Record {
	var columns [][]Record
	var nonSeqFields []string

	// Create a set of fields to expand for quick lookup
	fieldsToExpand := make(map[string]bool)
	if len(fields) > 0 {
		for _, field := range fields {
			fieldsToExpand[field] = true
		}
	}

	for f, value := range r {
		if isIterSeq(value) {
			// Check if this field should be expanded
			shouldExpand := len(fields) == 0 || fieldsToExpand[f]

			if shouldExpand {
				values := materializeSequence(value)
				var rs []Record
				for _, val := range values {
					// Create a record with this sequence value
					newRecord := Record{f: val}
					rs = append(rs, newRecord)
				}
				if len(rs) > 0 {
					columns = append(columns, rs)
				}
			} else {
				// Keep sequence field as-is (don't expand)
				nonSeqFields = append(nonSeqFields, f)
			}
		} else {
			// Non-sequence field
			nonSeqFields = append(nonSeqFields, f)
		}
	}

	// If no sequence fields to expand, return original record
	if len(columns) == 0 {
		return []Record{r}
	}

	// Create cartesian product of expanded fields
	crs := cartesianProduct(columns)

	// Add non-sequence fields to each result record
	for _, cr := range crs {
		for _, f := range nonSeqFields {
			cr[f] = r[f]
		}
	}

	return crs
}

// cartesianProduct performs cartesian product of record slices
func cartesianProduct(columns [][]Record) []Record {
	if len(columns) == 0 {
		return nil
	}
	if len(columns) == 1 {
		return columns[0]
	}
	var rs []Record
	for _, lr := range cartesianProduct(columns[1:]) {
		for _, rr := range columns[0] {
			r := make(Record)
			for f := range rr {
				r[f] = rr[f]
			}
			for f := range lr {
				r[f] = lr[f]
			}
			rs = append(rs, r)
		}
	}
	return rs
}

// isIterSeq checks if a value is an iter.Seq type using reflection
func isIterSeq(value any) bool {
	if value == nil {
		return false
	}
	// Check for common iter.Seq types in Value constraint
	switch value.(type) {
	case iter.Seq[int], iter.Seq[int8], iter.Seq[int16], iter.Seq[int32], iter.Seq[int64]:
		return true
	case iter.Seq[uint], iter.Seq[uint8], iter.Seq[uint16], iter.Seq[uint32], iter.Seq[uint64]:
		return true
	case iter.Seq[float32], iter.Seq[float64]:
		return true
	case iter.Seq[bool], iter.Seq[string], iter.Seq[time.Time]:
		return true
	case iter.Seq[Record]:
		return true
	default:
		return false
	}
}

// isSimpleValue checks if a value is a simple type suitable for grouping
func isSimpleValue(value any) bool {
	if value == nil {
		return true // nil is simple
	}
	switch value.(type) {
	// Integer types
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	// Float types
	case float32, float64:
		return true
	// Other basic types
	case bool, string, time.Time:
		return true
	// Complex types not allowed for grouping
	case Record:
		return false
	default:
		// Check if it's an iter.Seq (not allowed)
		return !isIterSeq(value)
	}
}

// convertToJSONValue converts any value to a JSON-serializable representation
func convertToJSONValue(value any) any {
	switch v := value.(type) {
	case Record:
		// Convert nested Records recursively
		jsonRecord := make(map[string]any)
		for key, val := range v {
			jsonRecord[key] = convertToJSONValue(val)
		}
		return jsonRecord
	default:
		if isIterSeq(value) {
			// Convert iter.Seq to array
			return materializeSequence(value)
		}
		// Return other values as-is (primitives, already JSON-serializable)
		return value
	}
}

// materializeSequence converts an iter.Seq to a slice of any
func materializeSequence(value any) []any {
	var result []any

	switch seq := value.(type) {
	case iter.Seq[int]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[int8]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[int16]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[int32]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[int64]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[uint]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[uint8]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[uint16]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[uint32]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[uint64]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[float32]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[float64]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[bool]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[string]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[time.Time]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[Record]:
		for v := range seq { result = append(result, v) }
	}

	return result
}