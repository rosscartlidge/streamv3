package main

import (
	"fmt"
	"iter"
	"slices"
	"strings"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🔄 DotFlatten and CrossFlatten Examples")
	fmt.Println("=====================================\n")

	// Test DotFlatten with dot product expansion
	fmt.Println("🔍 DotFlatten Example - Dot Product Expansion")
	fmt.Println("---------------------------------------------")

	tags1 := slices.Values([]string{"urgent", "work"})
	scores1 := slices.Values([]int{85, 92})

	record1 := streamv3.NewRecord().
		String("id", "task-123").
		String("user", "Alice").
		StringSeq("tags", tags1).
		IntSeq("scores", scores1).
		Build()

	fmt.Println("Input record:")
	fmt.Printf("  ID: %s, User: %s\n", streamv3.GetOr(record1, "id", ""), streamv3.GetOr(record1, "user", ""))

	// Show input sequences
	if tagsSeq, ok := streamv3.Get[iter.Seq[string]](record1, "tags"); ok {
		fmt.Print("  Tags: ")
		for tag := range tagsSeq { fmt.Printf("%s ", tag) }
		fmt.Println()
	}

	if scoresSeq, ok := streamv3.Get[iter.Seq[int]](record1, "scores"); ok {
		fmt.Print("  Scores: ")
		for score := range scoresSeq { fmt.Printf("%d ", score) }
		fmt.Println()
	}

	// Apply DotFlatten - pairs corresponding elements (dot product)
	dotResults := streamv3.DotFlatten(".", "tags", "scores")(slices.Values([]streamv3.Record{record1}))

	fmt.Println("\nDotFlatten Results (paired elements):")
	count := 1
	for result := range dotResults {
		id := streamv3.GetOr(result, "id", "")
		user := streamv3.GetOr(result, "user", "")
		tag := streamv3.GetOr(result, "tags", "")
		score := streamv3.GetOr(result, "scores", 0)
		fmt.Printf("  %d. ID: %s, User: %s, Tag: %s, Score: %d\n", count, id, user, tag, score)
		count++
	}

	fmt.Println("\n" + strings.Repeat("=", 50))

	// Test CrossFlatten with cartesian product expansion
	fmt.Println("\n🔍 CrossFlatten Example - Cartesian Product Expansion")
	fmt.Println("--------------------------------------------------")

	priorities := slices.Values([]string{"high", "low"})
	types := slices.Values([]string{"bug", "feature"})

	record2 := streamv3.NewRecord().
		String("project", "StreamV3").
		StringSeq("priorities", priorities).
		StringSeq("types", types).
		Build()

	fmt.Println("Input record:")
	fmt.Printf("  Project: %s\n", streamv3.GetOr(record2, "project", ""))

	if priSeq, ok := streamv3.Get[iter.Seq[string]](record2, "priorities"); ok {
		fmt.Print("  Priorities: ")
		for pri := range priSeq { fmt.Printf("%s ", pri) }
		fmt.Println()
	}

	if typSeq, ok := streamv3.Get[iter.Seq[string]](record2, "types"); ok {
		fmt.Print("  Types: ")
		for typ := range typSeq { fmt.Printf("%s ", typ) }
		fmt.Println()
	}

	// Apply CrossFlatten - all combinations (cartesian product)
	crossResults := streamv3.CrossFlatten(".", "priorities", "types")(slices.Values([]streamv3.Record{record2}))

	fmt.Println("\nCrossFlatten Results (all combinations):")
	count = 1
	for result := range crossResults {
		project := streamv3.GetOr(result, "project", "")
		priority := streamv3.GetOr(result, "priorities", "")
		typ := streamv3.GetOr(result, "types", "")
		fmt.Printf("  %d. Project: %s, Priority: %s, Type: %s\n", count, project, priority, typ)
		count++
	}

	fmt.Println("\n" + strings.Repeat("=", 50))

	// Test nested record flattening
	fmt.Println("\n🔍 Nested Record Flattening Example")
	fmt.Println("-----------------------------------")

	userInfo := streamv3.NewRecord().
		String("name", "Bob").
		String("email", "bob@example.com").
		Build()

	nestedRecord := streamv3.NewRecord().
		String("id", "order-456").
		Record("customer", userInfo).
		Float("amount", 199.99).
		Build()

	fmt.Println("Input nested record:")
	fmt.Printf("  ID: %s, Amount: $%.2f\n", streamv3.GetOr(nestedRecord, "id", ""), streamv3.GetOr(nestedRecord, "amount", 0.0))
	if customer, ok := streamv3.Get[streamv3.Record](nestedRecord, "customer"); ok {
		fmt.Printf("  Customer: %s (%s)\n", streamv3.GetOr(customer, "name", ""), streamv3.GetOr(customer, "email", ""))
	}

	// Apply DotFlatten to flatten nested structure
	nestedResults := streamv3.DotFlatten(".")(slices.Values([]streamv3.Record{nestedRecord}))

	fmt.Println("\nFlattened nested record:")
	for result := range nestedResults {
		for key, value := range result {
			fmt.Printf("  %s: %v\n", key, value)
		}
		break // Only one result expected
	}

	fmt.Println("\n✅ All flattening operations completed!")
	fmt.Println("💡 DotFlatten: Pairs corresponding elements (linear expansion)")
	fmt.Println("💡 CrossFlatten: All combinations (exponential expansion)")
}