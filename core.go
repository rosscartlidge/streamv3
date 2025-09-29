package streamv3

import (
	"fmt"
	"iter"
	"reflect"
	"strconv"
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

// Value constraint for type-safe record values - matches StreamV2
type Value interface {
	// Integer types
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |

		// Float types
		~float32 | ~float64 |

		// Other basic types
		~bool | ~string | time.Time |

		// Record type for nested structures
		Record |

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

// Nested adds a nested record field
func (tr *TypedRecord) Nested(key string, value Record) *TypedRecord {
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