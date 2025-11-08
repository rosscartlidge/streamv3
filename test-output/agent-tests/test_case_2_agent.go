package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rosscartlidge/ssql"
)

func main() {
	// Step 1: Create sample sales CSV data
	csvData := `product,revenue
Laptop,1200.50
Phone,850.75
Laptop,1100.00
Tablet,650.25
Phone,920.00
Laptop,1350.80
Desktop,1500.00
Tablet,700.50
Phone,880.25
Desktop,1450.75
Laptop,1250.00
Phone,900.00
Tablet,680.00
Desktop,1600.25
Laptop,1300.50`

	// Write sample data to temporary file
	tmpFile := "/tmp/sales.csv"
	if err := os.WriteFile(tmpFile, []byte(csvData), 0644); err != nil {
		log.Fatalf("Failed to write sample CSV: %v", err)
	}
	defer os.Remove(tmpFile)

	// Step 2: Read the CSV with error handling
	data, err := streamv3.ReadCSV(tmpFile)
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	// Step 3-6: Group by product, sum revenue, sort descending, limit to top 5
	result := streamv3.Chain(
		// Group by product name (namespace "sales")
		streamv3.GroupByFields("sales", "product"),
		// Aggregate total revenue for each product (SAME namespace!)
		streamv3.Aggregate("sales", map[string]streamv3.AggregateFunc{
			"total_revenue": streamv3.Sum("revenue"),
		}),
		// Sort by revenue descending (negative = descending order)
		streamv3.SortBy(func(r streamv3.Record) float64 {
			return -streamv3.GetOr(r, "total_revenue", float64(0))
		}),
		// Limit to top 5 products
		streamv3.Limit[streamv3.Record](5),
	)(data)

	// Step 7: Print results to console
	fmt.Println("Top 5 Products by Revenue:")
	fmt.Println("==========================")
	fmt.Printf("%-15s %15s\n", "Product", "Total Revenue")
	fmt.Println("----------------------------------------")

	for record := range result {
		product := streamv3.GetOr(record, "product", "")
		revenue := streamv3.GetOr(record, "total_revenue", float64(0))
		fmt.Printf("%-15s $%14.2f\n", product, revenue)
	}
}
