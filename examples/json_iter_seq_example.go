package main

import (
	"fmt"
	"github.com/rosscartlidge/streamv3"
	"iter"
	"slices"
)

func main() {
	fmt.Println("🧪 Testing WriteJSON with iter.Seq fields")
	fmt.Println("=========================================\n")

	// Create records with iter.Seq fields
	tags1 := slices.Values([]string{"urgent", "work"})
	tags2 := slices.Values([]string{"feature", "enhancement"})

	records := []streamv3.Record{
		streamv3.MakeMutableRecord().
			String("id", "TASK-001").
			String("title", "Fix bug").
			StringSeq("tags", tags1).
			Freeze(),
		streamv3.MakeMutableRecord().
			String("id", "TASK-002").
			String("title", "Add feature").
			StringSeq("tags", tags2).
			Freeze(),
	}

	fmt.Println("📊 Records with iter.Seq fields:")
	for i, record := range records {
		id := streamv3.GetOr(record, "id", "")
		title := streamv3.GetOr(record, "title", "")
		fmt.Printf("  %d. %s: %s\n", i+1, id, title)

		if tagsSeq, ok := streamv3.Get[iter.Seq[string]](record, "tags"); ok {
			fmt.Print("     Tags: ")
			for tag := range tagsSeq {
				fmt.Printf("%s ", tag)
			}
			fmt.Println()
		}
	}

	fmt.Println("\n🔧 Testing WriteJSON with iter.Seq field included:")

	// Create a stream from records
	stream := streamv3.From(records)

	// Try to write JSON including the iter.Seq field
	filename := "/tmp/test_with_iterseq.json"

	fmt.Printf("Writing to: %s\n", filename)

	err := streamv3.WriteJSON(stream, filename)
	if err != nil {
		fmt.Printf("❌ Error writing JSON: %v\n", err)
		return
	}

	fmt.Println("✅ JSON written successfully")

	fmt.Println("\n📖 Raw JSON file:")
	fmt.Println("------------------")
}
