package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	// Create sample sales data
	csvData := `product,revenue
Widget,15000
Gadget,23000
Gizmo,8500
Doohickey,31000
Thingamajig,12000
Whatchamacallit,19000
Contraption,7200
Device,28000
Apparatus,9800
Machine,34000`

	// Write sample data
	if err := os.WriteFile("/tmp/sales.csv", []byte(csvData), 0644); err != nil {
		log.Fatalf("Failed to create sample data: %v", err)
	}

	// Read CSV data
	data, err := streamv3.ReadCSV("/tmp/sales.csv")
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	// Group by product name and sum revenue, then get top 5
	results := streamv3.Chain(
		streamv3.GroupByFields("sales", "product"),
		streamv3.Aggregate("sales", map[string]streamv3.AggregateFunc{
			"total_revenue": streamv3.Sum("revenue"),
		}),
		streamv3.SortBy(func(r streamv3.Record) float64 {
			return -streamv3.GetOr(r, "total_revenue", float64(0))
		}),
		streamv3.Limit[streamv3.Record](5),
	)(data)

	// Print results
	fmt.Println("Top 5 Products by Revenue:")
	fmt.Println("Product\t\t\tRevenue")
	fmt.Println("-------\t\t\t-------")
	for record := range results {
		product := streamv3.GetOr(record, "product", "")
		revenue := streamv3.GetOr(record, "total_revenue", float64(0))
		fmt.Printf("%s\t\t$%.2f\n", product, revenue)
	}
}
