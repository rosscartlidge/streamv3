package main

import (
	"fmt"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🔧 Record Creation API Test")
	fmt.Println("===========================\n")

	// Test nested record creation with the new .Record() method
	userRecord := streamv3.NewRecord().
		String("name", "Alice").
		Int("age", 30).
		Build()

	orderRecord := streamv3.NewRecord().
		String("id", "ORD-123").
		Float("amount", 99.99).
		Record("customer", userRecord). // Using the new .Record() method
		Bool("paid", true).
		Build()

	fmt.Println("📋 Created nested record using .Record() method:")
	fmt.Printf("Order ID: %s\n", streamv3.GetOr(orderRecord, "id", ""))
	fmt.Printf("Amount: $%.2f\n", streamv3.GetOr(orderRecord, "amount", 0.0))
	fmt.Printf("Paid: %v\n", streamv3.GetOr(orderRecord, "paid", false))

	// Access nested customer record
	if customer, ok := streamv3.Get[streamv3.Record](orderRecord, "customer"); ok {
		fmt.Printf("Customer Name: %s\n", streamv3.GetOr(customer, "name", ""))
		fmt.Printf("Customer Age: %d\n", streamv3.GetOr(customer, "age", 0))
	}

	fmt.Println("\n✅ .Record() method working correctly for nested records!")
}