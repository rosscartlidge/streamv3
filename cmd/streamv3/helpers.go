package main

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/rosscartlidge/streamv3"
)

// Helper functions for command handlers

func extractNumeric(val any) float64 {
	switch v := val.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	case string:
		// For strings, use 0 (they'll maintain relative order)
		return 0
	default:
		return 0
	}
}

// applyOperator applies a comparison operator for where command
func applyOperator(fieldValue any, op string, compareValue string) bool {
	switch op {
	case "eq":
		return compareEqual(fieldValue, compareValue)
	case "ne":
		return !compareEqual(fieldValue, compareValue)
	case "gt":
		return compareGreater(fieldValue, compareValue)
	case "ge":
		return compareGreater(fieldValue, compareValue) || compareEqual(fieldValue, compareValue)
	case "lt":
		return compareLess(fieldValue, compareValue)
	case "le":
		return compareLess(fieldValue, compareValue) || compareEqual(fieldValue, compareValue)
	case "contains":
		return compareContains(fieldValue, compareValue)
	case "startswith":
		return compareStartsWith(fieldValue, compareValue)
	case "endswith":
		return compareEndsWith(fieldValue, compareValue)
	case "pattern", "regexp", "regex":
		return comparePattern(fieldValue, compareValue)
	default:
		return false
	}
}

func compareEqual(fieldValue any, compareValue string) bool {
	switch v := fieldValue.(type) {
	case string:
		return v == compareValue
	case int64:
		if num, err := strconv.ParseInt(compareValue, 10, 64); err == nil {
			return v == num
		}
	case float64:
		if num, err := strconv.ParseFloat(compareValue, 64); err == nil {
			return v == num
		}
	case bool:
		if b, err := strconv.ParseBool(compareValue); err == nil {
			return v == b
		}
	}
	return fmt.Sprintf("%v", fieldValue) == compareValue
}

func compareGreater(fieldValue any, compareValue string) bool {
	switch v := fieldValue.(type) {
	case int64:
		if num, err := strconv.ParseInt(compareValue, 10, 64); err == nil {
			return v > num
		}
	case float64:
		if num, err := strconv.ParseFloat(compareValue, 64); err == nil {
			return v > num
		}
	case string:
		return v > compareValue
	}
	return false
}

func compareLess(fieldValue any, compareValue string) bool {
	switch v := fieldValue.(type) {
	case int64:
		if num, err := strconv.ParseInt(compareValue, 10, 64); err == nil {
			return v < num
		}
	case float64:
		if num, err := strconv.ParseFloat(compareValue, 64); err == nil {
			return v < num
		}
	case string:
		return v < compareValue
	}
	return false
}

func compareContains(fieldValue any, compareValue string) bool {
	if str, ok := fieldValue.(string); ok {
		return contains(str, compareValue)
	}
	return false
}

func compareStartsWith(fieldValue any, compareValue string) bool {
	if str, ok := fieldValue.(string); ok {
		return len(str) >= len(compareValue) && str[:len(compareValue)] == compareValue
	}
	return false
}

func compareEndsWith(fieldValue any, compareValue string) bool {
	if str, ok := fieldValue.(string); ok {
		return len(str) >= len(compareValue) && str[len(str)-len(compareValue):] == compareValue
	}
	return false
}

func comparePattern(fieldValue any, pattern string) bool {
	if str, ok := fieldValue.(string); ok {
		matched, err := regexp.MatchString(pattern, str)
		if err != nil {
			return false
		}
		return matched
	}
	return false
}

func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// buildAggregator builds a StreamV3 AggregateFunc from a spec (for group-by command)
func buildAggregator(function, field string) (streamv3.AggregateFunc, error) {
	switch function {
	case "count":
		return streamv3.Count(), nil
	case "sum":
		return streamv3.Sum(field), nil
	case "avg":
		return streamv3.Avg(field), nil
	case "min":
		return streamv3.Min[float64](field), nil
	case "max":
		return streamv3.Max[float64](field), nil
	default:
		return nil, fmt.Errorf("unknown aggregation function: %s", function)
	}
}

// unionRecordToKey converts a record to a string key for deduplication (for union command)
func unionRecordToKey(r streamv3.Record) string {
	// Use JSON representation as unique key
	return fmt.Sprintf("%v", r)
}
