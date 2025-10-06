package main

import (
	"fmt"
	"slices"
	"strings"
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
		streamv3.NewRecord().String("region", "West").String("product", "Tablet").Float("amount", 600).Build(),
		streamv3.NewRecord().String("region", "North").String("product", "Tablet").Float("amount", 500).Build(),
	}

	fmt.Println("🔥 Functional Composition: Chained Operations")
	fmt.Println("===============================================\n")

	// Step 1: Apply preprocessing filters
	chained := streamv3.Chain(
		streamv3.Where(func(r streamv3.Record) bool {
			amount := streamv3.GetOr(r, "amount", 0.0)
			return amount >= 600 // Filter >= $600
		}),
		streamv3.Where(func(r streamv3.Record) bool {
			product := streamv3.GetOr(r, "product", "")
			return product != "Tablet" // Exclude tablets
		}),
	)(slices.Values(sales))

	// Apply limit separately since it has different type signature
	filtered := streamv3.Limit[streamv3.Record](10)(chained)

	// Step 2: Apply grouping operation
	groups := streamv3.GroupByFields("sales_data", "region")(filtered)

	// Final aggregation step
	results := streamv3.Aggregate("sales_data", map[string]streamv3.AggregateFunc{
		"total_revenue": streamv3.Sum("amount"),
		"avg_deal":      streamv3.Avg("amount"),
		"count":         streamv3.Count(),
	})(groups)

	fmt.Println("High-value non-tablet sales by region:")
	for result := range results {
		fmt.Printf("  %s: $%.0f revenue, $%.0f avg (%d sales)\n",
			result["region"],
			result["total_revenue"],
			result["avg_deal"],
			result["count"])
	}

	fmt.Println("\n✨ What happened:")
	fmt.Println("  1. Chain: Filter high-value sales + Exclude tablets + Limit")
	fmt.Println("  2. GroupByFields: Type change (Record → GroupedRecord)")
	fmt.Println("  3. Aggregate: Type change (GroupedRecord → Record)")
	fmt.Println("\n💡 One clear way to compose operations!")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🚀 Alternative: Step-by-Step Functional Pipeline")

	// Same thing but step-by-step functional composition
	fmt.Println("Same result using step-by-step function calls:")

	// Step by step functional composition
	filtered2 := streamv3.Where(func(r streamv3.Record) bool {
		amount := streamv3.GetOr(r, "amount", 0.0)
		product := streamv3.GetOr(r, "product", "")
		return amount >= 600 && product != "Tablet"
	})(slices.Values(sales))

	grouped2 := streamv3.GroupByFields("sales_data", "region")(filtered2)

	functionalResults := streamv3.Aggregate("sales_data", map[string]streamv3.AggregateFunc{
		"total_revenue": streamv3.Sum("amount"),
		"avg_deal":      streamv3.Avg("amount"),
		"count":         streamv3.Count(),
	})(grouped2)
	for result := range functionalResults {
		fmt.Printf("  %s: $%.0f revenue, $%.0f avg (%d sales)\n",
			result["region"],
			result["total_revenue"],
			result["avg_deal"],
			result["count"])
	}
}