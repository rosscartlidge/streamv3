package streamv3

import (
	"cmp"
	"fmt"
	"iter"
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
func InnerJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) FilterSameType[Record] {
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
						for k, v := range left {
							joined[k] = v
						}
						// Copy right record (with potential conflicts)
						for k, v := range right {
							joined[k] = v
						}
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
func LeftJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) FilterSameType[Record] {
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
						for k, v := range left {
							joined[k] = v
						}
						// Copy right record
						for k, v := range right {
							joined[k] = v
						}
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
func RightJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) FilterSameType[Record] {
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
						for k, v := range left {
							joined[k] = v
						}
						// Copy right record
						for k, v := range right {
							joined[k] = v
						}
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
func FullJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) FilterSameType[Record] {
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
						for k, v := range left {
							joined[k] = v
						}
						// Copy right record
						for k, v := range right {
							joined[k] = v
						}
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

// GroupedRecord represents a group of records with a key
type GroupedRecord struct {
	Key     any
	Records []Record
}

// GroupBy groups records by a key extraction function
func GroupBy[K comparable](keyFn func(Record) K) Filter[Record, GroupedRecord] {
	return func(input iter.Seq[Record]) iter.Seq[GroupedRecord] {
		return func(yield func(GroupedRecord) bool) {
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

			// Yield groups in order of first appearance
			for _, key := range keys {
				if !yield(GroupedRecord{
					Key:     key,
					Records: groups[key],
				}) {
					return
				}
			}
		}
	}
}

// GroupByFields groups records by specified field values
func GroupByFields(fields ...string) Filter[Record, GroupedRecord] {
	return GroupBy(func(r Record) string {
		var keyParts []string
		for _, field := range fields {
			if val, exists := r[field]; exists {
				keyParts = append(keyParts, fmt.Sprintf("%v", val))
			} else {
				keyParts = append(keyParts, "<nil>")
			}
		}
		return fmt.Sprintf("[%s]", strings.Join(keyParts, ","))
	})
}

// ============================================================================
// AGGREGATION OPERATIONS
// ============================================================================

// AggregateFunc defines an aggregation function over a group of records
type AggregateFunc func([]Record) any

// Aggregate applies aggregation functions to grouped records
func Aggregate(aggregations map[string]AggregateFunc) Filter[GroupedRecord, Record] {
	return func(input iter.Seq[GroupedRecord]) iter.Seq[Record] {
		return func(yield func(Record) bool) {
			for group := range input {
				result := make(Record)

				// Add the group key
				result["group_key"] = group.Key

				// Apply all aggregation functions
				for name, aggFn := range aggregations {
					result[name] = aggFn(group.Records)
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
			if val, exists := record[field]; exists {
				switch v := val.(type) {
				case int:
					sum += float64(v)
				case int32:
					sum += float64(v)
				case int64:
					sum += float64(v)
				case float32:
					sum += float64(v)
				case float64:
					sum += v
				}
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
			if val, exists := record[field]; exists {
				switch v := val.(type) {
				case int:
					sum += float64(v)
					count++
				case int32:
					sum += float64(v)
					count++
				case int64:
					sum += float64(v)
					count++
				case float32:
					sum += float64(v)
					count++
				case float64:
					sum += v
					count++
				}
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
			if val, exists := record[field]; exists {
				if v, ok := val.(T); ok {
					if !found || v < min {
						min = v
						found = true
					}
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
			if val, exists := record[field]; exists {
				if v, ok := val.(T); ok {
					if !found || v > max {
						max = v
						found = true
					}
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

