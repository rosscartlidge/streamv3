package streamv3

import "iter"

// Simple types for I/O compatibility - no fluent methods
// These exist only to support existing I/O function signatures

// Stream represents a simple iterator sequence
type Stream[T any] struct {
	seq iter.Seq[T]
}

// StreamWithErrors represents an error-aware iterator sequence
type StreamWithErrors[T any] struct {
	seq iter.Seq2[T, error]
}

// Iter returns the underlying iterator
func (s *Stream[T]) Iter() iter.Seq[T] {
	return s.seq
}

// Iter returns the underlying error-aware iterator
func (s *StreamWithErrors[T]) Iter() iter.Seq2[T, error] {
	return s.seq
}

// Collect materializes all elements into a slice
func (s *Stream[T]) Collect() []T {
	var result []T
	for item := range s.seq {
		result = append(result, item)
	}
	return result
}

// Collect materializes all elements, returning first error encountered
func (s *StreamWithErrors[T]) Collect() ([]T, error) {
	var result []T
	for item, err := range s.seq {
		if err != nil {
			return result, err
		}
		result = append(result, item)
	}
	return result, nil
}

// Tee splits the stream into multiple identical streams for parallel consumption
func (s *Stream[T]) Tee(n int) []*Stream[T] {
	streams := Tee(s.seq, n)
	result := make([]*Stream[T], len(streams))
	for i, stream := range streams {
		result[i] = &Stream[T]{seq: stream}
	}
	return result
}

// LazyTee splits the stream into multiple identical streams without buffering for infinite streams
func (s *Stream[T]) LazyTee(n int) []*Stream[T] {
	streams := LazyTee(s.seq, n)
	result := make([]*Stream[T], len(streams))
	for i, stream := range streams {
		result[i] = &Stream[T]{seq: stream}
	}
	return result
}

// From creates a Stream from a slice
func From[T any](data []T) *Stream[T] {
	return &Stream[T]{seq: func(yield func(T) bool) {
		for _, item := range data {
			if !yield(item) {
				return
			}
		}
	}}
}