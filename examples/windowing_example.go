package main

import (
	"fmt"
	"slices"
	"time"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🪟 StreamV3 Windowing Operations Demo")
	fmt.Println("=====================================\n")

	// Test 1: CountWindow - Fixed-size windows
	fmt.Println("📊 Test 1: CountWindow (groups of 3)")
	fmt.Println("====================================")
	testCountWindow()

	// Test 2: SlidingCountWindow - Overlapping windows
	fmt.Println("\n🔄 Test 2: SlidingCountWindow (window=3, step=1)")
	fmt.Println("===============================================")
	testSlidingCountWindow()

	// Test 3: TimeWindow - Time-based windows
	fmt.Println("\n⏰ Test 3: TimeWindow (1-minute windows)")
	fmt.Println("======================================")
	testTimeWindow()

	// Test 4: SlidingTimeWindow - Overlapping time windows
	fmt.Println("\n⏱️  Test 4: SlidingTimeWindow (2min window, 30s slide)")
	fmt.Println("===================================================")
	testSlidingTimeWindow()

	fmt.Println("\n✨ Key Benefits of Windowing:")
	fmt.Println("============================")
	fmt.Println("🌊 Essential for infinite stream processing")
	fmt.Println("📈 Enables real-time analytics and monitoring")
	fmt.Println("💾 Bounds memory usage for unbounded data")
	fmt.Println("⚡ Supports both count-based and time-based windows")
	fmt.Println("🔄 Overlapping windows for trend analysis")
}

func testCountWindow() {
	// Create test data
	data := []streamv3.Record{
		{"id": 1, "value": 10},
		{"id": 2, "value": 20},
		{"id": 3, "value": 30},
		{"id": 4, "value": 40},
		{"id": 5, "value": 50},
		{"id": 6, "value": 60},
		{"id": 7, "value": 70},
		{"id": 8, "value": 80},
	}

	fmt.Println("Input data:", len(data), "records")

	// Apply CountWindow
	windowOp := streamv3.CountWindow[streamv3.Record](3)
	windows := windowOp(slices.Values(data))

	windowNum := 1
	for window := range windows {
		fmt.Printf("  Window %d: %d records\n", windowNum, len(window))
		for _, record := range window {
			fmt.Printf("    ID: %v, Value: %v\n", record["id"], record["value"])
		}
		windowNum++
	}
}

func testSlidingCountWindow() {
	// Create test data
	data := []streamv3.Record{
		{"id": 1, "value": 100},
		{"id": 2, "value": 200},
		{"id": 3, "value": 300},
		{"id": 4, "value": 400},
		{"id": 5, "value": 500},
	}

	fmt.Println("Input data:", len(data), "records")

	// Apply SlidingCountWindow (window=3, step=1)
	windowOp := streamv3.SlidingCountWindow[streamv3.Record](3, 1)
	windows := windowOp(slices.Values(data))

	windowNum := 1
	for window := range windows {
		fmt.Printf("  Sliding Window %d: %d records\n", windowNum, len(window))
		var ids []any
		for _, record := range window {
			ids = append(ids, record["id"])
		}
		fmt.Printf("    IDs: %v\n", ids)
		windowNum++
	}
}

func testTimeWindow() {
	// Create test data with timestamps
	baseTime := time.Now().Truncate(time.Minute)
	data := []streamv3.Record{
		{"id": 1, "value": 10, "timestamp": baseTime.Format("2006-01-02 15:04:05")},
		{"id": 2, "value": 20, "timestamp": baseTime.Add(15 * time.Second).Format("2006-01-02 15:04:05")},
		{"id": 3, "value": 30, "timestamp": baseTime.Add(30 * time.Second).Format("2006-01-02 15:04:05")},
		{"id": 4, "value": 40, "timestamp": baseTime.Add(70 * time.Second).Format("2006-01-02 15:04:05")}, // Next minute
		{"id": 5, "value": 50, "timestamp": baseTime.Add(90 * time.Second).Format("2006-01-02 15:04:05")},
		{"id": 6, "value": 60, "timestamp": baseTime.Add(130 * time.Second).Format("2006-01-02 15:04:05")}, // Next minute
	}

	fmt.Println("Input data:", len(data), "records with timestamps")

	// Apply TimeWindow (1 minute windows)
	windowOp := streamv3.TimeWindow[streamv3.Record](time.Minute, "timestamp")
	windows := windowOp(slices.Values(data))

	windowNum := 1
	for window := range windows {
		fmt.Printf("  Time Window %d: %d records\n", windowNum, len(window))
		for _, record := range window {
			fmt.Printf("    ID: %v, Time: %v\n", record["id"], record["timestamp"])
		}
		windowNum++
	}
}

func testSlidingTimeWindow() {
	// Create test data spread over time
	baseTime := time.Now().Truncate(time.Minute)
	data := []streamv3.Record{
		{"id": 1, "event": "login", "timestamp": baseTime.Format(time.RFC3339)},
		{"id": 2, "event": "page_view", "timestamp": baseTime.Add(45 * time.Second).Format(time.RFC3339)},
		{"id": 3, "event": "click", "timestamp": baseTime.Add(90 * time.Second).Format(time.RFC3339)},
		{"id": 4, "event": "purchase", "timestamp": baseTime.Add(135 * time.Second).Format(time.RFC3339)},
		{"id": 5, "event": "logout", "timestamp": baseTime.Add(180 * time.Second).Format(time.RFC3339)},
	}

	fmt.Println("Input data:", len(data), "records over 3 minutes")

	// Apply SlidingTimeWindow (2 minute window, 30 second slide)
	windowOp := streamv3.SlidingTimeWindow[streamv3.Record](2*time.Minute, 30*time.Second, "timestamp")
	windows := windowOp(slices.Values(data))

	windowNum := 1
	for window := range windows {
		fmt.Printf("  Sliding Time Window %d: %d records\n", windowNum, len(window))
		for _, record := range window {
			fmt.Printf("    Event: %v at %v\n", record["event"], record["timestamp"])
		}
		windowNum++
	}
}