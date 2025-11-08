package main

import (
	"fmt"
	"github.com/rosscartlidge/ssql"
	"iter"
	"slices"
)

func main() {
	fmt.Println("🧪 Testing WriteCSV with iter.Seq fields")
	fmt.Println("========================================\n")

	// Create records with iter.Seq fields
	tags1 := slices.Values([]string{"urgent", "work"})
	tags2 := slices.Values([]string{"feature", "enhancement"})

	records := []ssql.Record{
		ssql.MakeMutableRecord().
			String("id", "TASK-001").
			String("title", "Fix bug").
			StringSeq("tags", tags1).
			Freeze(),
		ssql.MakeMutableRecord().
			String("id", "TASK-002").
			String("title", "Add feature").
			StringSeq("tags", tags2).
			Freeze(),
	}

	fmt.Println("📊 Records with iter.Seq fields:")
	for i, record := range records {
		id := ssql.GetOr(record, "id", "")
		title := ssql.GetOr(record, "title", "")
		fmt.Printf("  %d. %s: %s\n", i+1, id, title)

		if tagsSeq, ok := ssql.Get[iter.Seq[string]](record, "tags"); ok {
			fmt.Print("     Tags: ")
			for tag := range tagsSeq {
				fmt.Printf("%s ", tag)
			}
			fmt.Println()
		}
	}

	fmt.Println("\n🔧 Testing WriteCSV with iter.Seq field included:")

	// Create a stream from records
	stream := ssql.From(records)

	// Try to write CSV including the iter.Seq field
	filename := "/tmp/test_with_iterseq.csv"

	fmt.Printf("Writing to: %s\n", filename)

	err := ssql.WriteCSV(stream, filename)
	if err != nil {
		fmt.Printf("❌ Error writing CSV: %v\n", err)
		return
	}

	fmt.Println("✅ CSV written successfully")

	fmt.Println("\n📖 Reading back the CSV to see what happened:")

	// Read the CSV back to see what was written
	csvStream, err := ssql.ReadCSV(filename)
	if err != nil {
		fmt.Printf("❌ Error reading CSV: %v\n", err)
		return
	}

	for record := range csvStream {
		fmt.Printf("Record: %v\n", record)
	}

	fmt.Println("\n💡 Expected issue: iter.Seq field likely shows as function pointer or useless representation")
	fmt.Println("🔧 Solution needed: Enhanced formatValue() to handle iter.Seq fields properly")
}
