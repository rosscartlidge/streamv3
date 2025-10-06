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
		streamv3.NewRecord().String("region", "North").String("product", "Tablet").Float("amount", 500).Build(),
		streamv3.NewRecord().String("region", "East").String("product", "Phone").Float("amount", 850).Build(),
		streamv3.NewRecord().String("region", "South").String("product", "Tablet").Float("amount", 450).Build(),
	}

	fmt.Println("📊 Sales Analysis with Functional Composition")
	fmt.Println("==============================================\n")

	// Example 1: Filter high-value sales and group by region - FUNCTIONAL PIPELINE!
	fmt.Println("1. High-value sales (>= $1000) by region:")

	// Step 1: Filter high-value sales - using type-safe Get[T]
	filtered1 := streamv3.Where(func(r streamv3.Record) bool {
		amount := streamv3.GetOr(r, "amount", 0.0)
		return amount >= 1000
	})(slices.Values(sales))

	// Step 2: Group by region
	groups1 := streamv3.GroupByFields("sales_data", "region")(filtered1)

	// Step 3: Aggregate
	regionTotals := streamv3.Aggregate("sales_data", map[string]streamv3.AggregateFunc{
		"total_sales": streamv3.Sum("amount"),
		"count":       streamv3.Count(),
	})(groups1)

	for result := range regionTotals {
		fmt.Printf("   %s: $%.0f (%d sales)\n",
			result["region"],
			result["total_sales"],
			result["count"])
	}

	fmt.Println("\n2. All sales by product category:")

	// Example 2: Group all sales by product and calculate totals - FUNCTIONAL PIPELINE!
	groups2 := streamv3.GroupByFields("sales_data", "product")(slices.Values(sales))
	productTotals := streamv3.Aggregate("sales_data", map[string]streamv3.AggregateFunc{
		"total_revenue": streamv3.Sum("amount"),
		"avg_price":     streamv3.Avg("amount"),
		"sales_count":   streamv3.Count(),
	})(groups2)

	for result := range productTotals {
		fmt.Printf("   %s: $%.0f total, $%.0f avg (%d sales)\n",
			result["product"],
			result["total_revenue"],
			result["avg_price"],
			result["sales_count"])
	}

	fmt.Println("\n3. Chain everything together - North region laptop sales:")

	// Example 3: Complex pipeline with multiple filters and aggregation - FULL FUNCTIONAL PIPELINE!
	filtered3 := streamv3.Chain(
		streamv3.Where(func(r streamv3.Record) bool {
			region := streamv3.GetOr(r, "region", "")
			return region == "North"
		}),
		streamv3.Where(func(r streamv3.Record) bool {
			product := streamv3.GetOr(r, "product", "")
			return product == "Laptop"
		}),
	)(slices.Values(sales))

	groups3 := streamv3.GroupByFields("sales_data", "region")(filtered3)
	northLaptopTotals := streamv3.Aggregate("sales_data", map[string]streamv3.AggregateFunc{
		"total": streamv3.Sum("amount"),
		"count": streamv3.Count(),
	})(groups3)

	hasResults := false
	for result := range northLaptopTotals {
		fmt.Printf("   North region laptop sales: $%.0f (%d sales)\n",
			result["total"], result["count"])
		hasResults = true
	}

	if !hasResults {
		fmt.Println("   No laptop sales found in North region")
	}
}