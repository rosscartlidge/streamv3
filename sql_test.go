package streamv3

import (
	"iter"
	"slices"
	"testing"
)

// ============================================================================
// JOIN OPERATIONS TESTS
// ============================================================================

func TestInnerJoin(t *testing.T) {
	left := slices.Values([]Record{
		{"id": int64(1), "name": "Alice"},
		{"id": int64(2), "name": "Bob"},
		{"id": int64(3), "name": "Charlie"},
	})

	right := slices.Values([]Record{
		{"id": int64(1), "dept": "Engineering"},
		{"id": int64(2), "dept": "Sales"},
	})

	filter := InnerJoin(right, OnFields("id"))
	result := slices.Collect(filter(left))

	// Should match only records with id 1 and 2
	if len(result) != 2 {
		t.Fatalf("InnerJoin should return 2 records, got %d", len(result))
	}

	// Check first joined record
	if result[0]["name"] != "Alice" || result[0]["dept"] != "Engineering" {
		t.Errorf("First join failed: %v", result[0])
	}
}

func TestLeftJoin(t *testing.T) {
	left := slices.Values([]Record{
		{"id": int64(1), "name": "Alice"},
		{"id": int64(2), "name": "Bob"},
		{"id": int64(3), "name": "Charlie"},
	})

	right := slices.Values([]Record{
		{"id": int64(1), "dept": "Engineering"},
		{"id": int64(2), "dept": "Sales"},
	})

	filter := LeftJoin(right, OnFields("id"))
	result := slices.Collect(filter(left))

	// Should include all left records
	if len(result) != 3 {
		t.Fatalf("LeftJoin should return 3 records, got %d", len(result))
	}

	// Charlie should exist without dept field
	found := false
	for _, r := range result {
		if r["name"] == "Charlie" {
			found = true
			if r.Has("dept") {
				t.Error("Charlie should not have dept field in left join")
			}
		}
	}

	if !found {
		t.Error("LeftJoin should include Charlie")
	}
}

func TestRightJoin(t *testing.T) {
	left := slices.Values([]Record{
		{"id": int64(1), "name": "Alice"},
		{"id": int64(2), "name": "Bob"},
	})

	right := slices.Values([]Record{
		{"id": int64(1), "dept": "Engineering"},
		{"id": int64(2), "dept": "Sales"},
		{"id": int64(3), "dept": "Marketing"},
	})

	filter := RightJoin(right, OnFields("id"))
	result := slices.Collect(filter(left))

	// Should include all right records
	if len(result) != 3 {
		t.Fatalf("RightJoin should return 3 records, got %d", len(result))
	}

	// Marketing dept should exist without name field
	found := false
	for _, r := range result {
		if r["dept"] == "Marketing" {
			found = true
			if r.Has("name") {
				t.Error("Marketing record should not have name field")
			}
		}
	}

	if !found {
		t.Error("RightJoin should include Marketing dept")
	}
}

func TestFullJoin(t *testing.T) {
	left := slices.Values([]Record{
		{"id": int64(1), "name": "Alice"},
		{"id": int64(2), "name": "Bob"},
		{"id": int64(4), "name": "David"},
	})

	right := slices.Values([]Record{
		{"id": int64(1), "dept": "Engineering"},
		{"id": int64(2), "dept": "Sales"},
		{"id": int64(3), "dept": "Marketing"},
	})

	filter := FullJoin(right, OnFields("id"))
	result := slices.Collect(filter(left))

	// Should include matched (id 1, 2), unmatched left (David), unmatched right (Marketing)
	if len(result) != 4 {
		t.Fatalf("FullJoin should return 4 records, got %d", len(result))
	}

	hasDavid := false
	hasMarketing := false

	for _, r := range result {
		if r["name"] == "David" {
			hasDavid = true
		}
		if r["dept"] == "Marketing" {
			hasMarketing = true
		}
	}

	if !hasDavid {
		t.Error("FullJoin should include David (unmatched left)")
	}
	if !hasMarketing {
		t.Error("FullJoin should include Marketing (unmatched right)")
	}
}

func TestOnFields(t *testing.T) {
	predicate := OnFields("id", "type")

	left := Record{"id": int64(1), "type": "A", "name": "Alice"}
	right := Record{"id": int64(1), "type": "A", "dept": "Eng"}

	if !predicate(left, right) {
		t.Error("OnFields should match records with same id and type")
	}

	right2 := Record{"id": int64(1), "type": "B", "dept": "Sales"}
	if predicate(left, right2) {
		t.Error("OnFields should not match records with different type")
	}
}

func TestOnCondition(t *testing.T) {
	predicate := OnCondition(func(left, right Record) bool {
		leftAge, ok1 := Get[int64](left, "age")
		rightAge, ok2 := Get[int64](right, "age")
		return ok1 && ok2 && leftAge > rightAge
	})

	left := Record{"name": "Alice", "age": int64(30)}
	right := Record{"name": "Bob", "age": int64(25)}

	if !predicate(left, right) {
		t.Error("OnCondition should match when left age > right age")
	}

	right2 := Record{"name": "Charlie", "age": int64(35)}
	if predicate(left, right2) {
		t.Error("OnCondition should not match when left age < right age")
	}
}

// ============================================================================
// GROUPBY OPERATIONS TESTS
// ============================================================================

func TestGroupBy(t *testing.T) {
	input := slices.Values([]Record{
		{"name": "Alice", "dept": "Engineering", "salary": int64(100000)},
		{"name": "Bob", "dept": "Engineering", "salary": int64(110000)},
		{"name": "Charlie", "dept": "Sales", "salary": int64(90000)},
	})

	filter := GroupBy("employees", "department", func(r Record) string {
		return GetOr(r, "dept", "")
	})

	result := slices.Collect(filter(input))

	// Should create 2 groups: Engineering and Sales
	if len(result) != 2 {
		t.Fatalf("GroupBy should return 2 groups, got %d", len(result))
	}

	// Check that each group has the sequence field
	for _, group := range result {
		if !group.Has("department") {
			t.Error("Group should have 'department' field")
		}
		if !group.Has("employees") {
			t.Error("Group should have 'employees' field")
		}

		// Check sequence field type
		if seq, ok := group["employees"].(iter.Seq[Record]); ok {
			members := slices.Collect(seq)
			if len(members) == 0 {
				t.Error("Group should have members")
			}
		} else {
			t.Error("employees field should be iter.Seq[Record]")
		}
	}
}

func TestGroupByFields(t *testing.T) {
	input := slices.Values([]Record{
		{"dept": "Eng", "level": "Senior", "count": int64(5)},
		{"dept": "Eng", "level": "Senior", "count": int64(3)},
		{"dept": "Eng", "level": "Junior", "count": int64(10)},
		{"dept": "Sales", "level": "Senior", "count": int64(2)},
	})

	filter := GroupByFields("members", "dept", "level")
	result := slices.Collect(filter(input))

	// Should create 3 groups: (Eng,Senior), (Eng,Junior), (Sales,Senior)
	if len(result) != 3 {
		t.Fatalf("GroupByFields should return 3 groups, got %d", len(result))
	}

	// Find the (Eng, Senior) group
	for _, group := range result {
		if group["dept"] == "Eng" && group["level"] == "Senior" {
			// This group should have 2 members
			if seq, ok := group["members"].(iter.Seq[Record]); ok {
				members := slices.Collect(seq)
				if len(members) != 2 {
					t.Errorf("Eng/Senior group should have 2 members, got %d", len(members))
				}
			}
		}
	}
}

func TestGroupByFieldsWithComplexValues(t *testing.T) {
	// Test that records with complex GROUPING field values are skipped
	nestedRecord := Record{"city": "NYC"}

	input := slices.Values([]Record{
		{"dept": nestedRecord, "name": "Alice"},  // Complex dept field - should be skipped
		{"dept": "Sales", "location": "Boston"},  // Simple dept - should work
	})

	filter := GroupByFields("members", "dept")
	result := slices.Collect(filter(input))

	// Should only group the Sales record (Eng has complex dept value)
	if len(result) != 1 {
		t.Fatalf("GroupByFields should skip records with complex grouping field values, expected 1 group, got %d", len(result))
	}

	if result[0]["dept"] != "Sales" {
		t.Error("Should only have Sales group")
	}
}

// ============================================================================
// AGGREGATION OPERATIONS TESTS
// ============================================================================

func TestAggregate(t *testing.T) {
	// Create a grouped result first
	input := slices.Values([]Record{
		{"dept": "Eng", "salary": int64(100000)},
		{"dept": "Eng", "salary": int64(110000)},
		{"dept": "Sales", "salary": int64(90000)},
	})

	grouped := GroupByFields("employees", "dept")(input)

	// Now aggregate
	aggregated := Aggregate("employees", map[string]AggregateFunc{
		"count":       Count(),
		"total":       Sum("salary"),
		"avg_salary":  Avg("salary"),
	})(grouped)

	result := slices.Collect(aggregated)

	if len(result) != 2 {
		t.Fatalf("Aggregate should return 2 groups, got %d", len(result))
	}

	// Find Engineering department
	for _, r := range result {
		if r["dept"] == "Eng" {
			if r["count"] != int64(2) {
				t.Errorf("Eng count should be 2, got %v", r["count"])
			}
			if r["total"] != float64(210000) {
				t.Errorf("Eng total should be 210000, got %v", r["total"])
			}
			if r["avg_salary"] != float64(105000) {
				t.Errorf("Eng avg should be 105000, got %v", r["avg_salary"])
			}
		}
	}
}

func TestCount(t *testing.T) {
	records := []Record{
		{"name": "Alice"},
		{"name": "Bob"},
		{"name": "Charlie"},
	}

	countFn := Count()
	result := countFn(records)

	if result != int64(3) {
		t.Errorf("Count should return 3, got %v", result)
	}
}

func TestSum(t *testing.T) {
	records := []Record{
		{"value": int64(10)},
		{"value": int64(20)},
		{"value": int64(30)},
	}

	sumFn := Sum("value")
	result := sumFn(records)

	if result != float64(60) {
		t.Errorf("Sum should return 60, got %v", result)
	}
}

func TestSumMixedTypes(t *testing.T) {
	records := []Record{
		{"value": int64(10)},
		{"value": 20.5},
		{"value": "30"}, // Should convert
	}

	sumFn := Sum("value")
	result := sumFn(records)

	if result != float64(60.5) {
		t.Errorf("Sum should handle mixed types and return 60.5, got %v", result)
	}
}

func TestAvg(t *testing.T) {
	records := []Record{
		{"score": int64(80)},
		{"score": int64(90)},
		{"score": int64(100)},
	}

	avgFn := Avg("score")
	result := avgFn(records)

	if result != float64(90) {
		t.Errorf("Avg should return 90, got %v", result)
	}
}

func TestAvgEmpty(t *testing.T) {
	records := []Record{}

	avgFn := Avg("score")
	result := avgFn(records)

	if result != 0.0 {
		t.Errorf("Avg of empty should return 0, got %v", result)
	}
}

func TestMin(t *testing.T) {
	records := []Record{
		{"value": int64(50)},
		{"value": int64(10)},
		{"value": int64(30)},
	}

	minFn := Min[int64]("value")
	result := minFn(records)

	if result != int64(10) {
		t.Errorf("Min should return 10, got %v", result)
	}
}

func TestMinFloat(t *testing.T) {
	records := []Record{
		{"value": 5.5},
		{"value": 2.3},
		{"value": 7.8},
	}

	minFn := Min[float64]("value")
	result := minFn(records)

	if result != 2.3 {
		t.Errorf("Min should return 2.3, got %v", result)
	}
}

func TestMax(t *testing.T) {
	records := []Record{
		{"value": int64(50)},
		{"value": int64(10)},
		{"value": int64(100)},
	}

	maxFn := Max[int64]("value")
	result := maxFn(records)

	if result != int64(100) {
		t.Errorf("Max should return 100, got %v", result)
	}
}

func TestMaxString(t *testing.T) {
	records := []Record{
		{"name": "Alice"},
		{"name": "Charlie"},
		{"name": "Bob"},
	}

	maxFn := Max[string]("name")
	result := maxFn(records)

	if result != "Charlie" {
		t.Errorf("Max should return Charlie, got %v", result)
	}
}

func TestFirst(t *testing.T) {
	records := []Record{
		{"value": "first"},
		{"value": "second"},
		{"value": "third"},
	}

	firstFn := First("value")
	result := firstFn(records)

	if result != "first" {
		t.Errorf("First should return 'first', got %v", result)
	}
}

func TestFirstEmpty(t *testing.T) {
	records := []Record{}

	firstFn := First("value")
	result := firstFn(records)

	if result != nil {
		t.Errorf("First of empty should return nil, got %v", result)
	}
}

func TestLast(t *testing.T) {
	records := []Record{
		{"value": "first"},
		{"value": "second"},
		{"value": "third"},
	}

	lastFn := Last("value")
	result := lastFn(records)

	if result != "third" {
		t.Errorf("Last should return 'third', got %v", result)
	}
}

func TestLastEmpty(t *testing.T) {
	records := []Record{}

	lastFn := Last("value")
	result := lastFn(records)

	if result != nil {
		t.Errorf("Last of empty should return nil, got %v", result)
	}
}

func TestCollect(t *testing.T) {
	records := []Record{
		{"name": "Alice"},
		{"name": "Bob"},
		{"name": "Charlie"},
	}

	collectFn := Collect("name")
	result := collectFn(records)

	values, ok := result.([]any)
	if !ok {
		t.Fatal("Collect should return []any")
	}

	if len(values) != 3 {
		t.Fatalf("Collect should return 3 values, got %d", len(values))
	}

	if values[0] != "Alice" || values[1] != "Bob" || values[2] != "Charlie" {
		t.Errorf("Collect values incorrect: %v", values)
	}
}

// ============================================================================
// COMPLEX SQL-STYLE PIPELINE TESTS
// ============================================================================

func TestComplexGroupByAggregate(t *testing.T) {
	// Sales data
	input := slices.Values([]Record{
		{"product": "Widget", "region": "East", "sales": int64(1000)},
		{"product": "Widget", "region": "East", "sales": int64(1500)},
		{"product": "Widget", "region": "West", "sales": int64(2000)},
		{"product": "Gadget", "region": "East", "sales": int64(800)},
		{"product": "Gadget", "region": "West", "sales": int64(1200)},
	})

	// Group by product and region, then aggregate
	pipeline := Chain(
		GroupByFields("items", "product", "region"),
		Aggregate("items", map[string]AggregateFunc{
			"count":       Count(),
			"total_sales": Sum("sales"),
			"avg_sales":   Avg("sales"),
			"max_sale":    Max[int64]("sales"),
			"min_sale":    Min[int64]("sales"),
		}),
	)

	result := slices.Collect(pipeline(input))

	// Should have 4 groups: Widget/East, Widget/West, Gadget/East, Gadget/West
	if len(result) != 4 {
		t.Fatalf("Expected 4 groups, got %d", len(result))
	}

	// Find Widget/East group
	for _, r := range result {
		if r["product"] == "Widget" && r["region"] == "East" {
			if r["count"] != int64(2) {
				t.Errorf("Widget/East count should be 2, got %v", r["count"])
			}
			if r["total_sales"] != float64(2500) {
				t.Errorf("Widget/East total should be 2500, got %v", r["total_sales"])
			}
			if r["max_sale"] != int64(1500) {
				t.Errorf("Widget/East max should be 1500, got %v", r["max_sale"])
			}
		}
	}
}

func TestJoinThenGroup(t *testing.T) {
	// Employee data
	employees := slices.Values([]Record{
		{"id": int64(1), "name": "Alice", "dept_id": int64(10)},
		{"id": int64(2), "name": "Bob", "dept_id": int64(10)},
		{"id": int64(3), "name": "Charlie", "dept_id": int64(20)},
	})

	// Department data
	departments := slices.Values([]Record{
		{"dept_id": int64(10), "dept_name": "Engineering"},
		{"dept_id": int64(20), "dept_name": "Sales"},
	})

	// Join employees with departments, then group by department
	pipeline := Chain(
		InnerJoin(departments, OnFields("dept_id")),
		GroupByFields("members", "dept_name"),
		Aggregate("members", map[string]AggregateFunc{
			"count": Count(),
			"names": Collect("name"),
		}),
	)

	result := slices.Collect(pipeline(employees))

	if len(result) != 2 {
		t.Fatalf("Expected 2 departments, got %d", len(result))
	}

	// Find Engineering department
	for _, r := range result {
		if r["dept_name"] == "Engineering" {
			if r["count"] != int64(2) {
				t.Errorf("Engineering should have 2 employees, got %v", r["count"])
			}

			names, ok := r["names"].([]any)
			if !ok || len(names) != 2 {
				t.Errorf("Engineering should have 2 names, got %v", r["names"])
			}
		}
	}
}

func TestMultiFieldGrouping(t *testing.T) {
	// Test grouping by multiple fields
	input := slices.Values([]Record{
		{"year": int64(2024), "month": "Jan", "sales": int64(1000)},
		{"year": int64(2024), "month": "Jan", "sales": int64(1100)},
		{"year": int64(2024), "month": "Feb", "sales": int64(1200)},
		{"year": int64(2023), "month": "Jan", "sales": int64(900)},
	})

	pipeline := Chain(
		GroupByFields("records", "year", "month"),
		Aggregate("records", map[string]AggregateFunc{
			"total": Sum("sales"),
			"count": Count(),
		}),
	)

	result := slices.Collect(pipeline(input))

	// Should have 3 groups: (2024,Jan), (2024,Feb), (2023,Jan)
	if len(result) != 3 {
		t.Fatalf("Expected 3 groups, got %d", len(result))
	}

	// Check 2024/Jan group
	for _, r := range result {
		if r["year"] == int64(2024) && r["month"] == "Jan" {
			if r["count"] != int64(2) {
				t.Errorf("2024/Jan should have count 2, got %v", r["count"])
			}
			if r["total"] != float64(2100) {
				t.Errorf("2024/Jan should have total 2100, got %v", r["total"])
			}
		}
	}
}

// ============================================================================
// EDGE CASE TESTS
// ============================================================================

func TestJoinEmptyRight(t *testing.T) {
	left := slices.Values([]Record{
		{"id": int64(1), "name": "Alice"},
	})

	empty := func(yield func(Record) bool) {
		// Yield nothing
	}

	filter := InnerJoin(iter.Seq[Record](empty), OnFields("id"))
	result := slices.Collect(filter(left))

	if len(result) != 0 {
		t.Errorf("Inner join with empty right should return 0 records, got %d", len(result))
	}
}

func TestGroupByEmptyInput(t *testing.T) {
	empty := func(yield func(Record) bool) {
		// Yield nothing
	}

	filter := GroupByFields("members", "dept")
	result := slices.Collect(filter(iter.Seq[Record](empty)))

	if len(result) != 0 {
		t.Errorf("GroupBy on empty should return 0 groups, got %d", len(result))
	}
}

func TestAggregateNoSequenceField(t *testing.T) {
	input := slices.Values([]Record{
		{"dept": "Eng", "count": int64(5)}, // No sequence field
	})

	filter := Aggregate("members", map[string]AggregateFunc{
		"total": Count(),
	})

	result := slices.Collect(filter(input))

	// Should still yield the record without aggregations
	if len(result) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result))
	}

	if result[0].Has("total") {
		t.Error("Should not have aggregation when sequence field is missing")
	}
}
