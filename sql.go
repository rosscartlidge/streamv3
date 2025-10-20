package streamv3

import (
	"cmp"
	"fmt"
	"iter"
	"maps"
	"strings"
)

// ============================================================================
// SQL-STYLE OPERATIONS - JOIN, GROUPBY, AGGREGATION
// ============================================================================

// ============================================================================
// JOIN OPERATIONS
// ============================================================================

// JoinPredicate defines the condition for joining two records
type JoinPredicate func(left, right Record) bool

// InnerJoin performs an inner join between two record streams
func InnerJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) Filter[Record, Record] {
	return func(leftSeq iter.Seq[Record]) iter.Seq[Record] {
		return func(yield func(Record) bool) {
			// Materialize right side for multiple iterations
			var rightRecords []Record
			for r := range rightSeq {
				rightRecords = append(rightRecords, r)
			}

			for left := range leftSeq {
				for _, right := range rightRecords {
					if predicate(left, right) {
						joined := make(Record)
						// Copy left record
						maps.Copy(joined, left)
						// Copy right record (with potential conflicts)
						maps.Copy(joined, right)
						if !yield(joined) {
							return
						}
					}
				}
			}
		}
	}
}

// LeftJoin performs a left join between two record streams
func LeftJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) Filter[Record, Record] {
	return func(leftSeq iter.Seq[Record]) iter.Seq[Record] {
		return func(yield func(Record) bool) {
			// Materialize right side for multiple iterations
			var rightRecords []Record
			for r := range rightSeq {
				rightRecords = append(rightRecords, r)
			}

			for left := range leftSeq {
				matched := false
				for _, right := range rightRecords {
					if predicate(left, right) {
						joined := make(Record)
						// Copy left record
						maps.Copy(joined, left)
						// Copy right record
						maps.Copy(joined, right)
						if !yield(joined) {
							return
						}
						matched = true
					}
				}
				// If no match, yield left record only
				if !matched {
					if !yield(left) {
						return
					}
				}
			}
		}
	}
}

// RightJoin performs a right join between two record streams
func RightJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) Filter[Record, Record] {
	return func(leftSeq iter.Seq[Record]) iter.Seq[Record] {
		return func(yield func(Record) bool) {
			// Materialize both sides
			var leftRecords []Record
			for l := range leftSeq {
				leftRecords = append(leftRecords, l)
			}
			var rightRecords []Record
			for r := range rightSeq {
				rightRecords = append(rightRecords, r)
			}

			// Track which right records were matched
			matched := make([]bool, len(rightRecords))

			// First pass: yield matched records
			for _, left := range leftRecords {
				for i, right := range rightRecords {
					if predicate(left, right) {
						joined := make(Record)
						// Copy left record
						maps.Copy(joined, left)
						// Copy right record
						maps.Copy(joined, right)
						if !yield(joined) {
							return
						}
						matched[i] = true
					}
				}
			}

			// Second pass: yield unmatched right records
			for i, right := range rightRecords {
				if !matched[i] {
					if !yield(right) {
						return
					}
				}
			}
		}
	}
}

// FullJoin performs a full outer join between two record streams
func FullJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) Filter[Record, Record] {
	return func(leftSeq iter.Seq[Record]) iter.Seq[Record] {
		return func(yield func(Record) bool) {
			// Materialize both sides
			var leftRecords []Record
			for l := range leftSeq {
				leftRecords = append(leftRecords, l)
			}
			var rightRecords []Record
			for r := range rightSeq {
				rightRecords = append(rightRecords, r)
			}

			// Track which records were matched
			leftMatched := make([]bool, len(leftRecords))
			rightMatched := make([]bool, len(rightRecords))

			// First pass: yield matched records
			for i, left := range leftRecords {
				for j, right := range rightRecords {
					if predicate(left, right) {
						joined := make(Record)
						// Copy left record
						maps.Copy(joined, left)
						// Copy right record
						maps.Copy(joined, right)
						if !yield(joined) {
							return
						}
						leftMatched[i] = true
						rightMatched[j] = true
					}
				}
			}

			// Second pass: yield unmatched left records
			for i, left := range leftRecords {
				if !leftMatched[i] {
					if !yield(left) {
						return
					}
				}
			}

			// Third pass: yield unmatched right records
			for j, right := range rightRecords {
				if !rightMatched[j] {
					if !yield(right) {
						return
					}
				}
			}
		}
	}
}

// ============================================================================
// JOIN HELPER FUNCTIONS
// ============================================================================

// OnFields creates a join predicate that matches records on specified fields
func OnFields(fields ...string) JoinPredicate {
	return func(left, right Record) bool {
		for _, field := range fields {
			leftVal, leftExists := left[field]
			rightVal, rightExists := right[field]
			if !leftExists || !rightExists || leftVal != rightVal {
				return false
			}
		}
		return true
	}
}

// OnCondition creates a join predicate from a custom condition function
func OnCondition(condition func(left, right Record) bool) JoinPredicate {
	return condition
}

// ============================================================================
// GROUPBY OPERATIONS
// ============================================================================

// GroupBy groups records by a key extraction function
func GroupBy[K comparable](sequenceField string, keyField string, keyFn func(Record) K) Filter[Record, Record] {
	return func(input iter.Seq[Record]) iter.Seq[Record] {
		return func(yield func(Record) bool) {
			groups := make(map[K][]Record)
			var keys []K

			// Collect all records into groups
			for record := range input {
				key := keyFn(record)
				if _, exists := groups[key]; !exists {
					keys = append(keys, key)
				}
				groups[key] = append(groups[key], record)
			}

			// Yield records with key field + sequence field
			for _, key := range keys {
				result := make(Record)

				// Set the key field
				result[keyField] = key

				// Add the sequence of group members as an iter.Seq[Record]
				groupRecords := groups[key]
				result[sequenceField] = func() iter.Seq[Record] {
					return func(yield func(Record) bool) {
						for _, record := range groupRecords {
							if !yield(record) {
								return
							}
						}
					}
				}()

				if !yield(result) {
					return
				}
			}
		}
	}
}

// GroupByFields groups records by specified field values
// Returns Records with grouping fields + a sequence field containing group members
func GroupByFields(sequenceField string, fields ...string) Filter[Record, Record] {
	return func(input iter.Seq[Record]) iter.Seq[Record] {
		return func(yield func(Record) bool) {
			groups := make(map[string][]Record)
			groupFields := make(map[string]Record)
			var keys []string

			// Collect all records into groups
			for record := range input {
				var keyParts []string
				groupingFields := make(Record)
				hasComplexField := false

				for _, field := range fields {
					if val, exists := record[field]; exists {
						// Validate that the field value is simple (no iter.Seq or Record)
						if !isSimpleValue(val) {
							// Skip this entire record if any grouping field is complex
							hasComplexField = true
							break
						}
						keyParts = append(keyParts, fmt.Sprintf("%v", val))
						groupingFields[field] = val
					} else {
						keyParts = append(keyParts, "<nil>")
						groupingFields[field] = nil
					}
				}

				// Skip records with complex grouping field values
				if hasComplexField {
					continue
				}

				key := fmt.Sprintf("[%s]", strings.Join(keyParts, ","))
				if _, exists := groups[key]; !exists {
					keys = append(keys, key)
					groupFields[key] = groupingFields
				}
				groups[key] = append(groups[key], record)
			}

			// Yield records with grouping fields + sequence field
			for _, key := range keys {
				result := make(Record)

				// Copy the grouping field values
				maps.Copy(result, groupFields[key])

				// Add the sequence of group members as an iter.Seq[Record]
				groupRecords := groups[key]
				result[sequenceField] = func() iter.Seq[Record] {
					return func(yield func(Record) bool) {
						for _, record := range groupRecords {
							if !yield(record) {
								return
							}
						}
					}
				}()

				if !yield(result) {
					return
				}
			}
		}
	}
}

// ============================================================================
// AGGREGATION OPERATIONS
// ============================================================================

// AggregateFunc defines an aggregation function over a group of records
type AggregateFunc func([]Record) any

// Aggregate applies aggregation functions to records containing sequence fields
func Aggregate(sequenceField string, aggregations map[string]AggregateFunc) Filter[Record, Record] {
	return func(input iter.Seq[Record]) iter.Seq[Record] {
		return func(yield func(Record) bool) {
			for record := range input {
				result := make(Record)

				// Copy all fields except the sequence field
				for field, value := range record {
					if field != sequenceField {
						result[field] = value
					}
				}

				// Extract the sequence from the specified field
				if seqValue, exists := record[sequenceField]; exists {
					if seq, ok := seqValue.(iter.Seq[Record]); ok {
						// Materialize the sequence for aggregation functions
						var records []Record
						for r := range seq {
							records = append(records, r)
						}

						// Apply all aggregation functions
						for name, aggFn := range aggregations {
							result[name] = aggFn(records)
						}
					}
				}

				if !yield(result) {
					return
				}
			}
		}
	}
}

// ============================================================================
// COMMON AGGREGATION FUNCTIONS
// ============================================================================

// Count returns the number of records in a group
func Count() AggregateFunc {
	return func(records []Record) any {
		return int64(len(records))
	}
}

// Sum sums numeric values from a field across all records
func Sum(field string) AggregateFunc {
	return func(records []Record) any {
		var sum float64
		for _, record := range records {
			// Use type-safe Get with automatic conversion to float64
			if value, ok := Get[float64](record, field); ok {
				sum += value
			}
		}
		return sum
	}
}

// Avg calculates the average of numeric values from a field
func Avg(field string) AggregateFunc {
	return func(records []Record) any {
		var sum float64
		var count int64
		for _, record := range records {
			// Use type-safe Get with automatic conversion to float64
			if value, ok := Get[float64](record, field); ok {
				sum += value
				count++
			}
		}
		if count == 0 {
			return 0.0
		}
		return sum / float64(count)
	}
}

// Min finds the minimum value from a field across all records
func Min[T cmp.Ordered](field string) AggregateFunc {
	return func(records []Record) any {
		if len(records) == 0 {
			var zero T
			return zero
		}

		var min T
		found := false
		for _, record := range records {
			// Use type-safe Get with automatic conversion
			if value, ok := Get[T](record, field); ok {
				if !found || value < min {
					min = value
					found = true
				}
			}
		}
		return min
	}
}

// Max finds the maximum value from a field across all records
func Max[T cmp.Ordered](field string) AggregateFunc {
	return func(records []Record) any {
		if len(records) == 0 {
			var zero T
			return zero
		}

		var max T
		found := false
		for _, record := range records {
			// Use type-safe Get with automatic conversion
			if value, ok := Get[T](record, field); ok {
				if !found || value > max {
					max = value
					found = true
				}
			}
		}
		return max
	}
}

// First returns the first non-nil value from a field
func First(field string) AggregateFunc {
	return func(records []Record) any {
		for _, record := range records {
			if val, exists := record[field]; exists {
				return val
			}
		}
		return nil
	}
}

// Last returns the last non-nil value from a field
func Last(field string) AggregateFunc {
	return func(records []Record) any {
		var lastVal any
		for _, record := range records {
			if val, exists := record[field]; exists {
				lastVal = val
			}
		}
		return lastVal
	}
}

// Collect gathers all values from a field into a slice
func Collect(field string) AggregateFunc {
	return func(records []Record) any {
		var values []any
		for _, record := range records {
			if val, exists := record[field]; exists {
				values = append(values, val)
			}
		}
		return values
	}
}

