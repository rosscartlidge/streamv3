package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	if len(os.Args) < 2 {
		showUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "generate":
		generateData()
	case "filter":
		filterData()
	case "aggregate":
		aggregateData()
	case "format":
		formatOutput()
	case "demo":
		runChainDemo()
	default:
		showUsage()
	}
}

func showUsage() {
	fmt.Println("🔗 JSON Process Chain Demo")
	fmt.Println("==========================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run json_process_chain.go <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  generate  - Generate sample data (CSV → JSON)")
	fmt.Println("  filter    - Filter records (JSON → JSON)")
	fmt.Println("  aggregate - Aggregate data (JSON → JSON)")
	fmt.Println("  format    - Format output (JSON → human readable)")
	fmt.Println("  demo      - Run complete chain demonstration")
	fmt.Println()
	fmt.Println("Chain Example:")
	fmt.Println("  go run json_process_chain.go generate | \\")
	fmt.Println("  go run json_process_chain.go filter | \\")
	fmt.Println("  go run json_process_chain.go aggregate | \\")
	fmt.Println("  go run json_process_chain.go format")
}

// Step 1: Generate sample data and output as JSON
func generateData() {
	fmt.Fprintln(os.Stderr, "🏭 Generating sample sales data...")

	// Create sample sales data with complex types
	tags1 := slices.Values([]string{"electronics", "mobile"})
	tags2 := slices.Values([]string{"electronics", "computer"})
	tags3 := slices.Values([]string{"books", "education"})
	tags4 := slices.Values([]string{"clothing", "winter"})

	salesData := []streamv3.Record{
		streamv3.NewRecord().
			String("id", "SALE-001").
			String("product", "iPhone 15").
			String("category", "electronics").
			Float("price", 999.99).
			Int("quantity", 2).
			String("region", "North").
			StringSeq("tags", tags1).
			Build(),

		streamv3.NewRecord().
			String("id", "SALE-002").
			String("product", "MacBook Pro").
			String("category", "electronics").
			Float("price", 2499.99).
			Int("quantity", 1).
			String("region", "South").
			StringSeq("tags", tags2).
			Build(),

		streamv3.NewRecord().
			String("id", "SALE-003").
			String("product", "Python Guide").
			String("category", "books").
			Float("price", 49.99).
			Int("quantity", 3).
			String("region", "North").
			StringSeq("tags", tags3).
			Build(),

		streamv3.NewRecord().
			String("id", "SALE-004").
			String("product", "Winter Jacket").
			String("category", "clothing").
			Float("price", 199.99).
			Int("quantity", 2).
			String("region", "North").
			StringSeq("tags", tags4).
			Build(),

		streamv3.NewRecord().
			String("id", "SALE-005").
			String("product", "iPad Air").
			String("category", "electronics").
			Float("price", 599.99).
			Int("quantity", 1).
			String("region", "South").
			StringSeq("tags", tags1).
			Build(),
	}

	stream := streamv3.From(salesData)
	err := streamv3.WriteJSONToWriter(stream, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "✅ Generated 5 sales records")
}

// Step 2: Filter high-value sales (>= $500)
func filterData() {
	fmt.Fprintln(os.Stderr, "🔍 Filtering high-value sales (>= $500)...")

	// Read JSON from stdin
	stream := streamv3.ReadJSONFromReader(os.Stdin)

	// Process records manually for simplicity
	var records []streamv3.Record
	for record := range stream {
		price := streamv3.GetOr(record, "price", float64(0))

		// Filter: only high-value items
		if price >= 500.0 {
			// Add calculated fields
			quantity := streamv3.GetOr(record, "quantity", int64(0))
			record["total_value"] = price * float64(quantity)
			record["tier"] = "premium"
			records = append(records, record)
		}
	}

	err := streamv3.WriteJSONToWriter(streamv3.From(records), os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "✅ Filtered to %d premium sales\n", len(records))
}

// Step 3: Aggregate by region
func aggregateData() {
	fmt.Fprintln(os.Stderr, "📊 Aggregating sales by region...")

	// Read JSON from stdin
	stream := streamv3.ReadJSONFromReader(os.Stdin)

	// Collect all records
	var records []streamv3.Record
	for record := range stream {
		records = append(records, record)
	}

	// Group by region and aggregate
	results := streamv3.Chain(
		streamv3.GroupByFields("group_data", "region"),
		streamv3.Aggregate("group_data", map[string]streamv3.AggregateFunc{
			"total_sales":   streamv3.Count(),
			"total_revenue": streamv3.Sum("total_value"),
			"avg_price":     streamv3.Avg("price"),
			"products":      streamv3.Collect("product"),
		}),
	)(slices.Values(records))

	// Convert results to slice and output
	var aggregated []streamv3.Record
	for result := range results {
		aggregated = append(aggregated, result)
	}

	err := streamv3.WriteJSONToWriter(streamv3.From(aggregated), os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "✅ Aggregated into %d regional summaries\n", len(aggregated))
}

// Step 4: Format output for human consumption
func formatOutput() {
	fmt.Fprintln(os.Stderr, "📝 Formatting final report...")

	// Read JSON from stdin
	stream := streamv3.ReadJSONFromReader(os.Stdin)

	fmt.Println("🏪 REGIONAL SALES REPORT")
	fmt.Println("========================")
	fmt.Println()

	for record := range stream {
		region := streamv3.GetOr(record, "region", "Unknown")
		totalSales := streamv3.GetOr(record, "total_sales", int64(0))
		totalRevenue := streamv3.GetOr(record, "total_revenue", float64(0))
		avgPrice := streamv3.GetOr(record, "avg_price", float64(0))

		fmt.Printf("📍 Region: %s\n", region)
		fmt.Printf("   Sales Count: %d\n", totalSales)
		fmt.Printf("   Total Revenue: $%.2f\n", totalRevenue)
		fmt.Printf("   Average Price: $%.2f\n", avgPrice)

		// Show products if available
		if products, ok := streamv3.Get[[]any](record, "products"); ok {
			fmt.Print("   Products: ")
			for i, product := range products {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(product)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	fmt.Fprintln(os.Stderr, "✅ Report generated successfully")
}

// Demo: Run the complete chain in memory
func runChainDemo() {
	fmt.Println("🔗 JSON Process Chain Demo")
	fmt.Println("===========================\n")

	fmt.Println("This demo shows how JSON preserves data across process boundaries.")
	fmt.Println("Each step reads JSON from stdin and outputs JSON to stdout.\n")

	// Step 1: Generate data
	fmt.Println("Step 1: Generate → JSON")
	fmt.Println("Commands: go run json_process_chain.go generate")

	var step1Output strings.Builder
	tags := slices.Values([]string{"electronics", "premium"})
	sampleData := []streamv3.Record{
		streamv3.NewRecord().
			String("id", "DEMO-001").
			String("product", "Premium Laptop").
			Float("price", 1999.99).
			Int("quantity", 1).
			String("region", "North").
			StringSeq("tags", tags).
			Build(),
	}
	streamv3.WriteJSONToWriter(streamv3.From(sampleData), &step1Output)
	fmt.Printf("Output: %s\n", strings.TrimSpace(step1Output.String()))

	// Step 2: Filter data
	fmt.Println("\nStep 2: JSON → Filter → JSON")
	fmt.Println("Commands: ... | go run json_process_chain.go filter")

	step2Input := strings.NewReader(step1Output.String())
	inputStream := streamv3.ReadJSONFromReader(step2Input)

	var filteredRecords []streamv3.Record
	for record := range inputStream {
		price := streamv3.GetOr(record, "price", float64(0))
		if price >= 500.0 {
			quantity := streamv3.GetOr(record, "quantity", int64(0))
			record["total_value"] = price * float64(quantity)
			record["tier"] = "premium"
			filteredRecords = append(filteredRecords, record)
		}
	}

	var step2Output strings.Builder
	streamv3.WriteJSONToWriter(streamv3.From(filteredRecords), &step2Output)
	fmt.Printf("Output: %s\n", strings.TrimSpace(step2Output.String()))

	fmt.Println("\n✅ Key Benefits Demonstrated:")
	fmt.Println("   🔄 Full data preservation: iter.Seq, complex types maintained")
	fmt.Println("   🔗 Perfect process chaining: JSON → Process → JSON")
	fmt.Println("   🧪 Testable: Each step can be tested independently")
	fmt.Println("   🔧 Composable: Mix and match processing steps")
	fmt.Println("   📊 Type safety: JSONString prevents double-encoding")

	fmt.Println("\n🚀 Try the full pipeline:")
	fmt.Println("   go run json_process_chain.go generate | \\")
	fmt.Println("   go run json_process_chain.go filter | \\")
	fmt.Println("   go run json_process_chain.go aggregate | \\")
	fmt.Println("   go run json_process_chain.go format")
}