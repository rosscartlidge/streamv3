package streamv3

import (
	"iter"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// ============================================================================
// CHART CONFIG TESTS
// ============================================================================

func TestDefaultChartConfig(t *testing.T) {
	config := DefaultChartConfig()

	if config.ChartType != "line" {
		t.Errorf("Default chart type should be 'line', got %s", config.ChartType)
	}

	if config.Title != "Data Visualization" {
		t.Errorf("Default title should be 'Data Visualization', got %s", config.Title)
	}

	if config.Width != 1200 {
		t.Errorf("Default width should be 1200, got %d", config.Width)
	}

	if config.Height != 600 {
		t.Errorf("Default height should be 600, got %d", config.Height)
	}

	if !config.EnableZoom {
		t.Error("Default EnableZoom should be true")
	}
}

// ============================================================================
// INTERACTIVE CHART TESTS
// ============================================================================

func TestInteractiveChart(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test_chart.html")

	// Create test data
	input := slices.Values([]Record{
		{"x": int64(1), "y": int64(10)},
		{"x": int64(2), "y": int64(20)},
		{"x": int64(3), "y": int64(15)},
	})

	config := DefaultChartConfig()

	err := InteractiveChart(input, filename, config)
	if err != nil {
		t.Fatalf("InteractiveChart failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Chart file was not created")
	}

	// Read the file and check for basic content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read chart file: %v", err)
	}

	htmlContent := string(content)

	// Check for essential HTML elements
	if !strings.Contains(htmlContent, "<html") {
		t.Error("Chart should contain HTML tag")
	}
	if !strings.Contains(htmlContent, "chart") {
		t.Error("Chart should include chart-related content")
	}
	if !strings.Contains(htmlContent, "canvas") {
		t.Error("Chart should contain canvas element")
	}
}

func TestInteractiveChartCustomConfig(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "custom_chart.html")

	input := slices.Values([]Record{
		{"category": "A", "value": int64(100)},
		{"category": "B", "value": int64(200)},
		{"category": "C", "value": int64(150)},
	})

	config := DefaultChartConfig()
	config.ChartType = "bar"
	config.Title = "Custom Bar Chart"

	err := InteractiveChart(input, filename, config)
	if err != nil {
		t.Fatalf("InteractiveChart with custom config failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Custom chart file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read chart file: %v", err)
	}

	htmlContent := string(content)

	if !strings.Contains(htmlContent, "Custom Bar Chart") {
		t.Error("Chart should contain custom title")
	}
}

func TestInteractiveChartMultipleYFields(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "multi_series.html")

	input := slices.Values([]Record{
		{"month": "Jan", "sales": int64(1000), "profit": int64(200)},
		{"month": "Feb", "sales": int64(1200), "profit": int64(300)},
		{"month": "Mar", "sales": int64(1100), "profit": int64(250)},
	})

	config := DefaultChartConfig()

	err := InteractiveChart(input, filename, config)
	if err != nil {
		t.Fatalf("InteractiveChart with multiple Y fields failed: %v", err)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Multi-series chart file was not created")
	}
}

// ============================================================================
// QUICKCHART TESTS
// ============================================================================

func TestQuickChart(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "quick_chart.html")

	input := slices.Values([]Record{
		{"x": int64(1), "y": int64(10)},
		{"x": int64(2), "y": int64(20)},
		{"x": int64(3), "y": int64(15)},
	})

	err := QuickChart(input, "x", "y", filename)
	if err != nil {
		t.Fatalf("QuickChart failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("QuickChart file was not created")
	}

	// Read and verify basic content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read chart file: %v", err)
	}

	htmlContent := string(content)

	if !strings.Contains(htmlContent, "canvas") {
		t.Error("QuickChart should contain canvas element")
	}
}

func TestQuickChartFloatValues(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "float_chart.html")

	input := slices.Values([]Record{
		{"time": 1.0, "value": 10.5},
		{"time": 2.0, "value": 20.3},
		{"time": 3.0, "value": 15.7},
	})

	err := QuickChart(input, "time", "value", filename)
	if err != nil {
		t.Fatalf("QuickChart with float values failed: %v", err)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Float chart file was not created")
	}
}

// ============================================================================
// TIMESERIES CHART TESTS
// ============================================================================

func TestTimeSeriesChart(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "timeseries.html")

	now := time.Now()
	input := slices.Values([]Record{
		{"timestamp": now, "value": int64(100)},
		{"timestamp": now.Add(1 * time.Hour), "value": int64(150)},
		{"timestamp": now.Add(2 * time.Hour), "value": int64(120)},
	})

	err := TimeSeriesChart(input, "timestamp", []string{"value"}, filename)
	if err != nil {
		t.Fatalf("TimeSeriesChart failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("TimeSeriesChart file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read chart file: %v", err)
	}

	htmlContent := string(content)

	if !strings.Contains(htmlContent, "canvas") {
		t.Error("TimeSeriesChart should contain canvas element")
	}
}

func TestTimeSeriesChartMultipleValues(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "multi_timeseries.html")

	now := time.Now()
	input := slices.Values([]Record{
		{"time": now, "temp": int64(20), "humidity": int64(60)},
		{"time": now.Add(1 * time.Hour), "temp": int64(22), "humidity": int64(65)},
		{"time": now.Add(2 * time.Hour), "temp": int64(21), "humidity": int64(63)},
	})

	err := TimeSeriesChart(input, "time", []string{"temp", "humidity"}, filename)
	if err != nil {
		t.Fatalf("TimeSeriesChart with multiple values failed: %v", err)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Multi-value timeseries chart file was not created")
	}
}

func TestTimeSeriesChartWithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "custom_timeseries.html")

	now := time.Now()
	input := slices.Values([]Record{
		{"date": now, "sales": int64(1000)},
		{"date": now.Add(24 * time.Hour), "sales": int64(1200)},
	})

	config := DefaultChartConfig()
	config.Title = "Daily Sales"
	config.ChartType = "bar"

	err := TimeSeriesChart(input, "date", []string{"sales"}, filename, config)
	if err != nil {
		t.Fatalf("TimeSeriesChart with config failed: %v", err)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Custom timeseries chart file was not created")
	}

	// Verify custom title is in the file
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read chart file: %v", err)
	}

	if !strings.Contains(string(content), "Daily Sales") {
		t.Error("Chart should contain custom title 'Daily Sales'")
	}
}

// ============================================================================
// ERROR HANDLING TESTS
// ============================================================================

func TestInteractiveChartInvalidPath(t *testing.T) {
	input := slices.Values([]Record{
		{"x": int64(1), "y": int64(10)},
	})

	config := DefaultChartConfig()

	err := InteractiveChart(input, "/invalid/path/chart.html", config)
	if err == nil {
		t.Error("InteractiveChart should return error for invalid path")
	}
}

func TestQuickChartEmptyData(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "empty_chart.html")

	empty := func(yield func(Record) bool) {
		// Yield nothing
	}

	err := QuickChart(iter.Seq[Record](empty), "x", "y", filename)
	// Should handle empty data gracefully - may succeed or fail depending on implementation
	// Just ensure it doesn't panic
	if err != nil {
		t.Logf("QuickChart with empty data returned error: %v", err)
	}
}

func TestTimeSeriesChartEmptyData(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "empty_timeseries.html")

	empty := func(yield func(Record) bool) {
		// Yield nothing
	}

	err := TimeSeriesChart(iter.Seq[Record](empty), "time", []string{"value"}, filename)
	// Should handle empty data gracefully
	if err != nil {
		t.Logf("TimeSeriesChart with empty data returned error: %v", err)
	}
}

// ============================================================================
// INTEGRATION TESTS
// ============================================================================

func TestChartPipeline(t *testing.T) {
	tmpDir := t.TempDir()
	chartFile := filepath.Join(tmpDir, "pipeline_chart.html")

	// Create data, filter it, and chart it
	input := slices.Values([]Record{
		{"x": int64(1), "y": int64(10)},
		{"x": int64(2), "y": int64(5)},
		{"x": int64(3), "y": int64(20)},
		{"x": int64(4), "y": int64(8)},
		{"x": int64(5), "y": int64(25)},
	})

	// Filter: only values where y > 10
	filtered := Where(func(r Record) bool {
		y := GetOr(r, "y", int64(0))
		return y > 10
	})(input)

	err := QuickChart(filtered, "x", "y", chartFile)
	if err != nil {
		t.Fatalf("Chart pipeline failed: %v", err)
	}

	if _, err := os.Stat(chartFile); os.IsNotExist(err) {
		t.Error("Pipeline chart file was not created")
	}
}

func TestChartWithGroupedData(t *testing.T) {
	tmpDir := t.TempDir()
	chartFile := filepath.Join(tmpDir, "grouped_chart.html")

	// Create sales data
	input := slices.Values([]Record{
		{"dept": "Eng", "month": "Jan", "sales": int64(1000)},
		{"dept": "Eng", "month": "Feb", "sales": int64(1200)},
		{"dept": "Sales", "month": "Jan", "sales": int64(800)},
		{"dept": "Sales", "month": "Feb", "sales": int64(900)},
	})

	// Group by department and aggregate using Chain instead of Pipe
	grouped := Chain(
		GroupByFields("records", "dept"),
		Aggregate("records", map[string]AggregateFunc{
			"total_sales": Sum("sales"),
		}),
	)(input)

	err := QuickChart(grouped, "dept", "total_sales", chartFile)
	if err != nil {
		t.Fatalf("Grouped chart failed: %v", err)
	}

	if _, err := os.Stat(chartFile); os.IsNotExist(err) {
		t.Error("Grouped chart file was not created")
	}
}

// ============================================================================
// CHART TYPE TESTS
// ============================================================================

func TestDifferentChartTypes(t *testing.T) {
	tmpDir := t.TempDir()

	chartTypes := []string{"line", "bar", "scatter", "pie", "radar"}

	for _, chartType := range chartTypes {
		t.Run(chartType, func(t *testing.T) {
			filename := filepath.Join(tmpDir, chartType+"_chart.html")

			config := DefaultChartConfig()
			config.ChartType = chartType

			// Need to re-collect input for each test since it's consumed
			testInput := slices.Values([]Record{
				{"category": "A", "value": int64(10)},
				{"category": "B", "value": int64(20)},
				{"category": "C", "value": int64(15)},
			})

			err := InteractiveChart(testInput, filename, config)
			if err != nil {
				t.Fatalf("%s chart failed: %v", chartType, err)
			}

			if _, err := os.Stat(filename); os.IsNotExist(err) {
				t.Errorf("%s chart file was not created", chartType)
			}
		})
	}
}

func TestChartWithStringValues(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "string_chart.html")

	// CSV data often comes in as strings
	input := slices.Values([]Record{
		{"x": "1", "y": "10"},
		{"x": "2", "y": "20"},
		{"x": "3", "y": "15"},
	})

	config := DefaultChartConfig()

	err := InteractiveChart(input, filename, config)
	if err != nil {
		t.Fatalf("Chart with string values failed: %v", err)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("String chart file was not created")
	}
}

func TestChartConfigOptions(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "options_chart.html")

	input := slices.Values([]Record{
		{"x": int64(1), "y": int64(10)},
		{"x": int64(2), "y": int64(20)},
	})

	config := DefaultChartConfig()
	config.Title = "Test Options"
	config.Width = 1024
	config.Height = 768
	config.EnableZoom = false
	config.ShowLegend = false

	err := InteractiveChart(input, filename, config)
	if err != nil {
		t.Fatalf("Chart with custom options failed: %v", err)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Options chart file was not created")
	}

	// Verify custom settings in file
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read chart file: %v", err)
	}

	htmlContent := string(content)
	if !strings.Contains(htmlContent, "Test Options") {
		t.Error("Chart should contain custom title")
	}
}
