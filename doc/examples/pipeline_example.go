package main

import (
	"fmt"
	"slices"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	// Sample sales data
	sales := []streamv3.Record{
		streamv3.NewRecord().String("region", "North").String("product", "Laptop").Float("amount", 1200).Build(),
		streamv3.NewRecord().String("region", "South").String("product", "Phone").Float("amount", 800).Build(),
		streamv3.NewRecord().String("region", "North").String("product", "Phone").Float("amount", 900).Build(),
		streamv3.NewRecord().String("region", "East").String("product", "Laptop").Float("amount", 1100).Build(),
		streamv3.NewRecord().String("region", "South").String("product", "Laptop").Float("amount", 1300).Build(),
	}

	fmt.Println("📊 StreamV2-Style Pipeline: From → Where → GroupBy → Sum")
	fmt.Println("==========================================================\n")

	// Try StreamV2-style functional composition
	filterStep := streamv3.Where(func(r streamv3.Record) bool {
		amount := streamv3.GetOr(r, "amount", 0.0)
		return amount >= 1000
	})
	groupStep := streamv3.GroupByFields("region")

	// Apply steps manually
	filtered := filterStep(slices.Values(sales))
	groups := groupStep(filtered)

	// Apply aggregation
	aggregationStep := streamv3.Aggregate(map[string]streamv3.AggregateFunc{
		"total_sales": streamv3.Sum("amount"),
		"count":       streamv3.Count(),
	})

	totals := aggregationStep(groups)

	fmt.Println("High-value sales (>= $1000) by region:")
	for result := range totals {
		fmt.Printf("  %s: $%.0f (%d sales)\n",
			result["region"],
			result["total_sales"],
			result["count"])
	}

	fmt.Println("\n✅ Pipeline composition works!")
	fmt.Println("💡 This is the StreamV2-style functional approach!")
	fmt.Println("   Each step is a pure function that can be composed.")
	fmt.Println("   filterStep(data) → groupStep(filtered) → aggregationStep(groups)")
}