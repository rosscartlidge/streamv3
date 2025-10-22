package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	// Create sample sales.csv file
	csvPath := "/tmp/sales.csv"
	csvContent := `month,revenue
January,1000
January,1500
January,800
February,2000
February,1200
March,1800
March,2200
March,900
April,1600
April,1900
May,2500
May,2100
June,1700
June,1400
June,2300`

	err := os.WriteFile(csvPath, []byte(csvContent), 0644)
	if err != nil {
		log.Fatalf("Failed to create sample CSV: %v", err)
	}
	fmt.Printf("Created sample data at %s\n", csvPath)

	// Read the CSV file
	data, err := streamv3.ReadCSV(csvPath)
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	// Process the data: group by month and sum revenue
	results := streamv3.Chain(
		streamv3.GroupByFields("sales", "month"),
		streamv3.Aggregate("sales", map[string]streamv3.AggregateFunc{
			"total_revenue": streamv3.Sum("revenue"),
		}),
	)(data)

	// Create a bar chart
	chartPath := "/tmp/monthly_sales.html"
	err = streamv3.QuickChart(results, "month", "total_revenue", chartPath)
	if err != nil {
		log.Fatalf("Failed to create chart: %v", err)
	}

	fmt.Printf("✓ Successfully processed sales data\n")
	fmt.Printf("✓ Chart created at: %s\n", chartPath)
	fmt.Printf("✓ Open the file in a browser to view the interactive bar chart\n")
}
