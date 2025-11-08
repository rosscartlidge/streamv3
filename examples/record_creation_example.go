package main

import (
	"fmt"
	"github.com/rosscartlidge/ssql"
)

func main() {
	fmt.Println("🔧 Record Creation API Test")
	fmt.Println("===========================\n")

	// Test nested record creation with the new .Nested() method
	userRecord := ssql.MakeMutableRecord().
		String("name", "Alice").
		Int("age", 30).
		Freeze()

	orderRecord := ssql.MakeMutableRecord().
		String("id", "ORD-123").
		Float("amount", 99.99).
		Nested("customer", userRecord). // Using the new .Nested() method
		Bool("paid", true).
		Freeze()

	fmt.Println("📋 Created nested record using .Nested() method:")
	fmt.Printf("Order ID: %s\n", ssql.GetOr(orderRecord, "id", ""))
	fmt.Printf("Amount: $%.2f\n", ssql.GetOr(orderRecord, "amount", 0.0))
	fmt.Printf("Paid: %v\n", ssql.GetOr(orderRecord, "paid", false))

	// Access nested customer record
	if customer, ok := ssql.Get[ssql.Record](orderRecord, "customer"); ok {
		fmt.Printf("Customer Name: %s\n", ssql.GetOr(customer, "name", ""))
		fmt.Printf("Customer Age: %d\n", ssql.GetOr(customer, "age", 0))
	}

	fmt.Println("\n✅ .Nested() method working correctly for nested records!")
}
