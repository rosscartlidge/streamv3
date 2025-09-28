package streamv3

import (
	"cmp"
	"iter"
	"slices"
)

// ============================================================================
// FLUENT STREAMBUILDER API - ERGONOMIC ITERATOR CONSUMPTION
// ============================================================================

// ============================================================================
// CORE BUILDER TYPES
// ============================================================================

// StreamBuilder provides fluent API for simple iterator operations
type StreamBuilder[T any] struct {
	seq iter.Seq[T]
}

// StreamBuilderWithErrors provides fluent API with error handling
type StreamBuilderWithErrors[T any] struct {
	seq iter.Seq2[T, error]
}

// ============================================================================
// ENTRY POINTS - CREATING STREAMS
// ============================================================================

// From creates a StreamBuilder from a slice
func From[T any](data []T) *StreamBuilder[T] {
	return &StreamBuilder[T]{seq: slices.Values(data)}
}

// FromIter creates a StreamBuilder from an existing iterator
func FromIter[T any](seq iter.Seq[T]) *StreamBuilder[T] {
	return &StreamBuilder[T]{seq: seq}
}

// FromChannel creates a StreamBuilder from a channel
func FromChannel[T any](ch <-chan T) *StreamBuilder[T] {
	return &StreamBuilder[T]{
		seq: func(yield func(T) bool) {
			for v := range ch {
				if !yield(v) {
					return
				}
			}
		},
	}
}

// FromFunc creates a StreamBuilder from a generator function
func FromFunc[T any](fn func() (T, bool)) *StreamBuilder[T] {
	return &StreamBuilder[T]{
		seq: func(yield func(T) bool) {
			for {
				v, ok := fn()
				if !ok {
					return
				}
				if !yield(v) {
					return
				}
			}
		},
	}
}

// Range creates a StreamBuilder for integer sequences
func Range(start, end, step int) *StreamBuilder[int] {
	return &StreamBuilder[int]{
		seq: func(yield func(int) bool) {
			if step > 0 {
				for i := start; i < end; i += step {
					if !yield(i) {
						return
					}
				}
			} else if step < 0 {
				for i := start; i > end; i += step {
					if !yield(i) {
						return
					}
				}
			}
		},
	}
}

// ============================================================================
// ERROR-AWARE ENTRY POINTS
// ============================================================================

// FromSafe creates an error-aware StreamBuilder from a slice (never errors)
func FromSafe[T any](data []T) *StreamBuilderWithErrors[T] {
	return &StreamBuilderWithErrors[T]{seq: Safe(slices.Values(data))}
}

// FromIterSafe creates an error-aware StreamBuilder from an iter.Seq2
func FromIterSafe[T any](seq iter.Seq2[T, error]) *StreamBuilderWithErrors[T] {
	return &StreamBuilderWithErrors[T]{seq: seq}
}

// ============================================================================
// STREAMBUILDER CORE METHODS
// ============================================================================

// Apply applies any FilterSameType[T] to the stream
func (sb *StreamBuilder[T]) Apply(filter FilterSameType[T]) *StreamBuilder[T] {
	return &StreamBuilder[T]{seq: filter(sb.seq)}
}

// Safe converts to error-aware builder (never errors)
func (sb *StreamBuilder[T]) Safe() *StreamBuilderWithErrors[T] {
	return &StreamBuilderWithErrors[T]{seq: Safe(sb.seq)}
}

// ============================================================================
// GENERIC TRANSFORMATION HELPERS
// ============================================================================

// TransformTo applies any Filter[T, U] to change the stream type
func TransformTo[T, U any](sb *StreamBuilder[T], filter Filter[T, U]) *StreamBuilder[U] {
	return &StreamBuilder[U]{seq: filter(sb.seq)}
}

// TransformToWithErrors applies any FilterWithErrors[T, U] to change the stream type
func TransformToWithErrors[T, U any](sb *StreamBuilderWithErrors[T], filter FilterWithErrors[T, U]) *StreamBuilderWithErrors[U] {
	return &StreamBuilderWithErrors[U]{seq: filter(sb.seq)}
}

// MapTo transforms elements to a different type
func MapTo[T, U any](sb *StreamBuilder[T], fn func(T) U) *StreamBuilder[U] {
	return TransformTo(sb, Map(fn))
}

// FlatMapTo flattens nested iterators
func FlatMapTo[T, U any](sb *StreamBuilder[T], fn func(T) iter.Seq[U]) *StreamBuilder[U] {
	return TransformTo(sb, FlatMap(fn))
}

// MapToSafe transforms elements to a different type with error handling
func MapToSafe[T, U any](sb *StreamBuilderWithErrors[T], fn func(T) (U, error)) *StreamBuilderWithErrors[U] {
	return TransformToWithErrors(sb, MapSafe(fn))
}

// SortByKey sorts elements using a key extraction function
func SortByKey[T any, K cmp.Ordered](sb *StreamBuilder[T], keyFn func(T) K) *StreamBuilder[T] {
	return sb.Apply(SortBy(keyFn))
}

// DistinctByKey removes duplicates based on a key
func DistinctByKey[T any, K comparable](sb *StreamBuilder[T], keyFn func(T) K) *StreamBuilder[T] {
	return sb.Apply(DistinctBy(keyFn))
}

// WindowTo groups elements into fixed-size windows
func WindowTo[T any](sb *StreamBuilder[T], size int) *StreamBuilder[[]T] {
	return TransformTo(sb, Window[T](size))
}

// SortOrdered sorts elements that implement cmp.Ordered
func SortOrdered[T cmp.Ordered](sb *StreamBuilder[T]) *StreamBuilder[T] {
	return sb.Apply(Sort[T]())
}

// SortDescOrdered sorts elements in descending order for cmp.Ordered types
func SortDescOrdered[T cmp.Ordered](sb *StreamBuilder[T]) *StreamBuilder[T] {
	return sb.Apply(SortDesc[T]())
}

// DistinctComparable removes duplicate elements for comparable types
func DistinctComparable[T comparable](sb *StreamBuilder[T]) *StreamBuilder[T] {
	return sb.Apply(Distinct[T]())
}

// SumNumeric sums numeric elements for types that support addition
func SumNumeric[T Numeric](sb *StreamBuilder[T]) T {
	var sum T
	for v := range sb.seq {
		sum = sum + v
	}
	return sum
}

// MaxOrdered finds maximum element for cmp.Ordered types
func MaxOrdered[T cmp.Ordered](sb *StreamBuilder[T]) (T, bool) {
	var max T
	found := false
	for v := range sb.seq {
		if !found || v > max {
			max = v
			found = true
		}
	}
	return max, found
}

// MinOrdered finds minimum element for cmp.Ordered types
func MinOrdered[T cmp.Ordered](sb *StreamBuilder[T]) (T, bool) {
	var min T
	found := false
	for v := range sb.seq {
		if !found || v < min {
			min = v
			found = true
		}
	}
	return min, found
}

// ============================================================================
// STREAMBUILDER CONVENIENCE METHODS
// ============================================================================

// Where filters elements (SQL WHERE equivalent)
func (sb *StreamBuilder[T]) Where(predicate func(T) bool) *StreamBuilder[T] {
	return sb.Apply(Where(predicate))
}

// Note: Map and FlatMap with type parameters are implemented as standalone functions
// Use MapTo[U](sb, fn) and FlatMapTo[U](sb, fn) instead

// Limit restricts to first N elements (SQL LIMIT)
func (sb *StreamBuilder[T]) Limit(n int) *StreamBuilder[T] {
	return sb.Apply(Limit[T](n))
}

// Offset skips first N elements (SQL OFFSET)
func (sb *StreamBuilder[T]) Offset(n int) *StreamBuilder[T] {
	return sb.Apply(Offset[T](n))
}

// Note: Sort for cmp.Ordered types is implemented as a standalone function
// Use SortOrdered[T](sb) for types that implement cmp.Ordered

// Note: SortBy with type parameters is implemented as a standalone function
// Use SortByKey[T, K](sb, keyFn) instead

// Note: SortDesc for cmp.Ordered types is implemented as a standalone function
// Use SortDescOrdered[T](sb) for types that implement cmp.Ordered

// Note: Distinct for comparable types is implemented as a standalone function
// Use DistinctComparable[T](sb) for types that are comparable

// Note: DistinctBy with type parameters is implemented as a standalone function
// Use DistinctByKey[T, K](sb, keyFn) instead

// Reverse reverses element order
func (sb *StreamBuilder[T]) Reverse() *StreamBuilder[T] {
	return sb.Apply(Reverse[T]())
}

// Note: Window with type change is implemented as a standalone function
// Use WindowTo[T](sb, size) instead

// ============================================================================
// STREAMBUILDER WITH ERRORS METHODS
// ============================================================================

// Apply applies any FilterWithErrorsSameType[T] to the stream
func (sb *StreamBuilderWithErrors[T]) Apply(filter FilterWithErrorsSameType[T]) *StreamBuilderWithErrors[T] {
	return &StreamBuilderWithErrors[T]{seq: filter(sb.seq)}
}

// Note: Transform with type parameters is implemented as a standalone function
// Use TransformToWithErrors[T, U](sb, filter) instead

// Unsafe converts to simple builder (panics on errors)
func (sb *StreamBuilderWithErrors[T]) Unsafe() *StreamBuilder[T] {
	return &StreamBuilder[T]{seq: Unsafe(sb.seq)}
}

// IgnoreErrors converts to simple builder (ignores errors)
func (sb *StreamBuilderWithErrors[T]) IgnoreErrors() *StreamBuilder[T] {
	return &StreamBuilder[T]{seq: IgnoreErrors(sb.seq)}
}

// Error-aware convenience methods
func (sb *StreamBuilderWithErrors[T]) WhereSafe(predicate func(T) (bool, error)) *StreamBuilderWithErrors[T] {
	return sb.Apply(WhereSafe(predicate))
}

// Note: MapSafe with type parameters is implemented as a standalone function
// Use MapToSafe[T, U](sb, fn) instead

func (sb *StreamBuilderWithErrors[T]) LimitSafe(n int) *StreamBuilderWithErrors[T] {
	return sb.Apply(LimitSafe[T](n))
}

func (sb *StreamBuilderWithErrors[T]) OffsetSafe(n int) *StreamBuilderWithErrors[T] {
	return sb.Apply(OffsetSafe[T](n))
}

// ============================================================================
// FINALIZATION METHODS
// ============================================================================

// Iter returns the underlying iterator for range loops
func (sb *StreamBuilder[T]) Iter() iter.Seq[T] {
	return sb.seq
}

// Collect materializes all elements into a slice using standard library
func (sb *StreamBuilder[T]) Collect() []T {
	return slices.Collect(sb.seq)
}

// ForEach executes a function for each element
func (sb *StreamBuilder[T]) ForEach(fn func(T)) {
	for v := range sb.seq {
		fn(v)
	}
}

// Count returns the number of elements
func (sb *StreamBuilder[T]) Count() int {
	count := 0
	for range sb.seq {
		count++
	}
	return count
}

// First returns the first element (if any)
func (sb *StreamBuilder[T]) First() (T, bool) {
	for v := range sb.seq {
		return v, true
	}
	var zero T
	return zero, false
}

// Last returns the last element (requires full iteration)
func (sb *StreamBuilder[T]) Last() (T, bool) {
	var last T
	found := false
	for v := range sb.seq {
		last = v
		found = true
	}
	return last, found
}

// Any checks if any element satisfies the predicate
func (sb *StreamBuilder[T]) Any(predicate func(T) bool) bool {
	for v := range sb.seq {
		if predicate(v) {
			return true
		}
	}
	return false
}

// All checks if all elements satisfy the predicate
func (sb *StreamBuilder[T]) All(predicate func(T) bool) bool {
	for v := range sb.seq {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// ============================================================================
// ERROR-AWARE FINALIZATION METHODS
// ============================================================================

// Iter returns the underlying error-aware iterator
func (sb *StreamBuilderWithErrors[T]) Iter() iter.Seq2[T, error] {
	return sb.seq
}

// Collect materializes all elements, returning first error encountered
func (sb *StreamBuilderWithErrors[T]) Collect() ([]T, error) {
	var result []T
	for v, err := range sb.seq {
		if err != nil {
			return result, err
		}
		result = append(result, v)
	}
	return result, nil
}

// ForEach executes a function for each element, stopping on first error
func (sb *StreamBuilderWithErrors[T]) ForEach(fn func(T)) error {
	for v, err := range sb.seq {
		if err != nil {
			return err
		}
		fn(v)
	}
	return nil
}

// Count returns the number of non-error elements
func (sb *StreamBuilderWithErrors[T]) Count() (int, error) {
	count := 0
	for _, err := range sb.seq {
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// ============================================================================
// AGGREGATION METHODS
// ============================================================================

// Note: Sum for numeric types is implemented as a standalone function
// Use SumNumeric[T](sb) for types that support addition

// Note: Max/Min for ordered types are implemented as standalone functions
// Use MaxOrdered[T](sb) and MinOrdered[T](sb) for types that implement cmp.Ordered

// ============================================================================
// SQL-STYLE ERROR-AWARE EXTENSIONS
// ============================================================================

// Error-aware join operations would need to be implemented if we want to support
// joins on error-aware streams. For now, users can convert to unsafe streams,
// perform joins, then convert back to error-aware if needed.

// Note: Record-specific methods are in sql.go file
// These would be duplicates and should be removed or implemented differently