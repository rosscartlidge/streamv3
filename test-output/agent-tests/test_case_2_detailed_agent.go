package main

import (
	"fmt"
	"log"
	"os"
	"slices"

	streamv3 "github.com/rosscartlidge/streamv3"
)

func main() {
	// Step 1: Generate sample sales data CSV file
	csvPath := "/tmp/sales_data.csv"
	if err := generateSalesData(csvPath); err != nil {
		log.Fatalf("Error generating sales data: %v", err)
	}
	fmt.Printf("Generated sample sales data at: %s\n\n", csvPath)

	// Step 2: Read the CSV file into a sequence of Records
	records, err := streamv3.ReadCSV(csvPath)
	if err != nil {
		log.Fatalf("Error reading CSV: %v", err)
	}

	// Step 3: Calculate revenue for each record (quantity * price)
	// Add a new "revenue" field to each record
	withRevenue := streamv3.Select(func(r streamv3.Record) streamv3.Record {
		quantity := streamv3.GetOr(r, "quantity", int64(0))
		price := streamv3.GetOr(r, "price", float64(0.0))
		revenue := float64(quantity) * price

		// Immutably add the revenue field to the record
		return r.Float("revenue", revenue)
	})

	// Step 4: Group by product name and aggregate total revenue
	// Use SQL-style GroupByFields to group by product and sum revenue
	grouped := streamv3.Pipe(
		withRevenue,
		streamv3.GroupByFields("group_members", "product"),
	)

	// Step 5: Apply aggregation to sum revenue
	aggregated := streamv3.Pipe(
		grouped,
		streamv3.Aggregate("group_members", map[string]streamv3.AggregateFunc{
			"total_revenue": streamv3.Sum("revenue"),
		}),
	)

	// Step 6: Apply the complete pipeline to our records
	results := aggregated(records)

	// Step 7: Convert to slice so we can sort
	resultSlice := slices.Collect(results)

	// Step 8: Sort by total_revenue in descending order (highest first)
	slices.SortFunc(resultSlice, func(a, b streamv3.Record) int {
		aRevenue := streamv3.GetOr(a, "total_revenue", float64(0.0))
		bRevenue := streamv3.GetOr(b, "total_revenue", float64(0.0))

		// Sort descending (b - a)
		if bRevenue > aRevenue {
			return 1
		} else if bRevenue < aRevenue {
			return -1
		}
		return 0
	})

	// Step 9: Take the top 5 products
	top5Limit := streamv3.Limit[streamv3.Record](5)
	top5Results := slices.Collect(top5Limit(slices.Values(resultSlice)))

	// Step 10: Display the results
	fmt.Println("Top 5 Products by Revenue:")
	fmt.Println("==========================")
	fmt.Printf("%-20s %15s\n", "Product", "Total Revenue")
	fmt.Println("--------------------------------------")

	for _, record := range top5Results {
		product := streamv3.GetOr(record, "product", "")
		revenue := streamv3.GetOr(record, "total_revenue", float64(0.0))
		fmt.Printf("%-20s $%14.2f\n", product, revenue)
	}

	fmt.Println("======================================")
}

// generateSalesData creates a sample CSV file with sales transactions
func generateSalesData(path string) error {
	// Sample sales data with product, quantity, and price
	data := `product,quantity,price
Laptop,2,1299.99
Mouse,15,29.99
Keyboard,8,89.99
Monitor,5,349.99
Laptop,3,1299.99
Mouse,25,29.99
Headphones,12,149.99
Keyboard,10,89.99
Monitor,7,349.99
Laptop,1,1299.99
USB Cable,50,12.99
Headphones,8,149.99
Mouse,30,29.99
Monitor,3,349.99
Laptop,4,1299.99
Keyboard,15,89.99
Webcam,6,79.99
USB Cable,40,12.99
Headphones,10,149.99
Mouse,20,29.99
Monitor,8,349.99
Laptop,5,1299.99
Keyboard,12,89.99
Webcam,9,79.99
USB Cable,35,12.99
`

	// Write the CSV data to the file
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		return fmt.Errorf("writing CSV file: %w", err)
	}

	return nil
}
