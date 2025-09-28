package streamv3

import (
	"iter"
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
func (tr *TypedRecord) Set(key string, value any) *TypedRecord {
	tr.data[key] = value
	return tr
}

// String adds a string field
func (tr *TypedRecord) String(key, value string) *TypedRecord {
	tr.data[key] = value
	return tr
}

// Int adds an integer field
func (tr *TypedRecord) Int(key string, value int64) *TypedRecord {
	tr.data[key] = value
	return tr
}

// Float adds a float field
func (tr *TypedRecord) Float(key string, value float64) *TypedRecord {
	tr.data[key] = value
	return tr
}

// Bool adds a boolean field
func (tr *TypedRecord) Bool(key string, value bool) *TypedRecord {
	tr.data[key] = value
	return tr
}

// Time adds a time field
func (tr *TypedRecord) Time(key string, value time.Time) *TypedRecord {
	tr.data[key] = value
	return tr
}

// Nested adds a nested record field
func (tr *TypedRecord) Nested(key string, value Record) *TypedRecord {
	tr.data[key] = value
	return tr
}

// Build finalizes the record construction
func (tr *TypedRecord) Build() Record {
	return Record(tr.data)
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