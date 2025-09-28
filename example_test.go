package streamv3

import (
	"fmt"
	"os"
	"testing"
)

// TestStreamV3Basic demonstrates basic StreamV3 functionality
func TestStreamV3Basic(t *testing.T) {
	// Create a simple stream of integers
	numbers := From([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	// Use fluent API to filter even numbers, double them, and limit to 3
	result := numbers.
		Where(func(x int) bool { return x%2 == 0 }).
		Limit(3).
		Collect()

	expected := []int{2, 4, 6}
	if len(result) != len(expected) {
		t.Errorf("Expected %d items, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Expected %d at index %d, got %d", expected[i], i, v)
		}
	}
}

// TestStreamV3Records demonstrates Record processing
func TestStreamV3Records(t *testing.T) {
	// Create records using the fluent builder
	people := From([]Record{
		NewRecord().String("name", "Alice").Int("age", 30).Build(),
		NewRecord().String("name", "Bob").Int("age", 25).Build(),
		NewRecord().String("name", "Charlie").Int("age", 35).Build(),
	})

	// Filter adults over 25 and add a computed field
	adults := AddRecordField(
		people.Where(func(r Record) bool {
			if age, exists := r["age"]; exists {
				if ageInt, ok := age.(int64); ok {
					return ageInt > 25
				}
			}
			return false
		}),
		"category",
		func(r Record) any { return "adult" },
	)

	result := adults.Collect()

	if len(result) != 2 {
		t.Errorf("Expected 2 adults, got %d", len(result))
	}

	for _, record := range result {
		if category, exists := record["category"]; !exists || category != "adult" {
			t.Errorf("Expected category 'adult', got %v", category)
		}
	}
}

// TestStreamV3Aggregation demonstrates SQL-style operations
func TestStreamV3Aggregation(t *testing.T) {
	// Create sales records
	sales := From([]Record{
		NewRecord().String("product", "laptop").Float("amount", 1000.0).String("region", "US").Build(),
		NewRecord().String("product", "mouse").Float("amount", 25.0).String("region", "US").Build(),
		NewRecord().String("product", "laptop").Float("amount", 1200.0).String("region", "EU").Build(),
		NewRecord().String("product", "mouse").Float("amount", 30.0).String("region", "EU").Build(),
	})

	// Group by product and calculate total sales
	grouped := GroupRecordsByFields(sales, "product")
	aggregated := AggregateGroups(grouped, map[string]AggregateFunc{
		"total_sales": Sum("amount"),
		"count":       Count(),
	})

	result := aggregated.Collect()

	if len(result) != 2 {
		t.Errorf("Expected 2 product groups, got %d", len(result))
	}

	// Verify laptop sales
	for _, record := range result {
		if groupKey, exists := record["group_key"]; exists && groupKey == "[laptop]" {
			if totalSales, exists := record["total_sales"]; exists {
				if totalSalesFloat, ok := totalSales.(float64); ok {
					if totalSalesFloat != 2200.0 {
						t.Errorf("Expected laptop sales 2200.0, got %v", totalSalesFloat)
					}
				}
			}
		}
	}
}

// TestStreamV3ErrorHandling demonstrates error-aware processing
func TestStreamV3ErrorHandling(t *testing.T) {
	// Create error-aware stream
	numbers := FromSafe([]int{1, 2, 3, 4, 5})

	// Apply error-aware transformation
	processed := MapToSafe(numbers, func(x int) (string, error) {
		if x == 3 {
			return "", fmt.Errorf("error processing %d", x)
		}
		return fmt.Sprintf("item_%d", x), nil
	})

	result, err := processed.Collect()

	// Should get error when processing 3
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Should have processed items before the error
	if len(result) != 2 {
		t.Errorf("Expected 2 items before error, got %d", len(result))
	}
}

// TestStreamV3StandardLibraryIntegration demonstrates using Go 1.23 features
func TestStreamV3StandardLibraryIntegration(t *testing.T) {
	// Create a stream of ordered elements
	numbers := From([]int{5, 2, 8, 1, 9, 3})

	// Sort using standard library functions
	sorted := SortOrdered(numbers)
	result := sorted.Collect()

	expected := []int{1, 2, 3, 5, 8, 9}
	if len(result) != len(expected) {
		t.Errorf("Expected %d items, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Expected %d at index %d, got %d", expected[i], i, v)
		}
	}
}

// TestStreamV3CSV demonstrates CSV I/O
func TestStreamV3CSV(t *testing.T) {
	// Create test CSV data
	testFile := "/tmp/test_streamv3.csv"
	csvData := `name,age,city
Alice,30,New York
Bob,25,San Francisco
Charlie,35,Chicago`

	// Write test file
	if err := os.WriteFile(testFile, []byte(csvData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Read CSV using StreamV3
	records := ReadCSV(testFile)
	result := records.Collect()

	if len(result) != 3 {
		t.Errorf("Expected 3 records, got %d", len(result))
	}

	// Check first record
	if len(result) > 0 {
		first := result[0]
		if name, exists := first["name"]; !exists || name != "Alice" {
			t.Errorf("Expected name 'Alice', got %v", name)
		}
		if age, exists := first["age"]; !exists || age != int64(30) {
			t.Errorf("Expected age 30, got %v", age)
		}
	}
}

// Example_basicUsage demonstrates basic StreamV3 usage
func Example_basicUsage() {
	// Create a stream and process it
	numbers := From([]int{1, 2, 3, 4, 5})

	// Use fluent API
	for result := range numbers.Where(func(x int) bool { return x%2 == 0 }).Iter() {
		fmt.Printf("Even number: %d\n", result)
	}
	// Output:
	// Even number: 2
	// Even number: 4
}

// Example_records demonstrates Record processing
func Example_records() {
	// Create records
	people := From([]Record{
		NewRecord().String("name", "Alice").Int("age", 30).Build(),
		NewRecord().String("name", "Bob").Int("age", 25).Build(),
	})

	// Process records
	for person := range people.Iter() {
		fmt.Printf("Person: %s, Age: %v\n", person["name"], person["age"])
	}
	// Output:
	// Person: Alice, Age: 30
	// Person: Bob, Age: 25
}

// TestInteractiveChart demonstrates creating interactive charts
func TestInteractiveChart(t *testing.T) {
	// Create sample data
	sampleData := From([]Record{
		NewRecord().String("date", "2024-01-01").Float("sales", 1000).Float("profit", 200).String("region", "US").Build(),
		NewRecord().String("date", "2024-01-02").Float("sales", 1200).Float("profit", 240).String("region", "US").Build(),
		NewRecord().String("date", "2024-01-03").Float("sales", 950).Float("profit", 190).String("region", "EU").Build(),
		NewRecord().String("date", "2024-01-04").Float("sales", 1100).Float("profit", 220).String("region", "EU").Build(),
		NewRecord().String("date", "2024-01-05").Float("sales", 1300).Float("profit", 260).String("region", "US").Build(),
	})

	// Create interactive chart
	config := DefaultChartConfig()
	config.Title = "Sales and Profit Analysis"
	config.ChartType = "line"
	config.EnableCalculations = true
	config.ColorScheme = "vibrant"

	testFile := "/tmp/test_chart.html"
	err := InteractiveChart(sampleData, testFile, config)
	if err != nil {
		t.Fatalf("Failed to create chart: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Chart file was not created")
	}

	// Clean up
	defer os.Remove(testFile)
}

// TestQuickChart demonstrates the simple chart API
func TestQuickChart(t *testing.T) {
	// Create simple numeric data
	data := From([]Record{
		NewRecord().String("month", "Jan").Float("revenue", 10000).Build(),
		NewRecord().String("month", "Feb").Float("revenue", 12000).Build(),
		NewRecord().String("month", "Mar").Float("revenue", 11500).Build(),
		NewRecord().String("month", "Apr").Float("revenue", 13000).Build(),
	})

	testFile := "/tmp/quick_chart.html"
	err := QuickChart(data, "month", "revenue", testFile)
	if err != nil {
		t.Fatalf("Failed to create quick chart: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Quick chart file was not created")
	}

	// Clean up
	defer os.Remove(testFile)
}

// TestTimeSeriesChart demonstrates time-based charting
func TestTimeSeriesChart(t *testing.T) {
	// Create time series data
	timeSeries := From([]Record{
		NewRecord().String("timestamp", "2024-01-01 10:00:00").Float("cpu_usage", 45.2).Float("memory_usage", 67.8).Build(),
		NewRecord().String("timestamp", "2024-01-01 10:05:00").Float("cpu_usage", 52.1).Float("memory_usage", 69.2).Build(),
		NewRecord().String("timestamp", "2024-01-01 10:10:00").Float("cpu_usage", 38.7).Float("memory_usage", 65.4).Build(),
		NewRecord().String("timestamp", "2024-01-01 10:15:00").Float("cpu_usage", 61.3).Float("memory_usage", 72.1).Build(),
	})

	config := DefaultChartConfig()
	config.Title = "System Metrics Over Time"
	config.EnableCalculations = true

	testFile := "/tmp/timeseries_chart.html"
	err := TimeSeriesChart(timeSeries, "timestamp", []string{"cpu_usage", "memory_usage"}, testFile, config)
	if err != nil {
		t.Fatalf("Failed to create time series chart: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Time series chart file was not created")
	}

	// Clean up
	defer os.Remove(testFile)
}

// TestCommandOutput demonstrates parsing command output
func TestCommandOutput(t *testing.T) {
	// Create test command output file (like ps command output)
	testFile := "/tmp/test_command_output.txt"
	commandOutput := `F S   UID     PID    PPID  C PRI  NI ADDR SZ WCHAN  STIME TTY          TIME CMD
0 S     0       1       0  0  80   0 -  4234 -      Dec28 ?        00:00:02 /sbin/init
0 S     0       2       0  0  80   0 -     0 -      Dec28 ?        00:00:00 [kthreadd]
0 S     0       3       1  0  80   0 -  1234 -      Dec28 ?        00:00:01 /usr/bin/systemd`

	// Write test file
	if err := os.WriteFile(testFile, []byte(commandOutput), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Parse command output
	processes := ReadCommandOutput(testFile)
	result := processes.Collect()

	if len(result) != 3 {
		t.Errorf("Expected 3 processes, got %d", len(result))
	}

	// Check first process (init)
	if len(result) > 0 {
		init := result[0]
		if pid, exists := init["PID"]; !exists || pid != int64(1) {
			t.Errorf("Expected PID 1 for init, got %v", pid)
		}
		if cmd, exists := init["CMD"]; !exists || cmd != "/sbin/init" {
			t.Errorf("Expected CMD '/sbin/init', got %v", cmd)
		}
	}
}