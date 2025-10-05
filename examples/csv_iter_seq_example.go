package main

import (
	"fmt"
	"iter"
	"slices"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🧪 Testing WriteCSV with iter.Seq fields")
	fmt.Println("========================================\n")

	// Create records with iter.Seq fields
	tags1 := slices.Values([]string{"urgent", "work"})
	tags2 := slices.Values([]string{"feature", "enhancement"})

	records := []streamv3.Record{
		streamv3.NewRecord().
			String("id", "TASK-001").
			String("title", "Fix bug").
			StringSeq("tags", tags1).
			Build(),
		streamv3.NewRecord().
			String("id", "TASK-002").
			String("title", "Add feature").
			StringSeq("tags", tags2).
			Build(),
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

	fmt.Println("\n🔧 Testing WriteCSV with iter.Seq field included:")

	// Create a stream from records
	stream := streamv3.From(records)

	// Try to write CSV including the iter.Seq field
	fields := []string{"id", "title", "tags"}
	filename := "/tmp/test_with_iterseq.csv"

	fmt.Printf("Writing to: %s\n", filename)
	fmt.Printf("Fields: %v\n", fields)

	err := streamv3.WriteCSV(stream, filename, fields)
	if err != nil {
		fmt.Printf("❌ Error writing CSV: %v\n", err)
		return
	}

	fmt.Println("✅ CSV written successfully")

	fmt.Println("\n📖 Reading back the CSV to see what happened:")

	// Read the CSV back to see what was written
	csvStream := streamv3.ReadCSV(filename)

	for record := range csvStream {
		fmt.Printf("Record: %v\n", record)
	}

	fmt.Println("\n💡 Expected issue: iter.Seq field likely shows as function pointer or useless representation")
	fmt.Println("🔧 Solution needed: Enhanced formatValue() to handle iter.Seq fields properly")
}