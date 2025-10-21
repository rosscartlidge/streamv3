package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	// Create sample employee data
	csvData := `name,department,salary
Alice,Engineering,95000
Bob,Engineering,120000
Carol,Sales,75000
Dave,Engineering,85000
Eve,Marketing,90000
Frank,Sales,82000
Grace,Engineering,110000
Henry,Marketing,78000`

	// Write sample data
	if err := os.WriteFile("/tmp/employees.csv", []byte(csvData), 0644); err != nil {
		log.Fatalf("Failed to create sample data: %v", err)
	}

	// Read CSV data
	data, err := streamv3.ReadCSV("/tmp/employees.csv")
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	// Filter for employees with salary over 80000
	highSalaryEmployees := streamv3.Where(func(r streamv3.Record) bool {
		salary := streamv3.GetOr(r, "salary", int64(0))
		return salary > 80000
	})(data)

	// Group by department and count employees
	results := streamv3.Chain(
		streamv3.GroupByFields("analysis", "department"),
		streamv3.Aggregate("analysis", map[string]streamv3.AggregateFunc{
			"employee_count": streamv3.Count(),
		}),
	)(highSalaryEmployees)

	// Print results
	fmt.Println("Departments with high-salary employees:")
	fmt.Println("Department\tCount")
	fmt.Println("----------\t-----")
	for record := range results {
		dept := streamv3.GetOr(record, "department", "")
		count := streamv3.GetOr(record, "employee_count", int64(0))
		fmt.Printf("%s\t\t%d\n", dept, count)
	}
}
