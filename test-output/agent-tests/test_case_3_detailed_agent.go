package main

import (
	"fmt"
	"log"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("=== Customer Order Analysis ===")
	fmt.Println()

	// Step 1: Generate sample customer data
	fmt.Println("Step 1: Creating sample customer data...")
	customersFile := "/tmp/customers.csv"
	if err := createCustomersCSV(customersFile); err != nil {
		log.Fatalf("Failed to create customers CSV: %v", err)
	}
	fmt.Printf("✓ Created %s\n", customersFile)

	// Step 2: Generate sample order data
	fmt.Println("Step 2: Creating sample order data...")
	ordersFile := "/tmp/orders.csv"
	if err := createOrdersCSV(ordersFile); err != nil {
		log.Fatalf("Failed to create orders CSV: %v", err)
	}
	fmt.Printf("✓ Created %s\n", ordersFile)
	fmt.Println()

	// Step 3: Read customer data from CSV
	fmt.Println("Step 3: Reading customer data...")
	customers, err := streamv3.ReadCSV(customersFile)
	if err != nil {
		log.Fatalf("Failed to read customers CSV: %v", err)
	}
	fmt.Println("✓ Customer data loaded")

	// Step 4: Read order data from CSV for the join operation
	fmt.Println("Step 4: Reading order data...")
	ordersForJoin, err := streamv3.ReadCSV(ordersFile)
	if err != nil {
		log.Fatalf("Failed to read orders CSV for join: %v", err)
	}

	fmt.Println("Step 5: Joining customer and order data on customer_id...")
	// InnerJoin takes the right stream (orders) and a predicate
	// OnFields creates a predicate that matches records on the specified field(s)
	joined := streamv3.InnerJoin(
		ordersForJoin,
		streamv3.OnFields("customer_id"),
	)(customers)
	fmt.Println("✓ Data joined successfully")
	fmt.Println()

	// Step 6: Group by customer_id and aggregate total spending
	fmt.Println("Step 6: Grouping by customer and calculating total spending...")
	// GroupByFields groups records and stores group members in a sequence field
	// We use "orders" as the sequence field name to hold all orders for each customer
	grouped := streamv3.GroupByFields("orders", "customer_id", "customer_name")(joined)

	// Aggregate applies aggregation functions to the sequence field
	// We calculate the sum of all "amount" values for each customer
	results := streamv3.Aggregate("orders", map[string]streamv3.AggregateFunc{
		"total_spending": streamv3.Sum("amount"),
		"order_count":    streamv3.Count(),
		"avg_order":      streamv3.Avg("amount"),
	})(grouped)
	fmt.Println("✓ Aggregation complete")
	fmt.Println()

	// Step 7: Display results
	fmt.Println("Step 7: Displaying results...")
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║           Customer Spending Analysis Results                  ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("%-15s %-25s %15s %12s %15s\n",
		"Customer ID", "Name", "Total Spending", "Order Count", "Avg Order")
	fmt.Println(strings("─", 85))

	// Iterate through results and display
	for result := range results {
		// Extract fields with type-safe defaults
		// Use int64 for integers (canonical type) and float64 for decimals
		customerID := streamv3.GetOr(result, "customer_id", int64(0))
		customerName := streamv3.GetOr(result, "customer_name", "Unknown")
		totalSpending := streamv3.GetOr(result, "total_spending", float64(0.0))
		orderCount := streamv3.GetOr(result, "order_count", int64(0))
		avgOrder := streamv3.GetOr(result, "avg_order", float64(0.0))

		fmt.Printf("%-15d %-25s $%14.2f %12d $%14.2f\n",
			customerID, customerName, totalSpending, orderCount, avgOrder)
	}

	fmt.Println()
	fmt.Println("✓ Analysis complete!")
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Println("- Joined customer and order data on customer_id")
	fmt.Println("- Grouped by customer")
	fmt.Println("- Calculated total spending, order count, and average order value")
	fmt.Println("- Results displayed above")
}

// createCustomersCSV creates a sample customers CSV file
func createCustomersCSV(filename string) error {
	// Create sample customer records using MutableRecord for efficient building
	customers := []streamv3.Record{
		streamv3.MakeMutableRecord().
			Int("customer_id", int64(1)).
			String("customer_name", "Alice Johnson").
			String("email", "alice@example.com").
			Freeze(),
		streamv3.MakeMutableRecord().
			Int("customer_id", int64(2)).
			String("customer_name", "Bob Smith").
			String("email", "bob@example.com").
			Freeze(),
		streamv3.MakeMutableRecord().
			Int("customer_id", int64(3)).
			String("customer_name", "Charlie Brown").
			String("email", "charlie@example.com").
			Freeze(),
		streamv3.MakeMutableRecord().
			Int("customer_id", int64(4)).
			String("customer_name", "Diana Prince").
			String("email", "diana@example.com").
			Freeze(),
		streamv3.MakeMutableRecord().
			Int("customer_id", int64(5)).
			String("customer_name", "Eve Williams").
			String("email", "eve@example.com").
			Freeze(),
	}

	// Convert slice to iterator using From helper
	customerSeq := streamv3.From(customers)

	// Write to CSV file
	return streamv3.WriteCSV(customerSeq, filename)
}

// createOrdersCSV creates a sample orders CSV file
func createOrdersCSV(filename string) error {
	// Create sample order records
	// Multiple orders per customer to demonstrate aggregation
	orders := []streamv3.Record{
		// Customer 1 (Alice) - 3 orders
		streamv3.MakeMutableRecord().
			Int("order_id", int64(101)).
			Int("customer_id", int64(1)).
			Float("amount", 150.00).
			String("product", "Laptop").
			Freeze(),
		streamv3.MakeMutableRecord().
			Int("order_id", int64(102)).
			Int("customer_id", int64(1)).
			Float("amount", 75.50).
			String("product", "Mouse").
			Freeze(),
		streamv3.MakeMutableRecord().
			Int("order_id", int64(103)).
			Int("customer_id", int64(1)).
			Float("amount", 200.00).
			String("product", "Monitor").
			Freeze(),

		// Customer 2 (Bob) - 2 orders
		streamv3.MakeMutableRecord().
			Int("order_id", int64(104)).
			Int("customer_id", int64(2)).
			Float("amount", 89.99).
			String("product", "Keyboard").
			Freeze(),
		streamv3.MakeMutableRecord().
			Int("order_id", int64(105)).
			Int("customer_id", int64(2)).
			Float("amount", 125.00).
			String("product", "Webcam").
			Freeze(),

		// Customer 3 (Charlie) - 1 order
		streamv3.MakeMutableRecord().
			Int("order_id", int64(106)).
			Int("customer_id", int64(3)).
			Float("amount", 450.00).
			String("product", "Desk Chair").
			Freeze(),

		// Customer 4 (Diana) - 4 orders
		streamv3.MakeMutableRecord().
			Int("order_id", int64(107)).
			Int("customer_id", int64(4)).
			Float("amount", 299.99).
			String("product", "Tablet").
			Freeze(),
		streamv3.MakeMutableRecord().
			Int("order_id", int64(108)).
			Int("customer_id", int64(4)).
			Float("amount", 49.99).
			String("product", "USB Cable").
			Freeze(),
		streamv3.MakeMutableRecord().
			Int("order_id", int64(109)).
			Int("customer_id", int64(4)).
			Float("amount", 179.00).
			String("product", "Headphones").
			Freeze(),
		streamv3.MakeMutableRecord().
			Int("order_id", int64(110)).
			Int("customer_id", int64(4)).
			Float("amount", 89.50).
			String("product", "Phone Case").
			Freeze(),

		// Customer 5 (Eve) - 2 orders
		streamv3.MakeMutableRecord().
			Int("order_id", int64(111)).
			Int("customer_id", int64(5)).
			Float("amount", 599.00).
			String("product", "Smartphone").
			Freeze(),
		streamv3.MakeMutableRecord().
			Int("order_id", int64(112)).
			Int("customer_id", int64(5)).
			Float("amount", 39.99).
			String("product", "Screen Protector").
			Freeze(),
	}

	// Convert slice to iterator
	orderSeq := streamv3.From(orders)

	// Write to CSV file
	return streamv3.WriteCSV(orderSeq, filename)
}

// strings repeats a string n times - helper for formatting
func strings(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
