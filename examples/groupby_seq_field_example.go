package main

import (
	"fmt"
	"github.com/rosscartlidge/streamv3"
	"iter"
	"slices"
)

func main() {
	fmt.Println("🔍 GroupBy with iter.Seq Field Test")
	fmt.Println("===================================\n")

	// Create records with iter.Seq fields
	tags1 := slices.Values([]string{"urgent", "work"})
	tags2 := slices.Values([]string{"urgent", "work"}) // Same content, different sequence
	tags3 := slices.Values([]string{"personal"})

	records := []streamv3.Record{
		streamv3.MakeMutableRecord().String("user", "Alice").StringSeq("tags", tags1).Freeze(),
		streamv3.MakeMutableRecord().String("user", "Bob").StringSeq("tags", tags2).Freeze(),   // Same content as Alice
		streamv3.MakeMutableRecord().String("user", "Carol").StringSeq("tags", tags3).Freeze(), // Different content
	}

	fmt.Println("📊 Sample records:")
	for i, record := range records {
		user := streamv3.GetOr(record, "user", "")
		if tagsSeq, ok := streamv3.Get[iter.Seq[string]](record, "tags"); ok {
			fmt.Printf("  %d. %s with tags: ", i+1, user)
			for tag := range tagsSeq {
				fmt.Printf("%s ", tag)
			}
			fmt.Println()
		}
	}

	fmt.Println("\n🧪 Trying to group by 'tags' field (iter.Seq[string]):")

	// This will likely produce unexpected results
	results := streamv3.Chain(
		streamv3.GroupByFields("group_data", "tags"),
		streamv3.Aggregate("group_data", map[string]streamv3.AggregateFunc{
			"count": streamv3.Count(),
		}),
	)(slices.Values(records))

	fmt.Println("Group results:")
	groupCount := 0
	for result := range results {
		groupCount++
		count := streamv3.GetOr(result, "count", int64(0))

		// Try to show what the grouping key looks like
		if tagsField, ok := streamv3.Get[iter.Seq[string]](result, "tags"); ok {
			fmt.Printf("  Group %d: %d records, tags field = %T\n", groupCount, count, tagsField)
		}
	}

	fmt.Println("\n⚠️  Problem: iter.Seq fields don't group meaningfully!")
	fmt.Println("💡 Solution: Extract/materialize sequence content for grouping")
	fmt.Println("   or use a different field that represents the sequence content")
}
