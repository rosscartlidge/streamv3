package main

import (
	"fmt"
	"iter"
	"slices"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🚀 MaterializeJSON with Mixed Complex Types")
	fmt.Println("==========================================\n")

	// Create records with multiple complex field types
	tags := slices.Values([]string{"critical", "security"})
	scores := slices.Values([]int{95, 88, 92})
	metadata := streamv3.MakeMutableRecord().
		String("priority", "high").
		Int("version", 2).
		Float("weight", 1.5).
		Freeze()

	task := streamv3.MakeMutableRecord().
		String("id", "COMPLEX-001").
		String("title", "Security Patch").
		StringSeq("tags", tags).
		IntSeq("scores", scores).
		Nested("metadata", metadata).
		Freeze()

	fmt.Println("📊 Original complex record:")
	fmt.Printf("  ID: %s\n", streamv3.GetOr(task, "id", ""))
	fmt.Printf("  Title: %s\n", streamv3.GetOr(task, "title", ""))

	if tagsSeq, ok := streamv3.Get[iter.Seq[string]](task, "tags"); ok {
		fmt.Print("  Tags: ")
		for tag := range tagsSeq {
			fmt.Printf("%s ", tag)
		}
		fmt.Println()
	}

	if scoresSeq, ok := streamv3.Get[iter.Seq[int]](task, "scores"); ok {
		fmt.Print("  Scores: ")
		for score := range scoresSeq {
			fmt.Printf("%d ", score)
		}
		fmt.Println()
	}

	if meta, ok := streamv3.Get[streamv3.Record](task, "metadata"); ok {
		priority := streamv3.GetOr(meta, "priority", "")
		version := streamv3.GetOr(meta, "version", 0)
		weight := streamv3.GetOr(meta, "weight", 0.0)
		fmt.Printf("  Metadata: priority=%s, version=%d, weight=%.1f\n", priority, version, weight)
	}

	stream := streamv3.From([]streamv3.Record{task})

	fmt.Println("\n🔧 Test: MaterializeJSON with different complex field types")
	fmt.Println("-----------------------------------------------------------")

	// Test with string sequence
	fmt.Println("\n1. String sequence materialization:")
	tagsResult := streamv3.Chain(
		streamv3.MaterializeJSON("tags", "tags_json"),
	)(stream)
	for result := range tagsResult {
		fmt.Printf("   tags_json: %s\n", streamv3.GetOr(result, "tags_json", ""))
	}

	// Test with int sequence
	fmt.Println("\n2. Int sequence materialization:")
	scoresResult := streamv3.Chain(
		streamv3.MaterializeJSON("scores", "scores_json"),
	)(streamv3.From([]streamv3.Record{task}))
	for result := range scoresResult {
		fmt.Printf("   scores_json: %s\n", streamv3.GetOr(result, "scores_json", ""))
	}

	// Test with nested Record
	fmt.Println("\n3. Nested Record materialization:")
	metaResult := streamv3.Chain(
		streamv3.MaterializeJSON("metadata", "metadata_json"),
	)(streamv3.From([]streamv3.Record{task}))
	for result := range metaResult {
		fmt.Printf("   metadata_json: %s\n", streamv3.GetOr(result, "metadata_json", ""))
	}

	// Test with simple field (should work too)
	fmt.Println("\n4. Simple field materialization:")
	titleResult := streamv3.Chain(
		streamv3.MaterializeJSON("title", "title_json"),
	)(streamv3.From([]streamv3.Record{task}))
	for result := range titleResult {
		fmt.Printf("   title_json: %s\n", streamv3.GetOr(result, "title_json", ""))
	}

	fmt.Println("\n✅ MaterializeJSON successfully handles:")
	fmt.Println("  📚 iter.Seq[string] → JSON string arrays")
	fmt.Println("  🔢 iter.Seq[int] → JSON number arrays")
	fmt.Println("  🏗️ Record → JSON objects")
	fmt.Println("  📝 Simple values → JSON primitives")
	fmt.Println("  🌟 All types can now be used for consistent grouping!")
}