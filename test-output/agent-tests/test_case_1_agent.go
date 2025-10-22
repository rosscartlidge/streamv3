package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	// Step 1: Create sample CSV data
	csvData := `name,department,salary
Alice,Engineering,95000
Bob,Engineering,85000
Charlie,Sales,75000
Diana,Engineering,120000
Eve,Sales,90000
Frank,Marketing,70000
Grace,Marketing,88000
Henry,Engineering,82000`

	// Write sample data to /tmp/employees.csv
	err := os.WriteFile("/tmp/employees.csv", []byte(csvData), 0644)
	if err != nil {
		log.Fatalf("Failed to write sample CSV: %v", err)
	}
	fmt.Println("Created sample data in /tmp/employees.csv")

	// Step 2: Read CSV data
	data, err := streamv3.ReadCSV("/tmp/employees.csv")
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	// Step 3: Build the processing pipeline
	result := streamv3.Chain(
		// Filter for employees with salary over 80000
		streamv3.Where(func(r streamv3.Record) bool {
			salary := streamv3.GetOr(r, "salary", float64(0))
			return salary > 80000
		}),
		// Group by department
		streamv3.GroupByFields("analysis", "department"),
		// Count employees in each department
		streamv3.Aggregate("analysis", map[string]streamv3.AggregateFunc{
			"employee_count": streamv3.Count(),
		}),
	)(data)

	// Step 4: Print results
	fmt.Println("\nResults - Employees with salary > $80,000 by department:")
	fmt.Println("=========================================================")

	for record := range result {
		department := streamv3.GetOr(record, "department", "Unknown")
		count := streamv3.GetOr(record, "employee_count", int64(0))
		fmt.Printf("Department: %-15s | Employee Count: %d\n", department, count)
	}
}
