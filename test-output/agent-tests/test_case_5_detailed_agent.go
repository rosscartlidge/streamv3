package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/rosscartlidge/streamv3"
)

// This program demonstrates a complete StreamV3 workflow:
// 1. Generate sample monthly sales data
// 2. Read the CSV file
// 3. Group sales by month
// 4. Sum revenue for each month
// 5. Create an interactive bar chart

func main() {
	// Step 1: Create sample sales data CSV file
	csvPath := "/tmp/sales.csv"
	if err := createSampleSalesData(csvPath); err != nil {
		log.Fatalf("Failed to create sample data: %v", err)
	}
	fmt.Printf("Created sample sales data at: %s\n", csvPath)

	// Step 2: Read the CSV file
	// ReadCSV returns an iterator of Records with automatic type parsing
	salesData, err := streamv3.ReadCSV(csvPath)
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}
	fmt.Println("Successfully read sales data from CSV")

	// Step 3 & 4: Group by month and sum revenue
	// GroupByFields creates groups with a sequence field containing group members
	// Aggregate applies aggregation functions (Sum, Count, Avg, etc.) to each group
	monthlySummary := streamv3.Aggregate("sales", map[string]streamv3.AggregateFunc{
		"total_revenue": streamv3.Sum("revenue"),
		"num_sales":     streamv3.Count(),
		"avg_revenue":   streamv3.Avg("revenue"),
	})(streamv3.GroupByFields("sales", "month")(salesData))

	fmt.Println("\nMonthly Sales Summary:")
	fmt.Println("Month          | Total Revenue | Number of Sales | Average Revenue")
	fmt.Println("---------------|---------------|-----------------|------------------")

	// Display the aggregated results
	// We'll also collect them for charting
	var chartData []streamv3.Record
	for record := range monthlySummary {
		month := streamv3.GetOr(record, "month", "")
		totalRevenue := streamv3.GetOr(record, "total_revenue", float64(0))
		numSales := streamv3.GetOr(record, "num_sales", int64(0))
		avgRevenue := streamv3.GetOr(record, "avg_revenue", float64(0))

		fmt.Printf("%-14s | $%12.2f | %15d | $%15.2f\n",
			month, totalRevenue, numSales, avgRevenue)

		// Save for chart
		chartData = append(chartData, record)
	}

	// Step 5: Create an interactive bar chart
	// QuickChart is the easiest way to visualize data
	chartPath := "/tmp/monthly_sales_detailed.html"

	// Convert chartData to iterator for charting
	chartIterator := streamv3.From(chartData)

	// Configure chart for bar chart display
	config := streamv3.DefaultChartConfig()
	config.Title = "Monthly Sales Revenue Analysis"
	config.ChartType = "bar"
	config.Theme = "light"
	config.EnableCalculations = true
	config.ShowLegend = true
	config.Height = 600
	config.Width = 1200

	if err := streamv3.InteractiveChart(chartIterator, chartPath, config); err != nil {
		log.Fatalf("Failed to create chart: %v", err)
	}

	fmt.Printf("\n✓ Successfully created interactive bar chart at: %s\n", chartPath)
	fmt.Println("✓ Open the HTML file in your browser to view the chart")
	fmt.Println("✓ The chart includes:")
	fmt.Println("  - Interactive field selection (X-axis: month, Y-axis: revenue metrics)")
	fmt.Println("  - Zoom and pan controls")
	fmt.Println("  - Export to PNG and CSV")
	fmt.Println("  - Statistical overlays (trend lines, moving averages)")
}

// createSampleSalesData generates a CSV file with sample monthly sales data
func createSampleSalesData(filename string) error {
	// Ensure the directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	// Create the CSV file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	// Sample sales data for 12 months
	// Each row represents a sale with: month, product, revenue, quantity
	data := `month,product,revenue,quantity
January,Widget A,1250.50,25
January,Widget B,890.75,15
January,Widget C,2340.25,47
January,Widget A,1100.00,22
January,Widget B,950.50,16
February,Widget A,1450.75,29
February,Widget B,1120.25,19
February,Widget C,2890.50,58
February,Widget A,1300.00,26
February,Widget D,780.25,13
March,Widget A,1680.50,34
March,Widget B,1340.75,22
March,Widget C,3120.00,62
March,Widget A,1520.25,30
March,Widget B,1180.50,20
April,Widget A,1890.75,38
April,Widget B,1560.25,26
April,Widget C,3450.50,69
April,Widget D,920.00,15
April,Widget A,1720.75,34
May,Widget A,2100.50,42
May,Widget B,1780.25,30
May,Widget C,3890.75,78
May,Widget A,1940.00,39
May,Widget D,1050.50,18
June,Widget A,2340.25,47
June,Widget B,1980.75,33
June,Widget C,4230.50,85
June,Widget A,2180.00,44
June,Widget B,1820.25,31
July,Widget A,2580.75,52
July,Widget B,2190.50,37
July,Widget C,4560.25,91
July,Widget D,1180.75,20
July,Widget A,2420.00,48
August,Widget A,2890.50,58
August,Widget B,2450.25,41
August,Widget C,4980.75,100
August,Widget A,2680.00,54
August,Widget D,1340.50,22
September,Widget A,2720.25,54
September,Widget B,2340.75,39
September,Widget C,4680.50,94
September,Widget A,2560.00,51
September,Widget B,2180.25,37
October,Widget A,2450.75,49
October,Widget B,2120.50,35
October,Widget C,4340.25,87
October,Widget D,1260.75,21
October,Widget A,2320.00,46
November,Widget A,2180.50,44
November,Widget B,1890.25,32
November,Widget C,3980.75,80
November,Widget A,2050.00,41
November,Widget D,1120.50,19
December,Widget A,1920.75,38
December,Widget B,1680.50,28
December,Widget C,3560.25,71
December,Widget A,1820.00,36
December,Widget B,1580.75,26
`

	_, err = file.WriteString(data)
	if err != nil {
		return fmt.Errorf("writing data: %w", err)
	}

	return nil
}
