package main

import (
	"log"
	"os"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	// Create sample monthly sales data
	csvData := `month,revenue
January,45000
February,52000
March,48000
April,61000
May,58000
June,67000
July,71000
August,69000
September,63000
October,72000
November,78000
December,84000`

	// Write sample data
	if err := os.WriteFile("/tmp/sales_monthly.csv", []byte(csvData), 0644); err != nil {
		log.Fatalf("Failed to create sample data: %v", err)
	}

	// Read CSV data
	data, err := streamv3.ReadCSV("/tmp/sales_monthly.csv")
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	// Group by month and sum revenue
	results := streamv3.Chain(
		streamv3.GroupByFields("monthly_sales", "month"),
		streamv3.Aggregate("monthly_sales", map[string]streamv3.AggregateFunc{
			"total_revenue": streamv3.Sum("revenue"),
		}),
	)(data)

	// Create bar chart
	if err := streamv3.QuickChart(results, "month", "total_revenue", "/tmp/monthly_sales.html"); err != nil {
		log.Fatalf("Failed to create chart: %v", err)
	}

	log.Println("Chart created: /tmp/monthly_sales.html")
	log.Println("Sample data: /tmp/sales_monthly.csv")
}
