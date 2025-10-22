package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	// Create sample products.csv in /tmp
	csvContent := `product,price
Notebook,45.99
Laptop,899.99
Pen,2.50
Monitor,299.99
Desk,650.00
Mouse,25.00
Keyboard,120.00
Chair,450.00
`
	csvPath := "/tmp/products.csv"
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		log.Fatalf("Failed to create CSV file: %v", err)
	}

	// Read the CSV with error handling
	data, err := streamv3.ReadCSV(csvPath)
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	// Transform records to add price_tier field
	transformed := streamv3.Select(func(r streamv3.Record) streamv3.Record {
		// Get existing price field
		price := streamv3.GetOr(r, "price", float64(0))

		// Calculate price tier based on price ranges
		var tier string
		switch {
		case price < 100:
			tier = "Budget"
		case price >= 100 && price <= 500:
			tier = "Mid"
		case price > 500:
			tier = "Premium"
		}

		// Return new record with added price_tier field
		return streamv3.SetImmutable(r, "price_tier", tier)
	})(data)

	// Print results showing product, price, and tier
	fmt.Println("Product Categorization Results:")
	fmt.Println("================================")
	fmt.Printf("%-15s %-10s %-10s\n", "Product", "Price", "Tier")
	fmt.Println("-----------------------------------------------")

	for record := range transformed {
		product := streamv3.GetOr(record, "product", "")
		price := streamv3.GetOr(record, "price", float64(0))
		tier := streamv3.GetOr(record, "price_tier", "")

		fmt.Printf("%-15s $%-9.2f %-10s\n", product, price, tier)
	}
}
