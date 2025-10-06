package streamv3

import (
	"cmp"
	"iter"
	"slices"
)

// ============================================================================
// CORE STREAM OPERATIONS - FUNCTIONAL FILTER API
// ============================================================================

// ============================================================================
// TRANSFORM OPERATIONS
// ============================================================================

// Map transforms each element using the provided function
func Map[T, U any](fn func(T) U) Filter[T, U] {
	return func(input iter.Seq[T]) iter.Seq[U] {
		return func(yield func(U) bool) {
			for v := range input {
				if !yield(fn(v)) {
					return
				}
			}
		}
	}
}

// MapSafe transforms each element with error handling
func MapSafe[T, U any](fn func(T) (U, error)) FilterWithErrors[T, U] {
	return func(input iter.Seq2[T, error]) iter.Seq2[U, error] {
		return func(yield func(U, error) bool) {
			for v, err := range input {
				if err != nil {
					var zero U
					if !yield(zero, err) {
						return
					}
					continue
				}
				result, mapErr := fn(v)
				if !yield(result, mapErr) {
					return
				}
			}
		}
	}
}

// FlatMap flattens nested iterators
func FlatMap[T, U any](fn func(T) iter.Seq[U]) Filter[T, U] {
	return func(input iter.Seq[T]) iter.Seq[U] {
		return func(yield func(U) bool) {
			for v := range input {
				for u := range fn(v) {
					if !yield(u) {
						return
					}
				}
			}
		}
	}
}

// ============================================================================
// FILTER OPERATIONS
// ============================================================================

// Where filters elements based on a predicate (equivalent to SQL WHERE)
func Where[T any](predicate func(T) bool) FilterSameType[T] {
	return func(input iter.Seq[T]) iter.Seq[T] {
		return func(yield func(T) bool) {
			for v := range input {
				if predicate(v) && !yield(v) {
					return
				}
			}
		}
	}
}

// WhereSafe filters elements with error handling
func WhereSafe[T any](predicate func(T) (bool, error)) FilterWithErrorsSameType[T] {
	return func(input iter.Seq2[T, error]) iter.Seq2[T, error] {
		return func(yield func(T, error) bool) {
			for v, err := range input {
				if err != nil {
					if !yield(v, err) {
						return
					}
					continue
				}
				include, predErr := predicate(v)
				if predErr != nil {
					if !yield(v, predErr) {
						return
					}
					continue
				}
				if include && !yield(v, nil) {
					return
				}
			}
		}
	}
}

// ============================================================================
// LIMITING OPERATIONS - SQL-STYLE
// ============================================================================

// Limit restricts iterator to first N elements (equivalent to SQL LIMIT)
func Limit[T any](n int) FilterSameType[T] {
	return func(input iter.Seq[T]) iter.Seq[T] {
		return func(yield func(T) bool) {
			count := 0
			for v := range input {
				if count >= n {
					return
				}
				if !yield(v) {
					return
				}
				count++
			}
		}
	}
}

// LimitSafe restricts iterator with error handling
func LimitSafe[T any](n int) FilterWithErrorsSameType[T] {
	return func(input iter.Seq2[T, error]) iter.Seq2[T, error] {
		return func(yield func(T, error) bool) {
			count := 0
			for v, err := range input {
				if count >= n {
					return
				}
				if !yield(v, err) {
					return
				}
				if err == nil {
					count++
				}
			}
		}
	}
}

// Offset skips first N elements (equivalent to SQL OFFSET)
func Offset[T any](n int) FilterSameType[T] {
	return func(input iter.Seq[T]) iter.Seq[T] {
		return func(yield func(T) bool) {
			skipped := 0
			for v := range input {
				if skipped < n {
					skipped++
					continue
				}
				if !yield(v) {
					return
				}
			}
		}
	}
}

// OffsetSafe skips first N elements with error handling
func OffsetSafe[T any](n int) FilterWithErrorsSameType[T] {
	return func(input iter.Seq2[T, error]) iter.Seq2[T, error] {
		return func(yield func(T, error) bool) {
			skipped := 0
			for v, err := range input {
				if err != nil {
					if !yield(v, err) {
						return
					}
					continue
				}
				if skipped < n {
					skipped++
					continue
				}
				if !yield(v, nil) {
					return
				}
			}
		}
	}
}

// ============================================================================
// ORDERING OPERATIONS
// ============================================================================

// Sort sorts elements in ascending order using standard library
func Sort[T cmp.Ordered]() FilterSameType[T] {
	return func(input iter.Seq[T]) iter.Seq[T] {
		return slices.Values(slices.Sorted(input))
	}
}

// SortBy sorts elements using a key extraction function
func SortBy[T any, K cmp.Ordered](keyFn func(T) K) FilterSameType[T] {
	return func(input iter.Seq[T]) iter.Seq[T] {
		return slices.Values(slices.SortedFunc(input, func(a, b T) int {
			return cmp.Compare(keyFn(a), keyFn(b))
		}))
	}
}

// SortDesc sorts elements in descending order
func SortDesc[T cmp.Ordered]() FilterSameType[T] {
	return func(input iter.Seq[T]) iter.Seq[T] {
		return slices.Values(slices.SortedFunc(input, func(a, b T) int {
			return cmp.Compare(b, a) // Reverse comparison
		}))
	}
}

// ============================================================================
// UTILITY OPERATIONS
// ============================================================================

// Distinct removes duplicate elements (requires comparable type)
func Distinct[T comparable]() FilterSameType[T] {
	return func(input iter.Seq[T]) iter.Seq[T] {
		return func(yield func(T) bool) {
			seen := make(map[T]bool)
			for v := range input {
				if !seen[v] {
					seen[v] = true
					if !yield(v) {
						return
					}
				}
			}
		}
	}
}

// DistinctBy removes duplicates based on a key extraction function
func DistinctBy[T any, K comparable](keyFn func(T) K) FilterSameType[T] {
	return func(input iter.Seq[T]) iter.Seq[T] {
		return func(yield func(T) bool) {
			seen := make(map[K]bool)
			for v := range input {
				key := keyFn(v)
				if !seen[key] {
					seen[key] = true
					if !yield(v) {
						return
					}
				}
			}
		}
	}
}

// Reverse reverses the order of elements
func Reverse[T any]() FilterSameType[T] {
	return func(input iter.Seq[T]) iter.Seq[T] {
		return func(yield func(T) bool) {
			collected := slices.Collect(input)
			// Yield in reverse order
			for i := len(collected) - 1; i >= 0; i-- {
				if !yield(collected[i]) {
					return
				}
			}
		}
	}
}

// ============================================================================
// WINDOW OPERATIONS
// ============================================================================

// Window groups elements into fixed-size windows
func Window[T any](size int) Filter[T, []T] {
	return func(input iter.Seq[T]) iter.Seq[[]T] {
		return func(yield func([]T) bool) {
			window := make([]T, 0, size)
			for v := range input {
				window = append(window, v)
				if len(window) == size {
					if !yield(slices.Clone(window)) {
						return
					}
					window = window[:0] // Reset window
				}
			}
			// Yield final partial window if any
			if len(window) > 0 {
				yield(window)
			}
		}
	}
}

// SlidingWindow creates overlapping windows
func SlidingWindow[T any](size, step int) Filter[T, []T] {
	return func(input iter.Seq[T]) iter.Seq[[]T] {
		return func(yield func([]T) bool) {
			buffer := make([]T, 0, size)
			count := 0

			for v := range input {
				buffer = append(buffer, v)
				count++

				// Emit window when we have enough elements
				if len(buffer) == size {
					if !yield(slices.Clone(buffer)) {
						return
					}

					// Slide the window
					if step >= size {
						buffer = buffer[:0]
						// Skip elements if step > size
						for i := 1; i < step && count < step; i++ {
							count++
						}
					} else {
						// Shift buffer by step
						copy(buffer, buffer[step:])
						buffer = buffer[:len(buffer)-step]
					}
				}
			}
		}
	}
}

// ============================================================================
// STREAM UTILITIES
// ============================================================================

// Tee splits a stream into multiple identical streams for parallel consumption.
// Returns a slice of iterators that will each yield the same sequence of values.
// The source stream is fully consumed and buffered to enable multiple iterations.
func Tee[T any](input iter.Seq[T], n int) []iter.Seq[T] {
	if n <= 0 {
		return nil
	}

	// Collect all values from the source stream
	var values []T
	for v := range input {
		values = append(values, v)
	}

	// Create n identical iterators over the collected values
	streams := make([]iter.Seq[T], n)
	for i := 0; i < n; i++ {
		streams[i] = slices.Values(values)
	}

	return streams
}