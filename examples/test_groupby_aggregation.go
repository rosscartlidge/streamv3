package main

import (
	"fmt"
	"slices"
	"strings"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🧪 StreamV3 GroupBy and Aggregation Test")
	fmt.Println("========================================\n")

	// Create sample sales data
	salesData := []streamv3.Record{
		streamv3.MakeMutableRecord().String("region", "North").String("product", "Laptop").Float("amount", 1200).Int("quantity", 2).Freeze(),
		streamv3.MakeMutableRecord().String("region", "South").String("product", "Phone").Float("amount", 800).Int("quantity", 4).Freeze(),
		streamv3.MakeMutableRecord().String("region", "North").String("product", "Phone").Float("amount", 900).Int("quantity", 3).Freeze(),
		streamv3.MakeMutableRecord().String("region", "East").String("product", "Laptop").Float("amount", 1100).Int("quantity", 1).Freeze(),
		streamv3.MakeMutableRecord().String("region", "South").String("product", "Laptop").Float("amount", 1300).Int("quantity", 2).Freeze(),
		streamv3.MakeMutableRecord().String("region", "North").String("product", "Tablet").Float("amount", 500).Int("quantity", 5).Freeze(),
		streamv3.MakeMutableRecord().String("region", "East").String("product", "Phone").Float("amount", 850).Int("quantity", 2).Freeze(),
		streamv3.MakeMutableRecord().String("region", "South").String("product", "Tablet").Float("amount", 450).Int("quantity", 3).Freeze(),
		streamv3.MakeMutableRecord().String("region", "West").String("product", "Laptop").Float("amount", 1400).Int("quantity", 1).Freeze(),
		streamv3.MakeMutableRecord().String("region", "West").String("product", "Phone").Float("amount", 750).Int("quantity", 4).Freeze(),
	}

	fmt.Printf("📊 Sample Data (%d records):\n", len(salesData))
	for i, record := range salesData {
		region := streamv3.GetOr(record, "region", "")
		product := streamv3.GetOr(record, "product", "")
		amount := streamv3.GetOr(record, "amount", 0.0)
		quantity := streamv3.GetOr(record, "quantity", int64(0))
		fmt.Printf("  %d. %s/%s: $%.0f (qty: %d)\n", i+1, region, product, amount, quantity)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))

	// Test 1: Group by single field (region)
	fmt.Println("\n🔍 Test 1: Group by Region")
	fmt.Println(strings.Repeat("-", 25))

	results1 := streamv3.Chain(
		streamv3.GroupByFields("group_data", "region"),
		streamv3.Aggregate("group_data", map[string]streamv3.AggregateFunc{
			"total_sales":   streamv3.Sum("amount"),
			"total_qty":     streamv3.Sum("quantity"),
			"avg_amount":    streamv3.Avg("amount"),
			"count":         streamv3.Count(),
			"max_amount":    streamv3.Max[float64]("amount"),
			"min_amount":    streamv3.Min[float64]("amount"),
		}),
	)(slices.Values(salesData))

	fmt.Println("Results by Region:")
	for result := range results1 {
		region := streamv3.GetOr(result, "region", "")
		totalSales := streamv3.GetOr(result, "total_sales", 0.0)
		totalQty := streamv3.GetOr(result, "total_qty", 0.0)
		avgAmount := streamv3.GetOr(result, "avg_amount", 0.0)
		count := streamv3.GetOr(result, "count", int64(0))
		maxAmount := streamv3.GetOr(result, "max_amount", 0.0)
		minAmount := streamv3.GetOr(result, "min_amount", 0.0)

		fmt.Printf("  %s: $%.0f total, %.0f items, $%.0f avg (min: $%.0f, max: $%.0f, count: %d)\n",
			region, totalSales, totalQty, avgAmount, minAmount, maxAmount, count)
	}

	// Test 2: Group by multiple fields (region + product)
	fmt.Println("\n🔍 Test 2: Group by Region + Product")
	fmt.Println(strings.Repeat("-", 35))

	results2 := streamv3.Chain(
		streamv3.GroupByFields("group_data", "region", "product"),
		streamv3.Aggregate("group_data", map[string]streamv3.AggregateFunc{
			"total_sales": streamv3.Sum("amount"),
			"total_qty":   streamv3.Sum("quantity"),
			"count":       streamv3.Count(),
		}),
	)(slices.Values(salesData))

	fmt.Println("Results by Region + Product:")
	for result := range results2 {
		region := streamv3.GetOr(result, "region", "")
		product := streamv3.GetOr(result, "product", "")
		totalSales := streamv3.GetOr(result, "total_sales", 0.0)
		totalQty := streamv3.GetOr(result, "total_qty", 0.0)
		count := streamv3.GetOr(result, "count", int64(0))

		fmt.Printf("  %s/%s: $%.0f total, %.0f items (%d sales)\n",
			region, product, totalSales, totalQty, count)
	}

	// Test 3: Functional pipeline with filtering + grouping
	fmt.Println("\n🔍 Test 3: High-Value Sales (>= $1000) by Product")
	fmt.Println(strings.Repeat("-", 45))

	highValueSales := streamv3.Where(func(r streamv3.Record) bool {
		amount := streamv3.GetOr(r, "amount", 0.0)
		return amount >= 1000
	})(slices.Values(salesData))

	results3 := streamv3.Chain(
		streamv3.GroupByFields("group_data", "product"),
		streamv3.Aggregate("group_data", map[string]streamv3.AggregateFunc{
			"total_sales": streamv3.Sum("amount"),
			"avg_amount":  streamv3.Avg("amount"),
			"count":       streamv3.Count(),
		}),
	)(highValueSales)

	fmt.Println("High-Value Sales by Product:")
	for result := range results3 {
		product := streamv3.GetOr(result, "product", "")
		totalSales := streamv3.GetOr(result, "total_sales", 0.0)
		avgAmount := streamv3.GetOr(result, "avg_amount", 0.0)
		count := streamv3.GetOr(result, "count", int64(0))

		fmt.Printf("  %s: $%.0f total, $%.0f avg (%d sales)\n",
			product, totalSales, avgAmount, count)
	}

	// Test 4: Complex aggregation with chain
	fmt.Println("\n🔍 Test 4: Complex Pipeline - Non-Tablet Sales by Region")
	fmt.Println(strings.Repeat("-", 55))

	complexPipeline := streamv3.Chain(
		streamv3.Where(func(r streamv3.Record) bool {
			product := streamv3.GetOr(r, "product", "")
			return product != "Tablet"
		}),
		streamv3.Where(func(r streamv3.Record) bool {
			amount := streamv3.GetOr(r, "amount", 0.0)
			return amount >= 800
		}),
	)(slices.Values(salesData))

	results4 := streamv3.Chain(
		streamv3.GroupByFields("group_data", "region"),
		streamv3.Aggregate("group_data", map[string]streamv3.AggregateFunc{
			"total_sales": streamv3.Sum("amount"),
			"avg_amount":  streamv3.Avg("amount"),
			"count":       streamv3.Count(),
			"first_product": streamv3.First("product"),
			"last_product":  streamv3.Last("product"),
		}),
	)(complexPipeline)

	fmt.Println("Non-Tablet High-Value Sales by Region:")
	for result := range results4 {
		region := streamv3.GetOr(result, "region", "")
		totalSales := streamv3.GetOr(result, "total_sales", 0.0)
		avgAmount := streamv3.GetOr(result, "avg_amount", 0.0)
		count := streamv3.GetOr(result, "count", int64(0))
		firstProduct := streamv3.GetOr(result, "first_product", "")
		lastProduct := streamv3.GetOr(result, "last_product", "")

		fmt.Printf("  %s: $%.0f total, $%.0f avg (%d sales) [%s...%s]\n",
			region, totalSales, avgAmount, count, firstProduct, lastProduct)
	}

	fmt.Println("\n✅ All tests completed! You can modify this code to test your ideas.")
	fmt.Println("💡 Try changing grouping fields, aggregations, or filters to experiment.")
}