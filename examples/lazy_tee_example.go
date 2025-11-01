package main

import (
	"fmt"
	"time"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🚀 LazyTee vs Regular Tee Comparison")
	fmt.Println("====================================\n")

	// Test 1: Finite stream - both work fine
	fmt.Println("📊 Test 1: Finite Stream (both Tee methods work)")
	fmt.Println("===============================================")
	testFiniteStream()

	// Test 2: Simulated infinite stream - LazyTee advantage
	fmt.Println("\n🌊 Test 2: Simulated Infinite Stream (LazyTee advantage)")
	fmt.Println("======================================================")
	testInfiniteStream()

	// Test 3: Different consumption speeds
	fmt.Println("\n⏱️  Test 3: Different Consumer Speeds")
	fmt.Println("===================================")
	testDifferentSpeeds()

	fmt.Println("\n✨ Key Differences:")
	fmt.Println("==================")
	fmt.Println("📋 Regular Tee:")
	fmt.Println("   ✅ Perfect for finite streams")
	fmt.Println("   ✅ All consumers get identical data")
	fmt.Println("   ❌ Buffers entire stream in memory")
	fmt.Println("   ❌ Cannot handle infinite streams")
	fmt.Println()
	fmt.Println("⚡ LazyTee:")
	fmt.Println("   ✅ Perfect for infinite streams")
	fmt.Println("   ✅ Bounded memory usage")
	fmt.Println("   ✅ Real-time processing")
	fmt.Println("   ⚠️  May drop data if consumers are too slow")
	fmt.Println("   ⚠️  Uses goroutines and channels")
}

func testFiniteStream() {
	// Small finite dataset
	data := []streamv3.Record{
		streamv3.MakeMutableRecord().Int("id", 1).Int("value", 10).Freeze(),
		streamv3.MakeMutableRecord().Int("id", 2).Int("value", 20).Freeze(),
		streamv3.MakeMutableRecord().Int("id", 3).Int("value", 30).Freeze(),
		streamv3.MakeMutableRecord().Int("id", 4).Int("value", 40).Freeze(),
		streamv3.MakeMutableRecord().Int("id", 5).Int("value", 50).Freeze(),
	}

	stream := streamv3.From(data)

	// Test regular Tee
	fmt.Println("Regular Tee:")
	regularStreams := streamv3.Tee(stream, 2)

	// Consumer 1: Count
	count1 := 0
	for range regularStreams[0] {
		count1++
	}

	// Consumer 2: Sum
	sum1 := 0.0
	for record := range regularStreams[1] {
		sum1 += streamv3.GetOr(record, "value", 0.0)
	}

	fmt.Printf("  Stream 1 count: %d\n", count1)
	fmt.Printf("  Stream 2 sum: %.0f\n", sum1)

	// Test LazyTee
	fmt.Println("LazyTee:")
	stream2 := streamv3.From(data) // Create fresh stream
	lazyStreams := streamv3.LazyTee(stream2, 2)

	// Consumer 1: Count
	count2 := 0
	for range lazyStreams[0] {
		count2++
	}

	// Consumer 2: Sum
	sum2 := 0.0
	for record := range lazyStreams[1] {
		sum2 += streamv3.GetOr(record, "value", 0.0)
	}

	fmt.Printf("  Stream 1 count: %d\n", count2)
	fmt.Printf("  Stream 2 sum: %.0f\n", sum2)
}

func testInfiniteStream() {
	// Create a simulated infinite stream generator
	infiniteGenerator := func(yield func(streamv3.Record) bool) {
		for i := 0; i < 1000; i++ { // Simulate infinite with large number
			record := streamv3.MakeMutableRecord().
				Int("id", int64(i)).
				Float("value", float64(i*2)).
				String("timestamp", time.Now().Add(time.Duration(i)*time.Millisecond).Format("15:04:05.000")).
				Freeze()

			if !yield(record) {
				fmt.Printf("  Generator stopped at record %d\n", i)
				return
			}

			// Simulate real-time data
			if i%100 == 0 && i > 0 {
				time.Sleep(1 * time.Millisecond)
			}
		}
	}

	fmt.Println("Processing simulated infinite stream with LazyTee...")

	// Use LazyTee for infinite stream
	lazyStreams := streamv3.LazyTee(infiniteGenerator, 3)

	// Consumer 1: Count first 50 items
	go func() {
		count := 0
		for range lazyStreams[0] {
			count++
			if count >= 50 {
				fmt.Printf("  Consumer 1: Processed %d records\n", count)
				return
			}
		}
	}()

	// Consumer 2: Sum first 30 values
	go func() {
		sum := 0.0
		count := 0
		for record := range lazyStreams[1] {
			sum += streamv3.GetOr(record, "value", 0.0)
			count++
			if count >= 30 {
				fmt.Printf("  Consumer 2: Sum of %d records = %.0f\n", count, sum)
				return
			}
		}
	}()

	// Consumer 3: Find records with even IDs (first 20)
	go func() {
		evenCount := 0
		total := 0
		for record := range lazyStreams[2] {
			total++
			id := streamv3.GetOr(record, "id", -1)
			if id%2 == 0 {
				evenCount++
			}
			if total >= 20 {
				fmt.Printf("  Consumer 3: Found %d even IDs in %d records\n", evenCount, total)
				return
			}
		}
	}()

	// Give consumers time to process
	time.Sleep(100 * time.Millisecond)
}

func testDifferentSpeeds() {
	// Generator that produces data at steady rate
	slowGenerator := func(yield func(streamv3.Record) bool) {
		for i := 0; i < 20; i++ {
			record := streamv3.MakeMutableRecord().
				Int("batch", int64(i)).
				String("data", fmt.Sprintf("item_%d", i)).
				Freeze()

			if !yield(record) {
				return
			}

			// Simulate steady data production
			time.Sleep(5 * time.Millisecond)
		}
	}

	fmt.Println("Testing different consumer speeds with LazyTee...")

	lazyStreams := streamv3.LazyTee(slowGenerator, 2)

	// Fast consumer
	go func() {
		count := 0
		start := time.Now()
		for range lazyStreams[0] {
			count++
			// Process immediately
		}
		duration := time.Since(start)
		fmt.Printf("  Fast consumer: Processed %d records in %v\n", count, duration)
	}()

	// Slow consumer
	go func() {
		count := 0
		start := time.Now()
		for range lazyStreams[1] {
			count++
			// Simulate slow processing
			time.Sleep(10 * time.Millisecond)
		}
		duration := time.Since(start)
		fmt.Printf("  Slow consumer: Processed %d records in %v\n", count, duration)
	}()

	// Wait for processing to complete
	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n💡 Note: LazyTee may drop data for slow consumers to prevent blocking")
	fmt.Println("    This is a trade-off for handling infinite streams efficiently")
}