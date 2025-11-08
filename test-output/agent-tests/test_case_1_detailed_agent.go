package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/rosscartlidge/ssql"
)

func main() {
	// Step 1: Create sample employee data CSV file
	csvPath := "/tmp/employees.csv"
	if err := createSampleData(csvPath); err != nil {
		log.Fatalf("Failed to create sample data: %v", err)
	}
	fmt.Printf("Created sample data at: %s\n\n", csvPath)

	// Step 2: Read employee data from CSV
	employees, err := streamv3.ReadCSV(csvPath)
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	// Step 3: Filter for employees with salary over $80,000
	highSalaryEmployees := streamv3.Where(func(r streamv3.Record) bool {
		// CSV auto-parses numeric values to float64
		salary := streamv3.GetOr(r, "salary", 0.0)
		return salary > 80000
	})(employees)

	// Step 4: Group by department
	grouped := streamv3.GroupByFields("dept_analysis", "department")(highSalaryEmployees)

	// Step 5: Count employees per department
	results := streamv3.Aggregate("dept_analysis", map[string]streamv3.AggregateFunc{
		"employee_count": streamv3.Count(),
		"avg_salary":     streamv3.Avg("salary"),
		"max_salary":     streamv3.Max[float64]("salary"),
		"min_salary":     streamv3.Min[float64]("salary"),
	})(grouped)

	// Step 6: Sort results by employee count (descending)
	sortedResults := streamv3.SortBy(func(r streamv3.Record) float64 {
		count := streamv3.GetOr(r, "employee_count", int64(0))
		return -float64(count) // Negative for descending order
	})(results)

	// Step 7: Display results
	fmt.Println("High-salary employees (>$80,000) by department:")
	fmt.Println("=" + fmt.Sprintf("%60s", "="))
	fmt.Printf("%-20s %10s %15s %15s %15s\n",
		"Department", "Count", "Avg Salary", "Min Salary", "Max Salary")
	fmt.Println("-" + fmt.Sprintf("%60s", "-"))

	for result := range sortedResults {
		dept := streamv3.GetOr(result, "department", "Unknown")
		count := streamv3.GetOr(result, "employee_count", int64(0))
		avgSalary := streamv3.GetOr(result, "avg_salary", 0.0)
		minSalary := streamv3.GetOr(result, "min_salary", 0.0)
		maxSalary := streamv3.GetOr(result, "max_salary", 0.0)

		fmt.Printf("%-20s %10d $%14.2f $%14.2f $%14.2f\n",
			dept, count, avgSalary, minSalary, maxSalary)
	}
}

// createSampleData generates a sample employees.csv file with realistic data
func createSampleData(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	if err := writer.Write([]string{"employee_id", "name", "department", "salary", "hire_date"}); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}

	// Sample employee data with varied salaries across departments
	employees := [][]string{
		// Engineering department (mix of high and low salaries)
		{"E001", "Alice Johnson", "Engineering", "125000", "2020-01-15"},
		{"E002", "Bob Smith", "Engineering", "95000", "2019-03-22"},
		{"E003", "Carol Davis", "Engineering", "110000", "2021-06-10"},
		{"E004", "David Wilson", "Engineering", "75000", "2022-11-05"},
		{"E005", "Eva Brown", "Engineering", "88000", "2020-08-14"},
		{"E006", "Frank Miller", "Engineering", "102000", "2018-04-20"},

		// Sales department (mostly high earners)
		{"S001", "Grace Lee", "Sales", "135000", "2019-02-10"},
		{"S002", "Henry Chen", "Sales", "92000", "2021-09-15"},
		{"S003", "Isabel Garcia", "Sales", "118000", "2020-05-30"},
		{"S004", "Jack Martinez", "Sales", "85000", "2022-01-08"},

		// Marketing department (moderate salaries)
		{"M001", "Karen Taylor", "Marketing", "82000", "2020-07-12"},
		{"M002", "Leo Anderson", "Marketing", "78000", "2021-10-25"},
		{"M003", "Maria Thomas", "Marketing", "95000", "2019-12-03"},
		{"M004", "Nathan White", "Marketing", "70000", "2022-06-18"},

		// HR department (lower to moderate salaries)
		{"H001", "Olivia Harris", "HR", "72000", "2020-03-14"},
		{"H002", "Paul Clark", "HR", "68000", "2021-08-22"},
		{"H003", "Quinn Lewis", "HR", "81000", "2019-11-30"},

		// Finance department (high salaries)
		{"F001", "Rachel Walker", "Finance", "145000", "2018-01-10"},
		{"F002", "Sam Hall", "Finance", "108000", "2020-04-15"},
		{"F003", "Tina Allen", "Finance", "125000", "2019-07-08"},
		{"F004", "Uma Young", "Finance", "95000", "2021-12-20"},
	}

	// Write all employee records
	for _, employee := range employees {
		if err := writer.Write(employee); err != nil {
			return fmt.Errorf("writing record: %w", err)
		}
	}

	return nil
}
