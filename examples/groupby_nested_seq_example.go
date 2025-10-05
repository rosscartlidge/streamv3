package main

import (
	"fmt"
	"iter"
	"slices"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🔍 GroupBy with Nested iter.Seq in Record Test")
	fmt.Println("===============================================\n")

	// Create Records with iter.Seq fields inside them
	tags1 := slices.Values([]string{"urgent", "work"})
	tags2 := slices.Values([]string{"urgent", "work"}) // Same content, different sequence
	tags3 := slices.Values([]string{"personal"})

	// Nested records containing iter.Seq fields
	profile1 := streamv3.NewRecord().
		String("role", "developer").
		StringSeq("tags", tags1).
		Build()

	profile2 := streamv3.NewRecord().
		String("role", "developer").
		StringSeq("tags", tags2). // Same content as profile1, but different sequence instance
		Build()

	profile3 := streamv3.NewRecord().
		String("role", "manager").
		StringSeq("tags", tags3).
		Build()

	records := []streamv3.Record{
		streamv3.NewRecord().String("user", "Alice").Record("profile", profile1).Build(),
		streamv3.NewRecord().String("user", "Bob").Record("profile", profile2).Build(),   // Same profile content as Alice
		streamv3.NewRecord().String("user", "Carol").Record("profile", profile3).Build(), // Different profile
	}

	fmt.Println("📊 Sample records:")
	for i, record := range records {
		user := streamv3.GetOr(record, "user", "")
		if profile, ok := streamv3.Get[streamv3.Record](record, "profile"); ok {
			role := streamv3.GetOr(profile, "role", "")
			if tagsSeq, ok := streamv3.Get[iter.Seq[string]](profile, "tags"); ok {
				fmt.Printf("  %d. %s (%s) with tags: ", i+1, user, role)
				for tag := range tagsSeq {
					fmt.Printf("%s ", tag)
				}
				fmt.Println()
			}
		}
	}

	fmt.Println("\n🧪 Trying to group by 'profile' field (Record containing iter.Seq):")

	results := streamv3.Chain(
		streamv3.GroupByFields("group_data", "profile"),
		streamv3.Aggregate("group_data", map[string]streamv3.AggregateFunc{
			"count": streamv3.Count(),
		}),
	)(slices.Values(records))

	fmt.Println("Group results:")
	groupCount := 0
	for result := range results {
		groupCount++
		count := streamv3.GetOr(result, "count", int64(0))

		if profileField, exists := result["profile"]; exists {
			if profile, ok := profileField.(streamv3.Record); ok {
				role := streamv3.GetOr(profile, "role", "")
				fmt.Printf("  Group %d: %d records, role = %s\n", groupCount, count, role)
			}
		}
	}

	fmt.Println("\n💭 Expected: Alice and Bob should group together (same role + same tag content)")
	fmt.Printf("💭 Actual: Got %d groups\n", groupCount)

	if groupCount == 2 {
		fmt.Println("✅ Records with same content (including seq content) grouped together")
	} else if groupCount == 3 {
		fmt.Println("⚠️  Problem: iter.Seq fields in Records prevent proper grouping!")
		fmt.Println("   Alice and Bob have identical content but different sequence instances")
	} else {
		fmt.Printf("🤔 Unexpected result: %d groups\n", groupCount)
	}

	fmt.Println("\n🔍 This demonstrates the deep problem:")
	fmt.Println("   Records are only 'equal' for grouping if ALL nested iter.Seq instances are identical")
	fmt.Println("   Same content in different sequence instances = different groups")
}