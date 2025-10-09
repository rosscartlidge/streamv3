package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🔄 StreamV3 Safe/Unsafe Filter Mixing Examples")
	fmt.Println("===============================================\n")

	// Pattern 1: Start Normal, Add Error Handling, Continue Normal
	fmt.Println("📋 Pattern 1: Mixed Pipeline (Normal → Safe → Normal)")
	fmt.Println("======================================================")
	demonstrateMixedPipeline()

	// Pattern 2: I/O with Safe, Processing with Normal
	fmt.Println("\n📊 Pattern 2: I/O with Safe, Processing with Normal")
	fmt.Println("===================================================")
	demonstrateIOWithSafe()

	// Pattern 3: Fail-Fast with Unsafe
	fmt.Println("\n⚡ Pattern 3: Fail-Fast with Unsafe")
	fmt.Println("===================================")
	demonstrateFailFast()

	// Pattern 4: Best-Effort Processing
	fmt.Println("\n🎯 Pattern 4: Best-Effort Processing")
	fmt.Println("====================================")
	demonstrateBestEffort()

	fmt.Println("\n✨ Summary: Conversion Utilities")
	fmt.Println("================================")
	printConversionSummary()
}

func demonstrateMixedPipeline() {
	fmt.Println("Processing transaction data with mixed error handling...")

	// Start with normal data
	transactions := []streamv3.Record{
		{"id": "TXN001", "amount_str": "125.50", "category": "electronics"},
		{"id": "TXN002", "amount_str": "invalid", "category": "books"},
		{"id": "TXN003", "amount_str": "89.99", "category": "clothing"},
		{"id": "TXN004", "amount_str": "250.00", "category": "electronics"},
		{"id": "TXN005", "amount_str": "45.bad", "category": "food"},
		{"id": "TXN006", "amount_str": "180.25", "category": "electronics"},
	}

	// Start with normal iterator
	stream := streamv3.From(transactions)

	// Apply normal filter - remove empty IDs
	filtered := streamv3.Where(func(r streamv3.Record) bool {
		return streamv3.GetOr(r, "id", "") != ""
	})(stream)

	// Convert to Safe for error-prone parsing
	safeStream := streamv3.Safe(filtered)

	// Use Safe filter for parsing amounts
	parsed := streamv3.SelectSafe(func(r streamv3.Record) (streamv3.Record, error) {
		amountStr := streamv3.GetOr(r, "amount_str", "")
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return streamv3.Record{}, fmt.Errorf("invalid amount '%s' in record %s",
				amountStr, streamv3.GetOr(r, "id", "unknown"))
		}

		return streamv3.NewRecord().
			String("id", streamv3.GetOr(r, "id", "")).
			Float("amount", amount).
			String("category", streamv3.GetOr(r, "category", "")).
			Build(), nil
	})(safeStream)

	// Convert back to normal, ignoring errors
	cleanData := streamv3.IgnoreErrors(parsed)

	// Continue with normal filters
	final := streamv3.Chain(
		streamv3.Where(func(r streamv3.Record) bool {
			amount := streamv3.GetOr(r, "amount", 0.0)
			return amount > 100.0
		}),
		streamv3.SortBy(func(r streamv3.Record) float64 {
			// Return negative to sort in descending order
			return -streamv3.GetOr(r, "amount", 0.0)
		}),
	)(cleanData)

	// Display results
	fmt.Println("✅ Successfully processed transactions (amount > $100):")
	for record := range final {
		id := streamv3.GetOr(record, "id", "unknown")
		amount := streamv3.GetOr(record, "amount", 0.0)
		category := streamv3.GetOr(record, "category", "unknown")
		fmt.Printf("  • %s: $%.2f (%s)\n", id, amount, category)
	}
	fmt.Println("  ℹ️  Note: Records with invalid amounts were silently skipped")
}

func demonstrateIOWithSafe() {
	fmt.Println("Creating sample CSV file and processing with Safe filters...")

	// Create sample CSV file
	csvContent := `name,age,email
Alice,25,alice@example.com
Bob,invalid,bob@example.com
Carol,30,carol@example.com
David,28,
Eve,35,eve@example.com
Frank,abc,frank@example.com`

	tmpFile := "/tmp/streamv3_safe_example.csv"
	err := os.WriteFile(tmpFile, []byte(csvContent), 0644)
	if err != nil {
		fmt.Printf("❌ Error writing CSV: %v\n", err)
		return
	}
	defer os.Remove(tmpFile)

	// Read CSV with Safe version
	csvStream := streamv3.ReadCSVSafe(tmpFile)

	// Validate records with Safe filter
	validated := streamv3.WhereSafe(func(r streamv3.Record) (bool, error) {
		// Note: CSV parsing auto-converts numeric strings to int64/float64
		// So "25" becomes int64(25). We need to handle both cases.

		// Try to get age as int64 first (already parsed by CSV reader)
		age, ageOk := streamv3.Get[int64](r, "age")
		if !ageOk {
			return false, fmt.Errorf("invalid or missing age for %s",
				streamv3.GetOr(r, "name", "unknown"))
		}

		// Check if email exists
		email := streamv3.GetOr(r, "email", "")
		if email == "" {
			return false, fmt.Errorf("missing email for %s",
				streamv3.GetOr(r, "name", "unknown"))
		}

		// Only keep adults
		return age >= 18, nil
	})(csvStream)

	// Convert to normal for fast processing
	normalStream := streamv3.IgnoreErrors(validated)

	// Process with normal filters
	final := streamv3.Chain(
		streamv3.SortBy(func(r streamv3.Record) string {
			return streamv3.GetOr(r, "name", "")
		}),
	)(normalStream)

	// Display results
	fmt.Println("✅ Valid records (age ≥ 18, valid data):")
	for record := range final {
		name := streamv3.GetOr(record, "name", "unknown")
		age := streamv3.GetOr(record, "age", int64(0))
		email := streamv3.GetOr(record, "email", "unknown")
		fmt.Printf("  • %s (age %d): %s\n", name, age, email)
	}
	fmt.Println("  ℹ️  Note: Invalid records were filtered out during validation")
}

func demonstrateFailFast() {
	fmt.Println("Processing critical financial data with fail-fast error handling...")

	// Create sample CSV with critical data
	csvContent := `account,balance,status
ACC001,1250.50,active
ACC002,3400.75,active
ACC003,890.25,active`

	tmpFile := "/tmp/streamv3_failfast_example.csv"
	err := os.WriteFile(tmpFile, []byte(csvContent), 0644)
	if err != nil {
		fmt.Printf("❌ Error writing CSV: %v\n", err)
		return
	}
	defer os.Remove(tmpFile)

	// Process with fail-fast approach
	result := processFailFast(tmpFile)
	if result != nil {
		fmt.Printf("❌ Processing failed: %v\n", result)
		fmt.Println("  ℹ️  Note: Pipeline stopped immediately on first error")
		return
	}

	fmt.Println("✅ All critical records processed successfully!")

	// Now demonstrate with corrupted data
	fmt.Println("\nTesting fail-fast with corrupted data:")
	corruptedContent := `account,balance,status
ACC001,1250.50,active
ACC002,CORRUPTED,active
ACC003,890.25,active`

	tmpFile2 := "/tmp/streamv3_corrupted_example.csv"
	err = os.WriteFile(tmpFile2, []byte(corruptedContent), 0644)
	if err != nil {
		fmt.Printf("❌ Error writing CSV: %v\n", err)
		return
	}
	defer os.Remove(tmpFile2)

	result = processFailFast(tmpFile2)
	if result != nil {
		fmt.Printf("❌ Processing failed as expected: %v\n", result)
		fmt.Println("  ✅ Fail-fast successfully caught corrupted data!")
	}
}

func processFailFast(filename string) (err error) {
	// Use defer-recover to catch panics from Unsafe
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("critical error: %v", r)
		}
	}()

	// Read CSV with Safe version
	csvStream := streamv3.ReadCSVSafe(filename)

	// Validate and parse with Safe filter
	parsed := streamv3.SelectSafe(func(r streamv3.Record) (streamv3.Record, error) {
		// Note: CSV parsing auto-converts "1250.50" to float64(1250.50)
		balance, balanceOk := streamv3.Get[float64](r, "balance")
		if !balanceOk {
			return streamv3.Record{}, fmt.Errorf("invalid balance in account %s",
				streamv3.GetOr(r, "account", "unknown"))
		}

		return streamv3.NewRecord().
			String("account", streamv3.GetOr(r, "account", "")).
			Float("balance", balance).
			String("status", streamv3.GetOr(r, "status", "")).
			Build(), nil
	})(csvStream)

	// Convert to Unsafe - will panic on any error
	unsafeStream := streamv3.Unsafe(parsed)

	// Process all records - will panic on first error
	count := 0
	for record := range unsafeStream {
		account := streamv3.GetOr(record, "account", "unknown")
		balance := streamv3.GetOr(record, "balance", 0.0)
		fmt.Printf("  ✓ Processed %s: $%.2f\n", account, balance)
		count++
	}

	fmt.Printf("  📊 Total records processed: %d\n", count)
	return nil
}

func demonstrateBestEffort() {
	fmt.Println("Processing multiple data sources with best-effort approach...")

	// Create multiple CSV files with varying quality
	sources := map[string]string{
		"/tmp/streamv3_source1.csv": `product,price,stock
Laptop,999.99,15
Phone,599.99,25
Tablet,399.99,10`,
		"/tmp/streamv3_source2.csv": `product,price,stock
Monitor,299.99,8
Keyboard,invalid,20
Mouse,49.99,50`,
		"/tmp/streamv3_source3.csv": `product,price,stock
Headphones,149.99,30
Speaker,bad_price,12
Webcam,89.99,18`,
	}

	// Create all CSV files
	for path, content := range sources {
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			fmt.Printf("❌ Error writing %s: %v\n", path, err)
			continue
		}
		defer os.Remove(path)
	}

	// Process all sources with best-effort
	var allProducts []streamv3.Record

	for path, _ := range sources {
		filename := strings.TrimPrefix(path, "/tmp/streamv3_")
		fmt.Printf("\n📂 Processing %s...\n", filename)

		// Read CSV with Safe version
		csvStream := streamv3.ReadCSVSafe(path)

		// Parse prices with Safe filter
		parsed := streamv3.SelectSafe(func(r streamv3.Record) (streamv3.Record, error) {
			// Note: CSV parsing auto-converts "999.99" to float64(999.99) and "15" to int64(15)
			price, priceOk := streamv3.Get[float64](r, "price")
			if !priceOk {
				return streamv3.Record{}, fmt.Errorf("invalid price for %s",
					streamv3.GetOr(r, "product", "unknown"))
			}

			stock, stockOk := streamv3.Get[int64](r, "stock")
			if !stockOk {
				return streamv3.Record{}, fmt.Errorf("invalid stock for %s",
					streamv3.GetOr(r, "product", "unknown"))
			}

			return streamv3.NewRecord().
				String("product", streamv3.GetOr(r, "product", "")).
				Float("price", price).
				Int("stock", stock).
				String("source", filename).
				Build(), nil
		})(csvStream)

		// Use IgnoreErrors to collect valid records, skip invalid ones
		validRecords := streamv3.IgnoreErrors(parsed)

		validCount := 0
		for record := range validRecords {
			allProducts = append(allProducts, record)
			product := streamv3.GetOr(record, "product", "unknown")
			price := streamv3.GetOr(record, "price", 0.0)
			stock := streamv3.GetOr(record, "stock", int64(0))
			fmt.Printf("  ✓ %s: $%.2f (stock: %d)\n", product, price, stock)
			validCount++
		}
		fmt.Printf("  📊 Valid records from this source: %d\n", validCount)
	}

	// Display summary
	fmt.Println("\n📈 Final Inventory Summary:")
	fmt.Printf("  • Total valid products collected: %d\n", len(allProducts))

	// Calculate total inventory value
	totalValue := 0.0
	for _, record := range allProducts {
		price := streamv3.GetOr(record, "price", 0.0)
		stock := streamv3.GetOr(record, "stock", int64(0))
		totalValue += price * float64(stock)
	}
	fmt.Printf("  • Total inventory value: $%.2f\n", totalValue)
	fmt.Println("  ℹ️  Note: Invalid records were skipped, processing continued")
}

func printConversionSummary() {
	fmt.Println("🔄 Conversion Utility Reference:")
	fmt.Println()
	fmt.Println("1️⃣  Safe() - iter.Seq[T] → iter.Seq2[T, error]")
	fmt.Println("   • Converts simple iterator to error-aware")
	fmt.Println("   • Never produces errors (wraps values with nil error)")
	fmt.Println("   • Use before error-prone operations")
	fmt.Println()
	fmt.Println("2️⃣  Unsafe() - iter.Seq2[T, error] → iter.Seq[T]")
	fmt.Println("   • Converts error-aware to simple iterator")
	fmt.Println("   • PANICS on any error encountered")
	fmt.Println("   • Use for fail-fast, critical data processing")
	fmt.Println()
	fmt.Println("3️⃣  IgnoreErrors() - iter.Seq2[T, error] → iter.Seq[T]")
	fmt.Println("   • Converts error-aware to simple iterator")
	fmt.Println("   • Silently skips records with errors")
	fmt.Println("   • Use for best-effort, messy data, ETL pipelines")
	fmt.Println()
	fmt.Println("💡 Pattern Selection Guide:")
	fmt.Println("   • Use Safe() when entering error-aware processing")
	fmt.Println("   • Use Unsafe() when errors must stop processing")
	fmt.Println("   • Use IgnoreErrors() when partial data is acceptable")
}
