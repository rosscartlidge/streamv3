package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/rosscartlidge/streamv3"
)

// This program demonstrates adding a 'price_tier' field to product data
// based on price ranges: Budget (<100), Mid (100-500), Premium (>500)

func main() {
	// Step 1: Generate sample CSV data in /tmp
	csvFile := "/tmp/products.csv"
	if err := generateSampleData(csvFile); err != nil {
		log.Fatalf("Failed to generate sample data: %v", err)
	}
	fmt.Printf("✓ Sample data created at: %s\n\n", csvFile)

	// Step 2: Read product data from CSV
	// ReadCSV returns (iter.Seq[Record], error) - we need to check the error
	data, err := streamv3.ReadCSV(csvFile)
	if err != nil {
		log.Fatalf("Failed to read CSV file: %v", err)
	}
	fmt.Println("✓ Successfully loaded product data")

	// Step 3: Transform the data by adding price_tier field
	// We use Select to transform each record by adding a new field based on price
	enriched := streamv3.Select(func(r streamv3.Record) streamv3.Record {
		// Get the price field (CSV auto-parses to float64)
		price := streamv3.GetOr(r, "price", float64(0))

		// Determine price tier based on price ranges
		var priceTier string
		if price < 100 {
			priceTier = "Budget"
		} else if price <= 500 {
			priceTier = "Mid"
		} else {
			priceTier = "Premium"
		}

		// Create a new record with the additional field
		// Using immutable Record methods (creates a copy with new field)
		return r.String("price_tier", priceTier)
	})(data)

	fmt.Println("✓ Added price_tier field to all products\n")

	// Step 4: Display the results
	// Note: We need to materialize the iterator into a slice since iterators
	// can only be consumed once. We'll collect all records first.
	var results []streamv3.Record
	for record := range enriched {
		results = append(results, record)
	}

	fmt.Println("Product Data with Price Tiers:")
	fmt.Println(strings("-", 80))
	fmt.Printf("%-20s %-15s %-10s %-15s\n", "Product", "Category", "Price", "Price Tier")
	fmt.Println(strings("-", 80))

	// Display all records
	for _, record := range results {
		name := streamv3.GetOr(record, "name", "Unknown")
		category := streamv3.GetOr(record, "category", "Unknown")
		price := streamv3.GetOr(record, "price", float64(0))
		priceTier := streamv3.GetOr(record, "price_tier", "Unknown")

		fmt.Printf("%-20s %-15s $%-9.2f %-15s\n", name, category, price, priceTier)
	}

	fmt.Println(strings("-", 80))
	fmt.Printf("\n✓ Processing complete! Processed %d products.\n", len(results))
}

// generateSampleData creates a CSV file with sample product data
func generateSampleData(filename string) error {
	// Ensure the directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create sample product data with various price points
	// This will demonstrate all three price tiers: Budget, Mid, and Premium
	content := `name,category,price
Wireless Mouse,Electronics,45.99
USB Cable,Electronics,12.50
Mechanical Keyboard,Electronics,150.00
Gaming Headset,Electronics,89.99
Laptop Stand,Accessories,299.00
Monitor 24-inch,Electronics,450.00
External SSD 1TB,Storage,650.00
Webcam HD,Electronics,75.00
Desk Lamp,Furniture,35.00
Office Chair,Furniture,425.00
Ergonomic Mousepad,Accessories,25.00
Laptop Backpack,Accessories,85.00
Thunderbolt Cable,Electronics,120.00
Desktop PC,Electronics,1250.00
Graphics Tablet,Electronics,180.00
Bluetooth Speaker,Electronics,55.00
USB Hub,Electronics,40.00
Phone Stand,Accessories,18.50
Monitor Arm,Accessories,175.00
Cable Organizer,Accessories,15.00
`

	// Write the CSV file
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// strings is a helper function to repeat a string n times
func strings(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
