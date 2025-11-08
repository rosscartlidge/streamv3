package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rosscartlidge/ssql"
)

func main() {
	// Create sample product data
	csvData := `product,price
Budget Widget,45
Premium Gadget,650
Mid-Range Gizmo,250
Luxury Device,1200
Economy Tool,75
Standard Kit,180
Deluxe Package,420`

	// Write sample data
	if err := os.WriteFile("/tmp/products.csv", []byte(csvData), 0644); err != nil {
		log.Fatalf("Failed to create sample data: %v", err)
	}

	// Read CSV data
	data, err := streamv3.ReadCSV("/tmp/products.csv")
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	// Add price_tier field based on price ranges
	results := streamv3.Select(func(r streamv3.Record) streamv3.Record {
		price := streamv3.GetOr(r, "price", float64(0))

		var tier string
		switch {
		case price < 100:
			tier = "Budget"
		case price <= 500:
			tier = "Mid"
		default:
			tier = "Premium"
		}

		return streamv3.SetImmutable(r, "price_tier", tier)
	})(data)

	// Print results
	fmt.Println("Products with Price Tiers:")
	fmt.Println("Product\t\t\tPrice\tTier")
	fmt.Println("-------\t\t\t-----\t----")
	for record := range results {
		product := streamv3.GetOr(record, "product", "")
		price := streamv3.GetOr(record, "price", float64(0))
		tier := streamv3.GetOr(record, "price_tier", "")
		fmt.Printf("%s\t$%.2f\t%s\n", product, price, tier)
	}
}
