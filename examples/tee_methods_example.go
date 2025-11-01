package main

import (
	"fmt"
	"slices"

	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🔀 StreamV3 Tee Methods Comparison")
	fmt.Println("==================================\n")

	// Create sample data
	data := []streamv3.Record{
		streamv3.MakeMutableRecord().String("id", "1").String("name", "Alice").Float("score", 95.5).Freeze(),
		streamv3.MakeMutableRecord().String("id", "2").String("name", "Bob").Float("score", 87.2).Freeze(),
		streamv3.MakeMutableRecord().String("id", "3").String("name", "Carol").Float("score", 92.8).Freeze(),
	}

	fmt.Println("📋 Method 1: Using standalone Tee function")
	fmt.Println("==========================================")

	// Direct function call on iter.Seq
	streams1 := streamv3.Tee(slices.Values(data), 2)
	fmt.Printf("Created %d streams using standalone Tee function\n", len(streams1))

	// Use first stream
	fmt.Println("Stream 1 (first 2 records):")
	count := 0
	for record := range streams1[0] {
		if count >= 2 {
			break
		}
		name := streamv3.GetOr(record, "name", "Unknown")
		fmt.Printf("  %s\n", name)
		count++
	}

	// Use second stream
	fmt.Println("Stream 2 (count all):")
	total := 0
	for range streams1[1] {
		total++
	}
	fmt.Printf("  Total records: %d\n", total)

	fmt.Println("\n📋 Method 2: Using Stream.Tee() method")
	fmt.Println("=====================================")

	// Create a Stream and use the Tee method
	stream := streamv3.From(data)
	streams2 := streamv3.Tee(stream, 3)
	fmt.Printf("Created %d streams using Stream.Tee() method\n", len(streams2))

	// Use streams with different processing
	fmt.Println("Stream 1 - Names only:")
	for record := range streams2[0] {
		name := streamv3.GetOr(record, "name", "Unknown")
		fmt.Printf("  %s\n", name)
	}

	fmt.Println("Stream 2 - High scores (>90):")
	highScores := streamv3.Chain(
		streamv3.Where(func(r streamv3.Record) bool {
			score := streamv3.GetOr(r, "score", 0.0)
			return score > 90.0
		}),
	)(streams2[1])

	for record := range highScores {
		name := streamv3.GetOr(record, "name", "Unknown")
		score := streamv3.GetOr(record, "score", 0.0)
		fmt.Printf("  %s: %.1f\n", name, score)
	}

	fmt.Println("Stream 3 - Collect all:")
	allRecords := slices.Collect(streams2[2])
	fmt.Printf("  Collected %d records\n", len(allRecords))

	fmt.Println("\n💡 Usage Comparison:")
	fmt.Println("===================")
	fmt.Println("✅ Standalone function: streamv3.Tee(iter.Seq[T], n) []iter.Seq[T]")
	fmt.Println("   - Works directly with any iter.Seq[T]")
	fmt.Println("   - Returns raw iterators for maximum flexibility")
	fmt.Println("   - Good for functional composition")
	fmt.Println()
	fmt.Println("✅ Stream method: stream.Tee(n) []*Stream[T]")
	fmt.Println("   - Convenient when working with Stream objects")
	fmt.Println("   - Returns wrapped Stream objects")
	fmt.Println("   - Enables chaining with other Stream methods")
	fmt.Println()
	fmt.Println("Choose based on your workflow preferences!")
}