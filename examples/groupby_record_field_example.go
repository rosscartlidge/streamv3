package main

import (
	"fmt"
	"github.com/rosscartlidge/streamv3"
	"slices"
)

func main() {
	fmt.Println("🔍 GroupBy with Record Field Test")
	fmt.Println("=================================\n")

	// Create some nested records
	location1 := streamv3.MakeMutableRecord().String("city", "New York").String("country", "USA").Freeze()
	location2 := streamv3.MakeMutableRecord().String("city", "New York").String("country", "USA").Freeze() // Same content, different Record
	location3 := streamv3.MakeMutableRecord().String("city", "London").String("country", "UK").Freeze()    // Different content

	records := []streamv3.Record{
		streamv3.MakeMutableRecord().String("user", "Alice").Nested("location", location1).Freeze(),
		streamv3.MakeMutableRecord().String("user", "Bob").Nested("location", location2).Freeze(),   // Same location content as Alice
		streamv3.MakeMutableRecord().String("user", "Carol").Nested("location", location3).Freeze(), // Different location
	}

	fmt.Println("📊 Sample records:")
	for i, record := range records {
		user := streamv3.GetOr(record, "user", "")
		if loc, ok := streamv3.Get[streamv3.Record](record, "location"); ok {
			city := streamv3.GetOr(loc, "city", "")
			country := streamv3.GetOr(loc, "country", "")
			fmt.Printf("  %d. %s at %s, %s\n", i+1, user, city, country)
		}
	}

	fmt.Println("\n🧪 Trying to group by 'location' field (Record):")

	// Test grouping by Record field
	results := streamv3.Chain(
		streamv3.GroupByFields("group_data", "location"),
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
		if loc, ok := streamv3.Get[streamv3.Record](result, "location"); ok {
			city := streamv3.GetOr(loc, "city", "")
			country := streamv3.GetOr(loc, "country", "")
			fmt.Printf("  Group %d: %d records, location = %s, %s\n", groupCount, count, city, country)
		} else {
			fmt.Printf("  Group %d: %d records\n", groupCount, count)
		}
	}

	fmt.Println("\n💭 Question: Do Records with same content group together?")
	fmt.Println("   Alice and Bob both have {city: 'New York', country: 'USA'}")
	if groupCount == 2 {
		fmt.Println("✅ Success: Records with same content grouped together!")
	} else if groupCount == 3 {
		fmt.Println("⚠️  Issue: Each Record creates a separate group (like iter.Seq)")
	} else {
		fmt.Printf("🤔 Unexpected: Got %d groups\n", groupCount)
	}
}
