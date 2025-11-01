package main

import (
	"fmt"
	"slices"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🔀 StreamV3 Tee Function Demo")
	fmt.Println("=============================\n")

	// Create sample data
	data := []streamv3.Record{
		streamv3.MakeMutableRecord().String("id", "1").String("name", "Alice").Float("score", 95.5).Freeze(),
		streamv3.MakeMutableRecord().String("id", "2").String("name", "Bob").Float("score", 87.2).Freeze(),
		streamv3.MakeMutableRecord().String("id", "3").String("name", "Carol").Float("score", 92.8).Freeze(),
	}

	fmt.Println("Original data:")
	for _, record := range data {
		fmt.Printf("  %v\n", record)
	}

	// Test 1: Tee into 2 streams
	fmt.Println("\n📋 Test 1: Tee into 2 identical streams")
	fmt.Println("======================================")

	streams := streamv3.Tee(slices.Values(data), 2)
	if len(streams) != 2 {
		fmt.Printf("❌ Expected 2 streams, got %d\n", len(streams))
		return
	}

	fmt.Println("Stream 1 contents:")
	for record := range streams[0] {
		fmt.Printf("  %v\n", record)
	}

	fmt.Println("\nStream 2 contents:")
	for record := range streams[1] {
		fmt.Printf("  %v\n", record)
	}

	// Test 2: Tee into 3 streams and process differently
	fmt.Println("\n🔄 Test 2: Tee into 3 streams with different processing")
	fmt.Println("======================================================")

	streams3 := streamv3.Tee(slices.Values(data), 3)

	// Stream 1: Count records
	count := 0
	for range streams3[0] {
		count++
	}
	fmt.Printf("Stream 1 - Total records: %d\n", count)

	// Stream 2: Filter high scores (>90)
	fmt.Println("Stream 2 - High scores (>90):")
	highScores := streamv3.Chain(
		streamv3.Where(func(r streamv3.Record) bool {
			score := streamv3.GetOr(r, "score", 0.0)
			return score > 90.0
		}),
	)(streams3[1])

	for record := range highScores {
		name := streamv3.GetOr(record, "name", "Unknown")
		score := streamv3.GetOr(record, "score", 0.0)
		fmt.Printf("  %s: %.1f\n", name, score)
	}

	// Stream 3: Extract just names
	fmt.Println("Stream 3 - Names only:")
	nameExtractor := streamv3.Select(func(r streamv3.Record) string {
		return streamv3.GetOr(r, "name", "Unknown")
	})
	names := nameExtractor(streams3[2])

	for name := range names {
		fmt.Printf("  %s\n", name)
	}

	// Test 3: Tee with zero streams
	fmt.Println("\n🚫 Test 3: Tee with zero streams")
	fmt.Println("================================")

	zeroStreams := streamv3.Tee(slices.Values(data), 0)
	if zeroStreams == nil {
		fmt.Println("✅ Correctly returned nil for n=0")
	} else {
		fmt.Printf("❌ Expected nil, got %d streams\n", len(zeroStreams))
	}

	// Test 4: Demonstrate parallel processing pattern
	fmt.Println("\n⚡ Test 4: Parallel processing pattern")
	fmt.Println("====================================")

	bigData := make([]streamv3.Record, 0, 1000)
	for i := 0; i < 1000; i++ {
		bigData = append(bigData, streamv3.MakeMutableRecord().
			String("id", fmt.Sprintf("item_%d", i)).
			Float("value", float64(i)).
			Freeze())
	}

	streams2 := streamv3.Tee(slices.Values(bigData), 2)

	// One stream calculates sum
	var sum float64
	for record := range streams2[0] {
		value := streamv3.GetOr(record, "value", 0.0)
		sum += value
	}

	// Another stream counts items
	var itemCount int
	for range streams2[1] {
		itemCount++
	}

	fmt.Printf("Total items: %d\n", itemCount)
	fmt.Printf("Sum of values: %.0f\n", sum)
	fmt.Printf("Average: %.2f\n", sum/float64(itemCount))

	fmt.Println("\n✅ All tests completed successfully!")
	fmt.Println("\n💡 Key Benefits of Tee:")
	fmt.Println("   🔄 Split one stream into multiple identical copies")
	fmt.Println("   ⚡ Enable parallel processing of the same data")
	fmt.Println("   🧮 Apply different transformations to the same source")
	fmt.Println("   📊 Compute multiple aggregations in one pass")
}