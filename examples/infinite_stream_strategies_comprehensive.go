package main

import (
	"fmt"
	"time"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🌊 StreamV3 Infinite Stream Strategies - Comprehensive Demo")
	fmt.Println("===========================================================\n")

	// Demonstrate all four strategies working together
	fmt.Println("🎯 Strategy Integration: Real-World IoT Data Processing")
	fmt.Println("======================================================")
	testIntegratedStrategies()

	// Individual strategy demonstrations
	fmt.Println("\n📊 Strategy 1: Windowing & Chunking")
	fmt.Println("===================================")
	testWindowingStrategy()

	fmt.Println("\n🚀 Strategy 2: LazyTee Broadcasting")
	fmt.Println("==================================")
	testLazyTeeStrategy()

	fmt.Println("\n📈 Strategy 3: Streaming Aggregations")
	fmt.Println("====================================")
	testStreamingAggregationsStrategy()

	fmt.Println("\n🛑 Strategy 4: Early Termination")
	fmt.Println("================================")
	testEarlyTerminationStrategy()

	fmt.Println("\n✨ Summary: Complete Infinite Stream Solution")
	fmt.Println("=============================================")
	printStrategySummary()
}

func testIntegratedStrategies() {
	fmt.Println("Simulating infinite IoT sensor data stream...")
	fmt.Println("Combining all strategies for robust real-time processing:")

	// Create infinite IoT data generator
	iotDataGenerator := func(yield func(streamv3.Record) bool) {
		for i := 1; ; i++ { // Infinite stream
			record := streamv3.Record{
				"sensor_id":   fmt.Sprintf("SENSOR-%03d", (i%10)+1),
				"temperature": 20.0 + float64(i%50)*0.5, // Simulated sensor reading
				"humidity":    45.0 + float64(i%30)*1.2,
				"timestamp":   time.Now().Add(time.Duration(i) * time.Second),
				"batch_id":    i,
				"status":      getRandomStatus(i),
			}

			if !yield(record) {
				return
			}

			// Simulate real-time delay
			time.Sleep(50 * time.Millisecond)
		}
	}

	fmt.Println("\n🌊 Step 1: Apply Early Termination (Take 20 readings)")
	// Strategy 4: Early Termination - Limit infinite stream
	limitedStream := streamv3.Take[streamv3.Record](20)(iotDataGenerator)

	fmt.Println("🚀 Step 2: Apply LazyTee (Split into 3 processing pipelines)")
	// Strategy 2: LazyTee - Split stream for parallel processing
	splitStreams := streamv3.LazyTee(limitedStream, 3)

	// Pipeline 1: Windowing + Aggregations for temperature monitoring
	go func() {
		fmt.Println("  📊 Pipeline 1: Temperature analysis with windowing...")

		// Strategy 1: Windowing - Process in chunks of 5
		windowedStream := streamv3.CountWindow[streamv3.Record](5)(splitStreams[0])

		for window := range windowedStream {
			if len(window) > 0 {
				// Strategy 3: Calculate window statistics
				var tempSum, humiditySum float64
				count := len(window)

				for _, record := range window {
					tempSum += streamv3.GetOr(record, "temperature", 0.0)
					humiditySum += streamv3.GetOr(record, "humidity", 0.0)
				}

				avgTemp := tempSum / float64(count)
				avgHumidity := humiditySum / float64(count)

				fmt.Printf("    📈 Window Stats: Avg Temp %.1f°C, Avg Humidity %.1f%% (%d readings)\n",
					avgTemp, avgHumidity, count)
			}
		}
	}()

	// Pipeline 2: Real-time anomaly detection with streaming aggregations
	go func() {
		fmt.Println("  ⚡ Pipeline 2: Real-time anomaly detection...")

		// Strategy 3: Running statistics for anomaly detection
		runningAvgStream := streamv3.RunningAverage("temperature", 5)(splitStreams[1])

		for record := range runningAvgStream {
			currentTemp := streamv3.GetOr(record, "temperature", 0.0)
			movingAvg := streamv3.GetOr(record, "moving_avg", 0.0)
			sensorId := streamv3.GetOr(record, "sensor_id", "unknown")

			// Simple anomaly detection
			if movingAvg > 0 && (currentTemp > movingAvg+5 || currentTemp < movingAvg-5) {
				fmt.Printf("    🚨 ANOMALY: %s temp %.1f°C (avg %.1f°C)\n",
					sensorId, currentTemp, movingAvg)
			} else {
				fmt.Printf("    ✅ Normal: %s temp %.1f°C (avg %.1f°C)\n",
					sensorId, currentTemp, movingAvg)
			}
		}
	}()

	// Pipeline 3: Status monitoring with conditional termination
	go func() {
		fmt.Println("  🔍 Pipeline 3: Status monitoring...")

		// Strategy 4: TakeUntil - Stop if ERROR status detected
		monitoringStream := streamv3.TakeUntil(func(record streamv3.Record) bool {
			status := streamv3.GetOr(record, "status", "unknown")
			return status == "ERROR"
		})(splitStreams[2])

		for record := range monitoringStream {
			sensorId := streamv3.GetOr(record, "sensor_id", "unknown")
			status := streamv3.GetOr(record, "status", "unknown")
			batchId := streamv3.GetOr(record, "batch_id", 0)

			fmt.Printf("    📡 %s [Batch %d]: %s\n", sensorId, batchId, status)
		}

		fmt.Println("    🛑 Monitoring stopped - ERROR detected or stream ended")
	}()

	// Give pipelines time to process
	time.Sleep(3 * time.Second)
	fmt.Println("\n✅ Integrated strategy demonstration completed!")
}

func testWindowingStrategy() {
	fmt.Println("Processing streaming data in time-based windows...")

	// Create time-series generator
	timeSeriesGenerator := func(yield func(streamv3.Record) bool) {
		baseTime := time.Now()
		for i := 0; i < 15; i++ {
			record := streamv3.Record{
				"metric_id": fmt.Sprintf("M%03d", i+1),
				"value":     float64(100 + i*5),
				"timestamp": baseTime.Add(time.Duration(i*30) * time.Second),
			}
			if !yield(record) {
				return
			}
		}
	}

	// Apply 2-minute time windows
	timeWindowOp := streamv3.TimeWindow[streamv3.Record](2*time.Minute, "timestamp")
	windowedStream := timeWindowOp(timeSeriesGenerator)

	for window := range windowedStream {
		fmt.Printf("  📊 Time Window: %d metrics\n", len(window))
		if len(window) > 0 {
			if firstTime, exists := window[0]["timestamp"]; exists {
				if lastTime, exists := window[len(window)-1]["timestamp"]; exists {
					if t1, ok1 := firstTime.(time.Time); ok1 {
						if t2, ok2 := lastTime.(time.Time); ok2 {
							fmt.Printf("    ⏰ From %s to %s\n", t1.Format("15:04:05"), t2.Format("15:04:05"))
						}
					}
				}
			}
		}
	}
}

func testLazyTeeStrategy() {
	fmt.Println("Broadcasting stream to multiple consumers...")

	// Create data generator
	dataGenerator := func(yield func(streamv3.Record) bool) {
		for i := 1; i <= 10; i++ {
			record := streamv3.Record{
				"id":    i,
				"value": i * 10,
				"type":  getDataType(i),
			}
			if !yield(record) {
				return
			}
		}
	}

	// Split into 2 streams
	streams := streamv3.LazyTee(dataGenerator, 2)

	// Consumer 1: Count records
	go func() {
		count := 0
		for record := range streams[0] {
			count++
			id := streamv3.GetOr(record, "id", 0)
			fmt.Printf("  📊 Consumer 1 - Record %d (total: %d)\n", id, count)
		}
	}()

	// Consumer 2: Sum values
	go func() {
		sum := 0
		for record := range streams[1] {
			value := streamv3.GetOr(record, "value", 0)
			sum += value
			id := streamv3.GetOr(record, "id", 0)
			fmt.Printf("  💰 Consumer 2 - Record %d, Running sum: %d\n", id, sum)
		}
	}()

	time.Sleep(1 * time.Second)
}

func testStreamingAggregationsStrategy() {
	fmt.Println("Real-time aggregations on streaming data...")

	// Create metrics generator
	metricsGenerator := func(yield func(streamv3.Record) bool) {
		metrics := []float64{85.5, 92.1, 78.3, 95.7, 88.9, 91.2, 87.6, 93.4}

		for i, metric := range metrics {
			record := streamv3.Record{
				"metric_id": fmt.Sprintf("SYS%03d", i+1),
				"cpu_load":  metric,
				"timestamp": time.Now().Add(time.Duration(i) * time.Second),
			}
			if !yield(record) {
				return
			}
		}
	}

	// Apply running average with window of 3
	runningAvgOp := streamv3.RunningAverage("cpu_load", 3)
	aggregatedStream := runningAvgOp(metricsGenerator)

	for record := range aggregatedStream {
		metricId := streamv3.GetOr(record, "metric_id", "unknown")
		cpuLoad := streamv3.GetOr(record, "cpu_load", 0.0)
		movingAvg := streamv3.GetOr(record, "moving_avg", 0.0)
		windowSize := streamv3.GetOr(record, "window_size", int64(0))

		fmt.Printf("  📈 %s: %.1f%% CPU (3-point avg: %.1f%%, window: %d)\n",
			metricId, cpuLoad, movingAvg, windowSize)
	}
}

func testEarlyTerminationStrategy() {
	fmt.Println("Controlled termination of infinite streams...")

	// Create potentially infinite generator
	potentiallyInfiniteGenerator := func(yield func(streamv3.Record) bool) {
		for i := 1; i <= 1000; i++ { // Large but finite for demo
			record := streamv3.Record{
				"request_id": fmt.Sprintf("REQ%04d", i),
				"response_time": float64(50 + i%100),
				"status_code": getStatusCode(i),
			}
			if !yield(record) {
				return
			}
		}
	}

	// Demonstrate TakeWhile - process while response time < 100ms
	fmt.Println("  🎯 TakeWhile: Process while response time < 100ms")
	takeWhileOp := streamv3.TakeWhile(func(record streamv3.Record) bool {
		responseTime := streamv3.GetOr(record, "response_time", 0.0)
		return responseTime < 100.0
	})

	filteredStream := takeWhileOp(potentiallyInfiniteGenerator)
	count := 0

	for record := range filteredStream {
		count++
		requestId := streamv3.GetOr(record, "request_id", "unknown")
		responseTime := streamv3.GetOr(record, "response_time", 0.0)
		statusCode := streamv3.GetOr(record, "status_code", 0)

		fmt.Printf("    ⚡ %s: %.0fms (status: %d)\n", requestId, responseTime, statusCode)
	}

	fmt.Printf("  ✅ Processed %d requests before response time exceeded threshold\n", count)
}

func printStrategySummary() {
	fmt.Println("🎯 Four-Strategy Approach for Infinite Streams:")
	fmt.Println()
	fmt.Println("1️⃣  **Windowing & Chunking Operations**")
	fmt.Println("   • CountWindow, SlidingCountWindow, TimeWindow, SlidingTimeWindow")
	fmt.Println("   • Bounds memory usage for unbounded data")
	fmt.Println("   • Enables batch processing of infinite streams")
	fmt.Println()
	fmt.Println("2️⃣  **LazyTee Implementation**")
	fmt.Println("   • Channel-based stream broadcasting")
	fmt.Println("   • Backpressure handling for slow consumers")
	fmt.Println("   • Parallel processing pipelines")
	fmt.Println()
	fmt.Println("3️⃣  **Streaming Aggregations**")
	fmt.Println("   • RunningSum, RunningAverage, ExponentialMovingAverage")
	fmt.Println("   • RunningMinMax, RunningCount")
	fmt.Println("   • Real-time analytics without memory accumulation")
	fmt.Println()
	fmt.Println("4️⃣  **Early Termination Patterns**")
	fmt.Println("   • Take, TakeWhile, TakeUntil, Timeout, TimeBasedTimeout")
	fmt.Println("   • SkipWhile, SkipUntil for conditional processing")
	fmt.Println("   • Safe exit conditions for infinite streams")
	fmt.Println()
	fmt.Println("🌊 **Combined Benefits:**")
	fmt.Println("   ✅ Memory-bounded processing of infinite data")
	fmt.Println("   ✅ Real-time analytics and monitoring")
	fmt.Println("   ✅ Robust error handling and termination")
	fmt.Println("   ✅ Parallel processing with backpressure management")
	fmt.Println("   ✅ Production-ready infinite stream processing")
}

// Helper functions for data generation
func getRandomStatus(i int) string {
	statuses := []string{"OK", "OK", "OK", "WARNING", "OK", "OK", "ERROR"}
	return statuses[i%len(statuses)]
}

func getDataType(i int) string {
	types := []string{"sensor", "metric", "event", "log"}
	return types[i%len(types)]
}

func getStatusCode(i int) int {
	codes := []int{200, 200, 200, 404, 200, 500, 200}
	return codes[i%len(codes)]
}