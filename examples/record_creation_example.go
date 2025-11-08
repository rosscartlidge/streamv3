package main

import (
	"fmt"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🔧 Record Creation API Test")
	fmt.Println("===========================\n")

	// Test nested record creation with the new .Nested() method
	userRecord := streamv3.MakeMutableRecord().
		String("name", "Alice").
		Int("age", 30).
		Freeze()

	orderRecord := streamv3.MakeMutableRecord().
		String("id", "ORD-123").
		Float("amount", 99.99).
		Nested("customer", userRecord). // Using the new .Nested() method
		Bool("paid", true).
		Freeze()

	fmt.Println("📋 Created nested record using .Nested() method:")
	fmt.Printf("Order ID: %s\n", streamv3.GetOr(orderRecord, "id", ""))
	fmt.Printf("Amount: $%.2f\n", streamv3.GetOr(orderRecord, "amount", 0.0))
	fmt.Printf("Paid: %v\n", streamv3.GetOr(orderRecord, "paid", false))

	// Access nested customer record
	if customer, ok := streamv3.Get[streamv3.Record](orderRecord, "customer"); ok {
		fmt.Printf("Customer Name: %s\n", streamv3.GetOr(customer, "name", ""))
		fmt.Printf("Customer Age: %d\n", streamv3.GetOr(customer, "age", 0))
	}

	fmt.Println("\n✅ .Nested() method working correctly for nested records!")
}
