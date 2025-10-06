package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🔀 Real-World Tee Example: E-commerce Analytics")
	fmt.Println("===============================================\n")

	// Create realistic e-commerce transaction data
	transactions := generateTransactionData()
	fmt.Printf("📊 Processing %d transactions...\n\n", len(transactions))

	// Create a stream and split it into multiple analysis pipelines
	stream := streamv3.From(transactions)
	streams := stream.Tee(4) // Split into 4 parallel analysis streams

	fmt.Println("🚀 Running parallel analytics...")
	fmt.Println("==============================")

	// Analysis 1: Revenue Analytics
	fmt.Println("💰 Revenue Analytics:")
	revenueAnalysis(streams[0])

	// Analysis 2: Customer Segmentation
	fmt.Println("\n👥 Customer Segmentation:")
	customerSegmentation(streams[1])

	// Analysis 3: Product Performance
	fmt.Println("\n📦 Product Performance:")
	productPerformance(streams[2])

	// Analysis 4: Geographic Distribution
	fmt.Println("\n🌍 Geographic Distribution:")
	geographicAnalysis(streams[3])

	fmt.Println("\n✨ Key Benefits Demonstrated:")
	fmt.Println("   🔄 Single data pass for multiple analyses")
	fmt.Println("   ⚡ Parallel processing without data duplication")
	fmt.Println("   📊 Independent analytical workflows")
	fmt.Println("   🎯 Efficient resource utilization")
}

func generateTransactionData() []streamv3.Record {
	customers := []string{"Alice Johnson", "Bob Smith", "Carol Davis", "David Wilson", "Eva Brown"}
	products := []string{"Laptop", "Phone", "Headphones", "Tablet", "Watch"}
	regions := []string{"North", "South", "East", "West"}

	var transactions []streamv3.Record

	for i := 0; i < 100; i++ {
		transaction := streamv3.Record{
			"transaction_id": fmt.Sprintf("TXN-%04d", i+1),
			"customer":       customers[i%len(customers)],
			"product":        products[i%len(products)],
			"amount":         float64(50 + (i*17)%500), // Varying amounts
			"quantity":       int64(1 + i%5),           // 1-5 items
			"region":         regions[i%len(regions)],
			"timestamp":      time.Now().Add(-time.Duration(i) * time.Hour).Format("2006-01-02 15:04:05"),
			"customer_tier":  getTier(i),
			"category":       getCategory(products[i%len(products)]),
		}
		transactions = append(transactions, transaction)
	}

	return transactions
}

func getTier(i int) string {
	switch i % 3 {
	case 0:
		return "premium"
	case 1:
		return "standard"
	default:
		return "basic"
	}
}

func getCategory(product string) string {
	switch product {
	case "Laptop", "Tablet":
		return "computers"
	case "Phone":
		return "mobile"
	case "Headphones", "Watch":
		return "accessories"
	default:
		return "other"
	}
}

func revenueAnalysis(stream *streamv3.Stream[streamv3.Record]) {
	// Calculate total revenue, average order value, and transaction count
	var totalRevenue float64
	var transactionCount int64
	maxAmount := 0.0
	minAmount := float64(^uint(0) >> 1) // Max float64

	for record := range stream.Iter() {
		amount := streamv3.GetOr(record, "amount", 0.0)
		totalRevenue += amount
		transactionCount++

		if amount > maxAmount {
			maxAmount = amount
		}
		if amount < minAmount {
			minAmount = amount
		}
	}

	avgOrderValue := totalRevenue / float64(transactionCount)

	fmt.Printf("  💵 Total Revenue: $%.2f\n", totalRevenue)
	fmt.Printf("  📈 Average Order Value: $%.2f\n", avgOrderValue)
	fmt.Printf("  🔢 Transaction Count: %d\n", transactionCount)
	fmt.Printf("  📊 Range: $%.2f - $%.2f\n", minAmount, maxAmount)
}

func customerSegmentation(stream *streamv3.Stream[streamv3.Record]) {
	// Analyze customer tiers and their spending patterns
	tierStats := make(map[string]struct {
		count   int
		revenue float64
	})

	for record := range stream.Iter() {
		tier := streamv3.GetOr(record, "customer_tier", "unknown")
		amount := streamv3.GetOr(record, "amount", 0.0)

		stats := tierStats[tier]
		stats.count++
		stats.revenue += amount
		tierStats[tier] = stats
	}

	for tier, stats := range tierStats {
		avgSpend := stats.revenue / float64(stats.count)
		fmt.Printf("  🎯 %s: %d customers, $%.2f avg spend\n",
			strings.Title(tier), stats.count, avgSpend)
	}
}

func productPerformance(stream *streamv3.Stream[streamv3.Record]) {
	// Analyze top-selling products by revenue and quantity
	productStats := make(map[string]struct {
		revenue  float64
		quantity int64
		count    int
	})

	for record := range stream.Iter() {
		product := streamv3.GetOr(record, "product", "unknown")
		amount := streamv3.GetOr(record, "amount", 0.0)
		quantity := streamv3.GetOr(record, "quantity", int64(0))

		stats := productStats[product]
		stats.revenue += amount
		stats.quantity += quantity
		stats.count++
		productStats[product] = stats
	}

	// Find top performer by revenue
	var topProduct string
	var topRevenue float64

	for product, stats := range productStats {
		if stats.revenue > topRevenue {
			topRevenue = stats.revenue
			topProduct = product
		}
		fmt.Printf("  📱 %s: $%.2f revenue, %d units, %d orders\n",
			product, stats.revenue, stats.quantity, stats.count)
	}

	fmt.Printf("  🏆 Top Performer: %s ($%.2f)\n", topProduct, topRevenue)
}

func geographicAnalysis(stream *streamv3.Stream[streamv3.Record]) {
	// Analyze sales distribution by region
	regionStats := make(map[string]struct {
		revenue float64
		orders  int
	})

	totalRevenue := 0.0

	for record := range stream.Iter() {
		region := streamv3.GetOr(record, "region", "unknown")
		amount := streamv3.GetOr(record, "amount", 0.0)

		stats := regionStats[region]
		stats.revenue += amount
		stats.orders++
		regionStats[region] = stats

		totalRevenue += amount
	}

	for region, stats := range regionStats {
		percentage := (stats.revenue / totalRevenue) * 100
		avgOrder := stats.revenue / float64(stats.orders)
		fmt.Printf("  🗺️  %s: $%.2f (%.1f%%), %d orders, $%.2f avg\n",
			region, stats.revenue, percentage, stats.orders, avgOrder)
	}
}