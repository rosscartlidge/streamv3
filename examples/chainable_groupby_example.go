package main

import (
	"fmt"
	"slices"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🔗 Chainable GroupBy Example")
	fmt.Println("============================\n")

	// Sample data
	salesData := []streamv3.Record{
		streamv3.NewRecord().String("region", "North").String("product", "Laptop").Float("amount", 1200).Build(),
		streamv3.NewRecord().String("region", "South").String("product", "Phone").Float("amount", 800).Build(),
		streamv3.NewRecord().String("region", "North").String("product", "Phone").Float("amount", 900).Build(),
		streamv3.NewRecord().String("region", "East").String("product", "Laptop").Float("amount", 1100).Build(),
		streamv3.NewRecord().String("region", "South").String("product", "Laptop").Float("amount", 1300).Build(),
	}

	fmt.Println("🎯 Demonstrating Chain() with GroupBy + Aggregate + Where")
	fmt.Println("This shows NO TYPE CHANGES - pure iter.Seq[Record] throughout!\n")

	// Chainable pipeline: Group by region -> Aggregate -> Filter high-performing regions
	results := streamv3.Chain(
		streamv3.GroupByFields("sales_data", "region"),
		streamv3.Aggregate("sales_data", map[string]streamv3.AggregateFunc{
			"total_sales": streamv3.Sum("amount"),
			"avg_sales":   streamv3.Avg("amount"),
			"count":       streamv3.Count(),
		}),
		streamv3.Where(func(r streamv3.Record) bool {
			totalSales := streamv3.GetOr(r, "total_sales", 0.0)
			return totalSales >= 2000 // Only high-performing regions
		}),
	)(slices.Values(salesData))

	fmt.Println("High-performing regions (>= $2000 total sales):")
	for result := range results {
		region := streamv3.GetOr(result, "region", "")
		totalSales := streamv3.GetOr(result, "total_sales", 0.0)
		avgSales := streamv3.GetOr(result, "avg_sales", 0.0)
		count := streamv3.GetOr(result, "count", int64(0))

		fmt.Printf("  %s: $%.0f total, $%.0f avg (%d sales)\n",
			region, totalSales, avgSales, count)
	}

	fmt.Println("\n✅ Success! GroupBy is now fully chainable with no type changes!")
	fmt.Println("💡 You can add any Record-based operation before or after GroupBy/Aggregate.")
}