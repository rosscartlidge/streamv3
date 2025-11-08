package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rosscartlidge/ssql"
)

func main() {
	// Create sample customers.csv
	customersCSV := `customer_id,name
1,Alice Johnson
2,Bob Smith
3,Carol White
4,David Brown
`
	if err := os.WriteFile("/tmp/customers.csv", []byte(customersCSV), 0644); err != nil {
		log.Fatalf("Failed to create customers.csv: %v", err)
	}

	// Create sample orders.csv
	ordersCSV := `order_id,customer_id,amount
101,1,150.50
102,2,200.00
103,1,75.25
104,3,300.00
105,2,50.00
106,1,100.00
107,3,125.75
`
	if err := os.WriteFile("/tmp/orders.csv", []byte(ordersCSV), 0644); err != nil {
		log.Fatalf("Failed to create orders.csv: %v", err)
	}

	// Read customers CSV
	customers, err := streamv3.ReadCSV("/tmp/customers.csv")
	if err != nil {
		log.Fatalf("Failed to read customers.csv: %v", err)
	}

	// Read orders CSV
	orders, err := streamv3.ReadCSV("/tmp/orders.csv")
	if err != nil {
		log.Fatalf("Failed to read orders.csv: %v", err)
	}

	// Join customers with orders, group by customer, and calculate total spending
	result := streamv3.Chain(
		// Join orders on customer_id
		streamv3.InnerJoin(orders, streamv3.OnFields("customer_id")),
		// Group by customer_id and name
		streamv3.GroupByFields("analysis", "customer_id", "name"),
		// Aggregate to calculate total spending
		streamv3.Aggregate("analysis", map[string]streamv3.AggregateFunc{
			"total_spending": streamv3.Sum("amount"),
		}),
	)(customers)

	// Print results
	fmt.Println("Customer Spending Report")
	fmt.Println("========================")
	fmt.Printf("%-15s %15s\n", "Customer Name", "Total Spending")
	fmt.Println("----------------------------------------")

	// Iterate through results and print
	for record := range result {
		name := streamv3.GetOr(record, "name", "")
		totalSpending := streamv3.GetOr(record, "total_spending", float64(0))
		fmt.Printf("%-15s $%14.2f\n", name, totalSpending)
	}
}
