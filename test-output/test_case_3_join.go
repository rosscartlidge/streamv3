package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	// Create sample customer data
	customerData := `customer_id,name
1,Alice
2,Bob
3,Carol`

	// Create sample order data
	orderData := `order_id,customer_id,amount
101,1,150
102,2,200
103,1,300
104,3,120
105,2,180
106,1,220`

	// Write sample data
	if err := os.WriteFile("/tmp/customers.csv", []byte(customerData), 0644); err != nil {
		log.Fatalf("Failed to create customer data: %v", err)
	}
	if err := os.WriteFile("/tmp/orders.csv", []byte(orderData), 0644); err != nil {
		log.Fatalf("Failed to create order data: %v", err)
	}

	// Read CSV data
	customers, err := streamv3.ReadCSV("/tmp/customers.csv")
	if err != nil {
		log.Fatalf("Failed to read customers: %v", err)
	}

	orders, err := streamv3.ReadCSV("/tmp/orders.csv")
	if err != nil {
		log.Fatalf("Failed to read orders: %v", err)
	}

	// Join customers with orders and calculate total spending
	results := streamv3.Chain(
		streamv3.InnerJoin(customers, streamv3.OnFields("customer_id")),
		streamv3.GroupByFields("customer_spending", "customer_id", "name"),
		streamv3.Aggregate("customer_spending", map[string]streamv3.AggregateFunc{
			"total_spending": streamv3.Sum("amount"),
		}),
	)(orders)

	// Print results
	fmt.Println("Customer Spending:")
	fmt.Println("Customer\tTotal Spending")
	fmt.Println("--------\t--------------")
	for record := range results {
		name := streamv3.GetOr(record, "name", "")
		spending := streamv3.GetOr(record, "total_spending", float64(0))
		fmt.Printf("%s\t\t$%.2f\n", name, spending)
	}
}
