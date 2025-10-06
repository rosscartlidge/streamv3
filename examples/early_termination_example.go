package main

import (
	"fmt"
	"time"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🛑 Early Termination Patterns for Infinite Streams")
	fmt.Println("==================================================\n")

	// Test 1: Limit - Limit infinite stream to first N elements
	fmt.Println("📊 Test 1: Limit - Processing First N Elements")
	fmt.Println("===============================================")
	testLimit()

	// Test 2: TakeWhile - Stop when condition becomes false
	fmt.Println("\n🎯 Test 2: TakeWhile - Conditional Processing")
	fmt.Println("==============================================")
	testTakeWhile()

	// Test 3: TakeUntil - Stop when condition becomes true
	fmt.Println("\n⏰ Test 3: TakeUntil - Event-Driven Termination")
	fmt.Println("===============================================")
	testTakeUntil()

	// Test 4: Timeout - Time-based termination
	fmt.Println("\n⌛ Test 4: Timeout - Time-Based Termination")
	fmt.Println("==========================================")
	testTimeout()

	// Test 5: TimeBasedTimeout - Field-based time termination
	fmt.Println("\n📅 Test 5: TimeBasedTimeout - Field Time Termination")
	fmt.Println("====================================================")
	testTimeBasedTimeout()

	// Test 6: SkipWhile/SkipUntil - Conditional skipping
	fmt.Println("\n⏭️  Test 6: SkipWhile/SkipUntil - Conditional Skipping")
	fmt.Println("=====================================================")
	testSkipPatterns()

	fmt.Println("\n✨ Key Benefits of Early Termination Patterns:")
	fmt.Println("===============================================")
	fmt.Println("🛑 Safe exit conditions for infinite streams")
	fmt.Println("⏱️  Time-based processing limits")
	fmt.Println("🎯 Condition-based stream control")
	fmt.Println("💾 Prevents memory exhaustion")
	fmt.Println("🔄 Perfect for real-time data processing")
}

func testLimit() {
	// Simulate infinite sensor readings
	sensorGenerator := func(yield func(streamv3.Record) bool) {
		for i := 1; ; i++ { // Infinite loop
			record := streamv3.Record{
				"reading_id": fmt.Sprintf("SENSOR-%04d", i),
				"value":      20.0 + float64(i%10)*2.5, // Cycling values
				"timestamp":  time.Now().Add(time.Duration(i) * time.Second).Format("15:04:05"),
			}

			if !yield(record) {
				return
			}

			// Add small delay to simulate real-time data
			time.Sleep(10 * time.Millisecond)
		}
	}

	fmt.Println("Processing infinite sensor stream - taking first 5 readings...")

	// Apply Limit to limit to first 5 elements
	limitOp := streamv3.Limit[streamv3.Record](5)
	limitedStream := limitOp(sensorGenerator)

	for record := range limitedStream {
		readingId := streamv3.GetOr(record, "reading_id", "unknown")
		value := streamv3.GetOr(record, "value", 0.0)
		timestamp := streamv3.GetOr(record, "timestamp", "unknown")

		fmt.Printf("  %s: %.1f°C at %s\n", readingId, value, timestamp)
	}

	fmt.Println("  ✅ Successfully terminated infinite stream after 5 elements")
}

func testTakeWhile() {
	// Simulate stock price monitoring
	priceGenerator := func(yield func(streamv3.Record) bool) {
		prices := []float64{100.0, 102.5, 105.2, 108.1, 95.3, 92.8, 98.7, 101.2}

		for i, price := range prices {
			record := streamv3.Record{
				"symbol":    "TECH",
				"price":     price,
				"tick":      i + 1,
				"timestamp": time.Now().Add(time.Duration(i) * time.Minute).Format("15:04"),
			}

			if !yield(record) {
				return
			}
		}
	}

	fmt.Println("Processing stock prices - continue while price >= $100...")

	// Take while price is above $100
	takeWhileOp := streamv3.TakeWhile(func(record streamv3.Record) bool {
		price := streamv3.GetOr(record, "price", 0.0)
		return price >= 100.0
	})

	filteredStream := takeWhileOp(priceGenerator)

	for record := range filteredStream {
		symbol := streamv3.GetOr(record, "symbol", "unknown")
		price := streamv3.GetOr(record, "price", 0.0)
		tick := streamv3.GetOr(record, "tick", 0)
		timestamp := streamv3.GetOr(record, "timestamp", "unknown")

		fmt.Printf("  %s Tick %d: $%.2f at %s\n", symbol, tick, price, timestamp)
	}

	fmt.Println("  ✅ Stopped processing when price dropped below $100")
}

func testTakeUntil() {
	// Simulate system monitoring - stop when error occurs
	systemGenerator := func(yield func(streamv3.Record) bool) {
		statuses := []string{"OK", "OK", "WARNING", "OK", "ERROR", "OK", "OK"}

		for i, status := range statuses {
			record := streamv3.Record{
				"check_id":   fmt.Sprintf("SYS-%03d", i+1),
				"status":     status,
				"cpu_usage":  float64(30 + i*5),
				"timestamp":  time.Now().Add(time.Duration(i) * time.Second).Format("15:04:05"),
			}

			if !yield(record) {
				return
			}
		}
	}

	fmt.Println("Processing system checks - stop when ERROR is encountered...")

	// Take until we encounter an ERROR status
	takeUntilOp := streamv3.TakeUntil(func(record streamv3.Record) bool {
		status := streamv3.GetOr(record, "status", "unknown")
		return status == "ERROR"
	})

	filteredStream := takeUntilOp(systemGenerator)

	for record := range filteredStream {
		checkId := streamv3.GetOr(record, "check_id", "unknown")
		status := streamv3.GetOr(record, "status", "unknown")
		cpuUsage := streamv3.GetOr(record, "cpu_usage", 0.0)
		timestamp := streamv3.GetOr(record, "timestamp", "unknown")

		fmt.Printf("  %s: %s (CPU: %.1f%%) at %s\n", checkId, status, cpuUsage, timestamp)
	}

	fmt.Println("  ✅ Stopped processing when ERROR status was encountered")
}

func testTimeout() {
	// Simulate slow data source
	slowGenerator := func(yield func(streamv3.Record) bool) {
		for i := 1; i <= 10; i++ {
			record := streamv3.Record{
				"data_id": fmt.Sprintf("DATA-%03d", i),
				"value":   i * 10,
			}

			if !yield(record) {
				return
			}

			// Simulate processing delay
			time.Sleep(200 * time.Millisecond)
		}
	}

	fmt.Println("Processing slow data source with 500ms timeout...")

	// Apply timeout of 500ms
	timeoutOp := streamv3.Timeout[streamv3.Record](500 * time.Millisecond)
	timedStream := timeoutOp(slowGenerator)

	startTime := time.Now()
	count := 0

	for record := range timedStream {
		count++
		dataId := streamv3.GetOr(record, "data_id", "unknown")
		value := streamv3.GetOr(record, "value", 0)

		fmt.Printf("  %s: %d\n", dataId, value)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("  ✅ Processed %d elements in %v (terminated by timeout)\n", count, elapsed)
}

func testTimeBasedTimeout() {
	// Simulate time-series data
	timeSeriesGenerator := func(yield func(streamv3.Record) bool) {
		baseTime := time.Now()

		for i := 0; i < 10; i++ {
			record := streamv3.Record{
				"event_id":  fmt.Sprintf("EVT-%03d", i+1),
				"value":     float64(100 + i*5),
				"timestamp": baseTime.Add(time.Duration(i) * time.Second),
			}

			if !yield(record) {
				return
			}
		}
	}

	fmt.Println("Processing time-series data with 5-second time window...")

	// Apply time-based timeout of 5 seconds
	timeBasedOp := streamv3.TimeBasedTimeout("timestamp", 5*time.Second)
	timedStream := timeBasedOp(timeSeriesGenerator)

	for record := range timedStream {
		eventId := streamv3.GetOr(record, "event_id", "unknown")
		value := streamv3.GetOr(record, "value", 0.0)

		if timestamp, exists := record["timestamp"]; exists {
			if t, ok := timestamp.(time.Time); ok {
				fmt.Printf("  %s: %.1f at %s\n", eventId, value, t.Format("15:04:05"))
			} else {
				fmt.Printf("  %s: %.1f\n", eventId, value)
			}
		} else {
			fmt.Printf("  %s: %.1f\n", eventId, value)
		}
	}

	fmt.Println("  ✅ Stopped processing after 5-second time window")
}

func testSkipPatterns() {
	// Simulate log file processing with headers
	logGenerator := func(yield func(streamv3.Record) bool) {
		lines := []string{
			"# Log started at 2024-01-01",
			"# Version 1.0",
			"# Configuration loaded",
			"INFO: Application started",
			"INFO: Database connected",
			"WARN: High memory usage",
			"ERROR: Connection timeout",
			"INFO: Retrying connection",
		}

		for i, line := range lines {
			record := streamv3.Record{
				"line_no": i + 1,
				"content": line,
				"type":    getLogType(line),
			}

			if !yield(record) {
				return
			}
		}
	}

	fmt.Println("Processing log file - skipping header comments...")

	// Skip while lines start with '#' (comments)
	skipWhileOp := streamv3.SkipWhile(func(record streamv3.Record) bool {
		content := streamv3.GetOr(record, "content", "")
		return len(content) > 0 && content[0] == '#'
	})

	filteredStream := skipWhileOp(logGenerator)

	for record := range filteredStream {
		lineNo := streamv3.GetOr(record, "line_no", 0)
		content := streamv3.GetOr(record, "content", "unknown")
		logType := streamv3.GetOr(record, "type", "unknown")

		fmt.Printf("  Line %d [%s]: %s\n", lineNo, logType, content)
	}

	fmt.Println("  ✅ Successfully skipped header comments and processed log entries")
}

func getLogType(line string) string {
	if len(line) > 0 && line[0] == '#' {
		return "COMMENT"
	}
	if len(line) > 4 {
		switch line[:4] {
		case "INFO":
			return "INFO"
		case "WARN":
			return "WARN"
		case "ERRO":
			return "ERROR"
		}
	}
	return "UNKNOWN"
}