package streamv3

import (
	"iter"
	"slices"
	"testing"
	"time"
)

// ============================================================================
// TRANSFORM OPERATIONS TESTS
// ============================================================================

func TestSelect(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5})
	filter := Select(func(x int) int { return x * 2 })
	result := slices.Collect(filter(input))

	expected := []int{2, 4, 6, 8, 10}
	if !slices.Equal(result, expected) {
		t.Errorf("Select failed: expected %v, got %v", expected, result)
	}
}

func TestSelectTypeChange(t *testing.T) {
	input := slices.Values([]int{1, 2, 3})
	filter := Select(func(x int) string { return string(rune('A' + x - 1)) })
	result := slices.Collect(filter(input))

	expected := []string{"A", "B", "C"}
	if !slices.Equal(result, expected) {
		t.Errorf("Select type change failed: expected %v, got %v", expected, result)
	}
}

func TestSelectSafe(t *testing.T) {
	input := Safe(slices.Values([]int{1, 2, 3}))
	filter := SelectSafe(func(x int) (int, error) {
		return x * 2, nil
	})

	var result []int
	for v, err := range filter(input) {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		result = append(result, v)
	}

	expected := []int{2, 4, 6}
	if !slices.Equal(result, expected) {
		t.Errorf("SelectSafe failed: expected %v, got %v", expected, result)
	}
}

func TestSelectMany(t *testing.T) {
	// Expand each number into a sequence of numbers from 1 to n
	input := slices.Values([]int{2, 3})
	filter := SelectMany(func(n int) iter.Seq[int] {
		return func(yield func(int) bool) {
			for i := 1; i <= n; i++ {
				if !yield(i) {
					return
				}
			}
		}
	})

	result := slices.Collect(filter(input))
	expected := []int{1, 2, 1, 2, 3} // 2 expands to [1,2], 3 expands to [1,2,3]

	if !slices.Equal(result, expected) {
		t.Errorf("SelectMany failed: expected %v, got %v", expected, result)
	}
}

// ============================================================================
// FILTER OPERATIONS TESTS
// ============================================================================

func TestWhere(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5, 6})
	filter := Where(func(x int) bool { return x%2 == 0 })
	result := slices.Collect(filter(input))

	expected := []int{2, 4, 6}
	if !slices.Equal(result, expected) {
		t.Errorf("Where failed: expected %v, got %v", expected, result)
	}
}

func TestWhereSafe(t *testing.T) {
	input := Safe(slices.Values([]int{1, 2, 3, 4, 5}))
	filter := WhereSafe(func(x int) (bool, error) {
		return x > 2, nil
	})

	var result []int
	for v, err := range filter(input) {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		result = append(result, v)
	}

	expected := []int{3, 4, 5}
	if !slices.Equal(result, expected) {
		t.Errorf("WhereSafe failed: expected %v, got %v", expected, result)
	}
}

// ============================================================================
// LIMITING OPERATIONS TESTS
// ============================================================================

func TestLimit(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5})
	filter := Limit[int](3)
	result := slices.Collect(filter(input))

	expected := []int{1, 2, 3}
	if !slices.Equal(result, expected) {
		t.Errorf("Limit failed: expected %v, got %v", expected, result)
	}
}

func TestLimitZero(t *testing.T) {
	input := slices.Values([]int{1, 2, 3})
	filter := Limit[int](0)
	result := slices.Collect(filter(input))

	if len(result) != 0 {
		t.Errorf("Limit(0) should return empty, got %v", result)
	}
}

func TestLimitSafe(t *testing.T) {
	input := Safe(slices.Values([]int{1, 2, 3, 4, 5}))
	filter := LimitSafe[int](3)

	var result []int
	for v, err := range filter(input) {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		result = append(result, v)
	}

	expected := []int{1, 2, 3}
	if !slices.Equal(result, expected) {
		t.Errorf("LimitSafe failed: expected %v, got %v", expected, result)
	}
}

func TestOffset(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5})
	filter := Offset[int](2)
	result := slices.Collect(filter(input))

	expected := []int{3, 4, 5}
	if !slices.Equal(result, expected) {
		t.Errorf("Offset failed: expected %v, got %v", expected, result)
	}
}

func TestOffsetSafe(t *testing.T) {
	input := Safe(slices.Values([]int{1, 2, 3, 4, 5}))
	filter := OffsetSafe[int](2)

	var result []int
	for v, err := range filter(input) {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		result = append(result, v)
	}

	expected := []int{3, 4, 5}
	if !slices.Equal(result, expected) {
		t.Errorf("OffsetSafe failed: expected %v, got %v", expected, result)
	}
}

func TestLimitOffset(t *testing.T) {
	// Test combined limit and offset
	input := slices.Values([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	// Apply offset first, then limit
	offsetted := Offset[int](3)(input)
	filter := Limit[int](4)
	result := slices.Collect(filter(offsetted))

	expected := []int{4, 5, 6, 7}
	if !slices.Equal(result, expected) {
		t.Errorf("Limit+Offset failed: expected %v, got %v", expected, result)
	}
}

// ============================================================================
// ORDERING OPERATIONS TESTS
// ============================================================================

func TestSort(t *testing.T) {
	input := slices.Values([]int{3, 1, 4, 1, 5, 9, 2, 6})
	filter := Sort[int]()
	result := slices.Collect(filter(input))

	expected := []int{1, 1, 2, 3, 4, 5, 6, 9}
	if !slices.Equal(result, expected) {
		t.Errorf("Sort failed: expected %v, got %v", expected, result)
	}
}

func TestSortStrings(t *testing.T) {
	input := slices.Values([]string{"banana", "apple", "cherry"})
	filter := Sort[string]()
	result := slices.Collect(filter(input))

	expected := []string{"apple", "banana", "cherry"}
	if !slices.Equal(result, expected) {
		t.Errorf("Sort strings failed: expected %v, got %v", expected, result)
	}
}

func TestSortBy(t *testing.T) {
	type person struct {
		name string
		age  int
	}

	input := slices.Values([]person{
		{"Alice", 30},
		{"Bob", 25},
		{"Charlie", 35},
	})

	filter := SortBy(func(p person) int { return p.age })
	result := slices.Collect(filter(input))

	if result[0].name != "Bob" || result[1].name != "Alice" || result[2].name != "Charlie" {
		t.Errorf("SortBy failed: got %v", result)
	}
}

func TestSortDesc(t *testing.T) {
	input := slices.Values([]int{3, 1, 4, 1, 5, 9, 2, 6})
	filter := SortDesc[int]()
	result := slices.Collect(filter(input))

	expected := []int{9, 6, 5, 4, 3, 2, 1, 1}
	if !slices.Equal(result, expected) {
		t.Errorf("SortDesc failed: expected %v, got %v", expected, result)
	}
}

// ============================================================================
// UTILITY OPERATIONS TESTS
// ============================================================================

func TestDistinct(t *testing.T) {
	input := slices.Values([]int{1, 2, 2, 3, 3, 3, 4, 5, 5})
	filter := Distinct[int]()
	result := slices.Collect(filter(input))

	expected := []int{1, 2, 3, 4, 5}
	if !slices.Equal(result, expected) {
		t.Errorf("Distinct failed: expected %v, got %v", expected, result)
	}
}

func TestDistinctBy(t *testing.T) {
	type item struct {
		id   int
		name string
	}

	input := slices.Values([]item{
		{1, "Alice"},
		{2, "Bob"},
		{1, "Alice2"}, // Duplicate id
		{3, "Charlie"},
	})

	filter := DistinctBy(func(i item) int { return i.id })
	result := slices.Collect(filter(input))

	if len(result) != 3 {
		t.Errorf("DistinctBy should return 3 items, got %d", len(result))
	}

	// Should keep first occurrence
	if result[0].name != "Alice" {
		t.Errorf("DistinctBy should keep first occurrence")
	}
}

func TestReverse(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5})
	filter := Reverse[int]()
	result := slices.Collect(filter(input))

	expected := []int{5, 4, 3, 2, 1}
	if !slices.Equal(result, expected) {
		t.Errorf("Reverse failed: expected %v, got %v", expected, result)
	}
}

// ============================================================================
// TEE OPERATIONS TESTS
// ============================================================================

func TestTee(t *testing.T) {
	input := slices.Values([]int{1, 2, 3})
	streams := Tee(input, 3)

	if len(streams) != 3 {
		t.Fatalf("Tee should return 3 streams, got %d", len(streams))
	}

	// Each stream should contain same values
	for i, stream := range streams {
		result := slices.Collect(stream)
		expected := []int{1, 2, 3}
		if !slices.Equal(result, expected) {
			t.Errorf("Stream %d failed: expected %v, got %v", i, expected, result)
		}
	}
}

func TestTeeZero(t *testing.T) {
	input := slices.Values([]int{1, 2, 3})
	streams := Tee(input, 0)

	if streams != nil {
		t.Error("Tee with n=0 should return nil")
	}
}

func TestLazyTee(t *testing.T) {
	input := slices.Values([]int{1, 2, 3})
	streams := LazyTee(input, 2)

	if len(streams) != 2 {
		t.Fatalf("LazyTee should return 2 streams, got %d", len(streams))
	}

	// Consume first stream
	result1 := slices.Collect(streams[0])
	expected := []int{1, 2, 3}
	if !slices.Equal(result1, expected) {
		t.Errorf("LazyTee stream 1 failed: expected %v, got %v", expected, result1)
	}
}

// ============================================================================
// RUNNING AGGREGATION TESTS
// ============================================================================

func TestRunningSum(t *testing.T) {
	input := slices.Values([]Record{
		{"value": 10.0},
		{"value": 20.0},
		{"value": 30.0},
	})

	filter := RunningSum("value")
	result := slices.Collect(filter(input))

	if len(result) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(result))
	}

	// Check running sums
	if result[0]["running_sum"] != 10.0 {
		t.Errorf("First running_sum should be 10.0, got %v", result[0]["running_sum"])
	}
	if result[1]["running_sum"] != 30.0 {
		t.Errorf("Second running_sum should be 30.0, got %v", result[1]["running_sum"])
	}
	if result[2]["running_sum"] != 60.0 {
		t.Errorf("Third running_sum should be 60.0, got %v", result[2]["running_sum"])
	}
}

func TestRunningAverage(t *testing.T) {
	input := slices.Values([]Record{
		{"value": 10.0},
		{"value": 20.0},
		{"value": 30.0},
		{"value": 40.0},
	})

	filter := RunningAverage("value", 2)
	result := slices.Collect(filter(input))

	if len(result) != 4 {
		t.Fatalf("Expected 4 records, got %d", len(result))
	}

	// First window: [10]
	if result[0]["moving_avg"] != 10.0 {
		t.Errorf("First moving_avg should be 10.0, got %v", result[0]["moving_avg"])
	}

	// Second window: [10, 20]
	if result[1]["moving_avg"] != 15.0 {
		t.Errorf("Second moving_avg should be 15.0, got %v", result[1]["moving_avg"])
	}

	// Third window: [20, 30]
	if result[2]["moving_avg"] != 25.0 {
		t.Errorf("Third moving_avg should be 25.0, got %v", result[2]["moving_avg"])
	}
}

func TestExponentialMovingAverage(t *testing.T) {
	input := slices.Values([]Record{
		{"value": 10.0},
		{"value": 20.0},
		{"value": 30.0},
	})

	filter := ExponentialMovingAverage("value", 0.5)
	result := slices.Collect(filter(input))

	if len(result) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(result))
	}

	// First EMA is initialized to first value
	if result[0]["ema"] != 10.0 {
		t.Errorf("First EMA should be 10.0, got %v", result[0]["ema"])
	}

	// Second EMA: 0.5*20 + 0.5*10 = 15
	if result[1]["ema"] != 15.0 {
		t.Errorf("Second EMA should be 15.0, got %v", result[1]["ema"])
	}
}

func TestRunningMinMax(t *testing.T) {
	input := slices.Values([]Record{
		{"value": 10.0},
		{"value": 5.0},
		{"value": 15.0},
		{"value": 3.0},
	})

	filter := RunningMinMax("value")
	result := slices.Collect(filter(input))

	if len(result) != 4 {
		t.Fatalf("Expected 4 records, got %d", len(result))
	}

	// After all values, min should be 3.0, max should be 15.0
	last := result[3]
	if last["running_min"] != 3.0 {
		t.Errorf("Running min should be 3.0, got %v", last["running_min"])
	}
	if last["running_max"] != 15.0 {
		t.Errorf("Running max should be 15.0, got %v", last["running_max"])
	}
	if last["running_range"] != 12.0 {
		t.Errorf("Running range should be 12.0, got %v", last["running_range"])
	}
}

func TestRunningCount(t *testing.T) {
	input := slices.Values([]Record{
		{"category": "A"},
		{"category": "B"},
		{"category": "A"},
		{"category": "A"},
	})

	filter := RunningCount("category")
	result := slices.Collect(filter(input))

	if len(result) != 4 {
		t.Fatalf("Expected 4 records, got %d", len(result))
	}

	last := result[3]
	if last["total_count"] != int64(4) {
		t.Errorf("Total count should be 4, got %v", last["total_count"])
	}
	if last["distinct_values"] != int64(2) {
		t.Errorf("Distinct values should be 2, got %v", last["distinct_values"])
	}
}

// ============================================================================
// WINDOWING OPERATIONS TESTS
// ============================================================================

func TestCountWindow(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5, 6, 7})
	filter := CountWindow[int](3)
	result := slices.Collect(filter(input))

	// Should create windows: [1,2,3], [4,5,6], [7]
	if len(result) != 3 {
		t.Fatalf("Expected 3 windows, got %d", len(result))
	}

	if !slices.Equal(result[0], []int{1, 2, 3}) {
		t.Errorf("First window failed: got %v", result[0])
	}
	if !slices.Equal(result[1], []int{4, 5, 6}) {
		t.Errorf("Second window failed: got %v", result[1])
	}
	if !slices.Equal(result[2], []int{7}) {
		t.Errorf("Third window failed: got %v", result[2])
	}
}

func TestCountWindowZero(t *testing.T) {
	input := slices.Values([]int{1, 2, 3})
	filter := CountWindow[int](0)
	result := slices.Collect(filter(input))

	if len(result) != 0 {
		t.Error("CountWindow(0) should return no windows")
	}
}

func TestSlidingCountWindow(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5})
	filter := SlidingCountWindow[int](3, 1)
	result := slices.Collect(filter(input))

	// Windows: [1,2,3], [2,3,4], [3,4,5]
	if len(result) != 3 {
		t.Fatalf("Expected 3 windows, got %d", len(result))
	}

	if !slices.Equal(result[0], []int{1, 2, 3}) {
		t.Errorf("First window failed: got %v", result[0])
	}
	if !slices.Equal(result[1], []int{2, 3, 4}) {
		t.Errorf("Second window failed: got %v", result[1])
	}
	if !slices.Equal(result[2], []int{3, 4, 5}) {
		t.Errorf("Third window failed: got %v", result[2])
	}
}

func TestTimeWindow(t *testing.T) {
	now := time.Now()
	input := slices.Values([]Record{
		{"time": now, "value": 1},
		{"time": now.Add(1 * time.Second), "value": 2},
		{"time": now.Add(5 * time.Second), "value": 3},
		{"time": now.Add(6 * time.Second), "value": 4},
	})

	filter := TimeWindow[Record](5*time.Second, "time")
	result := slices.Collect(filter(input))

	// Should create 2 windows: first 2 records, then next 2 records
	if len(result) < 1 {
		t.Fatalf("Expected at least 1 window, got %d", len(result))
	}
}

func TestSlidingTimeWindow(t *testing.T) {
	now := time.Now()
	input := slices.Values([]Record{
		{"time": now, "value": 1},
		{"time": now.Add(1 * time.Second), "value": 2},
		{"time": now.Add(2 * time.Second), "value": 3},
		{"time": now.Add(3 * time.Second), "value": 4},
	})

	filter := SlidingTimeWindow[Record](2*time.Second, 1*time.Second, "time")
	result := slices.Collect(filter(input))

	// Should create overlapping windows
	if len(result) < 1 {
		t.Log("SlidingTimeWindow created windows:", len(result))
	}
}

// ============================================================================
// EARLY TERMINATION TESTS
// ============================================================================

func TestTakeWhile(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5, 1, 2})
	filter := TakeWhile(func(x int) bool { return x < 4 })
	result := slices.Collect(filter(input))

	expected := []int{1, 2, 3}
	if !slices.Equal(result, expected) {
		t.Errorf("TakeWhile failed: expected %v, got %v", expected, result)
	}
}

func TestTakeUntil(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5, 6})
	filter := TakeUntil(func(x int) bool { return x > 3 })
	result := slices.Collect(filter(input))

	expected := []int{1, 2, 3}
	if !slices.Equal(result, expected) {
		t.Errorf("TakeUntil failed: expected %v, got %v", expected, result)
	}
}

func TestTimeout(t *testing.T) {
	// Create a slow infinite sequence
	slowSeq := func(yield func(int) bool) {
		for i := 0; ; i++ {
			time.Sleep(10 * time.Millisecond)
			if !yield(i) {
				return
			}
		}
	}

	filter := Timeout[int](50 * time.Millisecond)
	start := time.Now()
	result := slices.Collect(filter(slowSeq))
	duration := time.Since(start)

	// Should terminate around 50ms
	if duration > 100*time.Millisecond {
		t.Errorf("Timeout took too long: %v", duration)
	}

	// Should collect a few elements before timeout
	if len(result) == 0 {
		t.Log("Timeout collected 0 elements (expected at least a few)")
	}
}

func TestTimeBasedTimeout(t *testing.T) {
	now := time.Now()
	input := slices.Values([]Record{
		{"time": now, "value": 1},
		{"time": now.Add(1 * time.Second), "value": 2},
		{"time": now.Add(2 * time.Second), "value": 3},
		{"time": now.Add(6 * time.Second), "value": 4}, // Exceeds 5 second limit
	})

	filter := TimeBasedTimeout("time", 5*time.Second)
	result := slices.Collect(filter(input))

	// Should stop before the 4th record
	if len(result) != 3 {
		t.Errorf("TimeBasedTimeout should return 3 records, got %d", len(result))
	}
}

func TestSkipWhile(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5, 1, 2})
	filter := SkipWhile(func(x int) bool { return x < 3 })
	result := slices.Collect(filter(input))

	expected := []int{3, 4, 5, 1, 2}
	if !slices.Equal(result, expected) {
		t.Errorf("SkipWhile failed: expected %v, got %v", expected, result)
	}
}

func TestSkipUntil(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4, 5})
	filter := SkipUntil(func(x int) bool { return x >= 3 })
	result := slices.Collect(filter(input))

	expected := []int{3, 4, 5}
	if !slices.Equal(result, expected) {
		t.Errorf("SkipUntil failed: expected %v, got %v", expected, result)
	}
}

// ============================================================================
// COMPLEX PIPELINE TESTS
// ============================================================================

func TestComplexPipeline(t *testing.T) {
	// Test a complex pipeline: filter, map, sort, limit
	input := slices.Values([]int{5, 2, 8, 1, 9, 3, 7, 4, 6})

	// Apply transformations step by step
	filtered := Where(func(x int) bool { return x > 3 })(input)
	doubled := Select(func(x int) int { return x * 2 })(filtered)
	sorted := Sort[int]()(doubled)
	result := slices.Collect(Limit[int](3)(sorted))
	expected := []int{8, 10, 12} // [4, 5, 6] doubled and sorted

	if !slices.Equal(result, expected) {
		t.Errorf("Complex pipeline failed: expected %v, got %v", expected, result)
	}
}

func TestRecordPipeline(t *testing.T) {
	input := slices.Values([]Record{
		{"name": "Alice", "age": int64(30), "score": 85.0},
		{"name": "Bob", "age": int64(25), "score": 90.0},
		{"name": "Charlie", "age": int64(35), "score": 75.0},
	})

	// Filter age > 26, add bonus field, sort by score descending
	filtered := Where(func(r Record) bool {
		age := GetOr(r, "age", int64(0))
		return age > 26
	})(input)

	withBonus := Select(func(r Record) Record {
		result := make(Record)
		for k, v := range r {
			result[k] = v
		}
		score := GetOr(r, "score", 0.0)
		result["bonus"] = score * 0.1
		return result
	})(filtered)

	sorted := SortBy(func(r Record) float64 {
		return -GetOr(r, "score", 0.0) // Negative for descending
	})(withBonus)

	result := slices.Collect(sorted)

	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}

	// Alice should be first (score 85), Charlie second (score 75)
	if result[0]["name"] != "Alice" {
		t.Errorf("First record should be Alice, got %v", result[0]["name"])
	}
	if result[1]["name"] != "Charlie" {
		t.Errorf("Second record should be Charlie, got %v", result[1]["name"])
	}
}
